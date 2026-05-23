package cmds

import (
	"context"
	"fmt"

	"github.com/go-go-golems/font-util/pkg/shape"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type InspectFontCommand struct {
	*cmds.CommandDescription
}

type InspectFontSettings struct {
	Font      string  `glazed:"font"`
	Texts     string  `glazed:"text"`
	PointSize float64 `glazed:"point-size"`
}

func NewInspectFontCommand() (*InspectFontCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"inspect-font",
		cmds.WithShort("Inspect font metrics and shaping widths"),
		cmds.WithLong(`
Load an OpenType/TrueType font and print its metrics (ascender, descender,
x-height, cap height, etc.) and optionally shape text examples to see
glyph counts, advance widths, and missing glyphs.

Supports both individual font files (.otf, .ttf) and TrueType Collections (.ttc).
For TTC files, the first font in the collection is inspected.

Examples:
  font-util inspect-font ./font.otf
  font-util inspect-font ./font.otf --text "AV,To,fi,office" --output json
  font-util inspect-font fonts.ttc --text "fi"
`),
		cmds.WithFlags(
			fields.New(
				"font",
				fields.TypeString,
				fields.WithHelp("Path to the font file (.otf, .ttf, or .ttc)"),
				fields.WithIsArgument(true),
			),
			fields.New(
				"text",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Comma-separated text examples to shape"),
			),
			fields.New(
				"point-size",
				fields.TypeFloat,
				fields.WithDefault(54.0),
				fields.WithHelp("Point size for shaping examples"),
			),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)

	return &InspectFontCommand{CommandDescription: cmdDesc}, nil
}

func (c *InspectFontCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &InspectFontSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	if s.Font == "" {
		return fmt.Errorf("--font is required")
	}

	loaded, err := loadFont(s.Font)
	if err != nil {
		return err
	}

	// Emit metrics as a row
	row := types.NewRow(
		types.MRP("font_name", loaded.Metrics.FontName),
		types.MRP("units_per_em", loaded.Metrics.UnitsPerEm),
		types.MRP("glyph_count", loaded.Metrics.GlyphCount),
		types.MRP("ascender", loaded.Metrics.Ascender),
		types.MRP("descender", loaded.Metrics.Descender),
		types.MRP("line_gap", loaded.Metrics.LineGap),
		types.MRP("x_height", loaded.Metrics.XHeight),
		types.MRP("cap_height", loaded.Metrics.CapHeight),
		types.MRP("source", loaded.Metrics.Source),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}

	// Shape text examples if provided
	if s.Texts != "" {
		sh := shape.NewWithBytes(loaded.Bytes, loaded.Font, loaded.Metrics)
		for _, t := range splitCSV(s.Texts) {
			r, err := sh.Shape(t, shape.Options{PointSize: s.PointSize, Kern: true, Liga: true})
			if err != nil {
				return err
			}
			shapingRow := types.NewRow(
				types.MRP("font_name", loaded.Metrics.FontName),
				types.MRP("text", r.Text),
				types.MRP("glyphs", len(r.Glyphs)),
				types.MRP("advance_pt", fmt.Sprintf("%.2f", r.AdvancePt)),
				types.MRP("missing_glyphs", r.MissingGlyphs),
				types.MRP("engine", r.Engine),
			)
			if err := gp.AddRow(ctx, shapingRow); err != nil {
				return err
			}
		}
	}

	return nil
}
