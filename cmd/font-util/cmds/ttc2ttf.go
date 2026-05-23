package cmds

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type Ttc2TtfCommand struct {
	*cmds.CommandDescription
}

type Ttc2TtfSettings struct {
	InputFile string `glazed:"input-file"`
	OutputDir string `glazed:"output-dir"`
	Force     bool   `glazed:"force"`
}

func NewTtc2TtfCommand() (*Ttc2TtfCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"ttc2ttf",
		cmds.WithShort("Extract individual TTF files from a TTC collection"),
		cmds.WithLong(`
Extract each font embedded in a TrueType Collection (.ttc) file
into a standalone TrueType Font (.ttf) file.

The output files are named using the PostScript name (Name ID 6)
from each font's 'name' table. If the name cannot be extracted,
the fallback name is "font-{index}.ttf".

Examples:
  font-util ttc2ttf fonts.ttc
  font-util ttc2ttf fonts.ttc --output-dir ./extracted
  font-util ttc2ttf fonts.ttc --force
`),
		cmds.WithFlags(
			fields.New(
				"output-dir",
				fields.TypeString,
				fields.WithDefault("."),
				fields.WithHelp("Directory to write extracted TTF files to"),
			),
			fields.New(
				"force",
				fields.TypeBool,
				fields.WithDefault(false),
				fields.WithHelp("Overwrite existing output files"),
			),
		),
		cmds.WithArguments(
			fields.New(
				"input-file",
				fields.TypeString,
				fields.WithHelp("Path to the .ttc file to extract"),
				fields.WithIsArgument(true),
			),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)

	return &Ttc2TtfCommand{CommandDescription: cmdDesc}, nil
}

func (c *Ttc2TtfCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &Ttc2TtfSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	// Stub: will be replaced with actual extraction logic
	row := types.NewRow(
		types.MRP("status", "not yet implemented"),
		types.MRP("input_file", s.InputFile),
		types.MRP("output_dir", s.OutputDir),
	)
	return gp.AddRow(ctx, row)
}
