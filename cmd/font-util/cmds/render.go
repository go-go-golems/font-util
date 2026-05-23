package cmds

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-go-golems/font-util/pkg/layout"
	"github.com/go-go-golems/font-util/pkg/renderpdf"
	"github.com/go-go-golems/font-util/pkg/shape"
	"github.com/go-go-golems/font-util/pkg/spec"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
)

type RenderCommand struct {
	*cmds.CommandDescription
}

type RenderSettings struct {
	YamlTemplate string  `glazed:"yaml-template"`
	Font         string  `glazed:"font"`
	FontIndex    int     `glazed:"font-index"`
	Out          string  `glazed:"out"`
	Text         string  `glazed:"text"`
	Glyphs       string  `glazed:"glyphs"`
	BlankLines   int     `glazed:"blank-lines"`
	DryRun       bool    `glazed:"dry-run"`
	DebugShaping bool    `glazed:"debug-shaping"`
	PointSize    float64 `glazed:"point-size"`
}

func NewRenderCommand() (*RenderCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"render",
		cmds.WithShort("Render a typography practice PDF from a font and template"),
		cmds.WithLong(`
Render a typography copy-practice PDF. You can either use a YAML template
(created with init-template) or use quick-mode flags for simple sheets.

The PDF includes model rows (with the font drawn as vector outlines) and
blank practice rows with baseline, x-height, and cap-height helper lines.

Examples:
  font-util render --yaml-template practice.yaml
  font-util render --font ./font.otf --text "A,V,AV,To,fi,office" --blank-lines 3 --out practice.pdf
  font-util render --yaml-template practice.yaml --dry-run
  font-util render --font ./font.otf --text "AV" --debug-shaping
  font-util render --font fonts.ttc --font-index 1 --text "AV"
`),
		cmds.WithFlags(
			fields.New(
				"yaml-template",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("YAML template path (created by init-template)"),
			),
			fields.New(
				"font",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Font file path (overrides template)"),
			),
			fields.New(
				"font-index",
				fields.TypeInteger,
				fields.WithDefault(0),
				fields.WithHelp("Index of the font to use within a TTC file (0-based)"),
			),
			fields.New(
				"out",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("PDF output path (overrides template)"),
			),
			fields.New(
				"text",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Comma-separated text items for quick-mode rendering"),
			),
			fields.New(
				"glyphs",
				fields.TypeString,
				fields.WithDefault(""),
				fields.WithHelp("Comma-separated glyph items for quick-mode rendering"),
			),
			fields.New(
				"blank-lines",
				fields.TypeInteger,
				fields.WithDefault(-1),
				fields.WithHelp("Number of blank practice lines after model row (-1 = use template default)"),
			),
			fields.New(
				"dry-run",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Print layout JSON instead of creating a PDF"),
			),
			fields.New(
				"debug-shaping",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Print shaping runs to stderr"),
			),
			fields.New(
				"point-size",
				fields.TypeFloat,
				fields.WithDefault(54.0),
				fields.WithHelp("Point size for quick-mode rendering (default: 54)"),
			),
		),
	)

	return &RenderCommand{CommandDescription: cmdDesc}, nil
}

func (c *RenderCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	_ middlewares.Processor,
) error {
	s := &RenderSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	var sheetSpec spec.SheetSpec
	var err error

	if s.YamlTemplate != "" {
		sheetSpec, err = spec.Load(s.YamlTemplate)
		if err != nil {
			return err
		}
	} else {
		items := splitCSV(firstNonEmpty(s.Text, s.Glyphs))
		if len(items) == 0 && s.BlankLines < 0 {
			return fmt.Errorf("quick render requires --text, --glyphs, or --blank-lines")
		}
		blankLines := s.BlankLines
		blankLinesPtr := &blankLines
		if blankLines < 0 {
			blankLinesPtr = nil
		}
		row := spec.RowSpec{Items: items, BlankLines: blankLinesPtr}
		sheetSpec = spec.SheetSpec{
			Version:  1,
			Font:     s.Font,
			Output:   firstNonEmpty(s.Out, "practice.pdf"),
			Sections: []spec.SectionSpec{{Title: "Practice", Rows: []spec.RowSpec{row}}},
		}
		spec.ApplyDefaults(&sheetSpec)
		sheetSpec.Layout.Wrap = true
	}

	if s.Font != "" {
		sheetSpec.Font = s.Font
	}
	if s.Out != "" {
		sheetSpec.Output = s.Out
	}
	if s.BlankLines >= 0 {
		for si := range sheetSpec.Sections {
			for ri := range sheetSpec.Sections[si].Rows {
				sheetSpec.Sections[si].Rows[ri].BlankLines = &s.BlankLines
			}
		}
	}

	rs, err := spec.Resolve(sheetSpec)
	if err != nil {
		return err
	}

	loaded, err := loadFont(rs.Font, s.FontIndex)
	if err != nil {
		return err
	}

	sh := shape.NewWithBytes(loaded.Bytes, loaded.Font, loaded.Metrics)
	doc, err := layout.Build(rs, loaded.Metrics, sh)
	if err != nil {
		return err
	}

	if s.DebugShaping {
		for _, p := range doc.Pages {
			for _, row := range p.Rows {
				for _, it := range row.Items {
					fmt.Fprintf(os.Stderr, "%s: glyphs=%d advance=%.2f note=%s\n", it.Text, len(it.Run.Glyphs), it.Run.AdvancePt, it.Run.Note)
				}
			}
		}
	}

	if s.DryRun {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(doc)
	}

	err = renderpdf.Render(doc, rs, loaded.Metrics, loaded.Font, rs.Output)
	if err != nil {
		return err
	}

	fmt.Printf("Created %s (%d page(s), font: %s)\n", rs.Output, len(doc.Pages), loaded.Metrics.FontName)
	return nil
}
