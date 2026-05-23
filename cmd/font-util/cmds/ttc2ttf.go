package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/font-util/pkg/ttc"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
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
	)

	return &Ttc2TtfCommand{CommandDescription: cmdDesc}, nil
}

func (c *Ttc2TtfCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	_ middlewares.Processor,
) error {
	s := &Ttc2TtfSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	if s.InputFile == "" {
		return fmt.Errorf("input-file is required")
	}

	if _, err := os.Stat(s.InputFile); err != nil {
		return fmt.Errorf("input file not found: %s", s.InputFile)
	}

	outputPaths, fontNames, err := ttc.ExtractAllFonts(s.InputFile, s.OutputDir, s.Force)
	if err != nil {
		return err
	}

	for i, outputPath := range outputPaths {
		fmt.Printf("  %s -> %s\n", fontNames[i], outputPath)
	}
	fmt.Printf("\nExtracted %d font(s)\n", len(outputPaths))

	return nil
}
