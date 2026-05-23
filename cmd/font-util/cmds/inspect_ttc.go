package cmds

import (
	"context"
	"fmt"

	"github.com/go-go-golems/font-util/pkg/ttc"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
)

type InspectTtcCommand struct {
	*cmds.CommandDescription
}

type InspectTtcSettings struct {
	File string `glazed:"file"`
}

// Verify interface compliance
var _ cmds.GlazeCommand = (*InspectTtcCommand)(nil)

func NewInspectTtcCommand() (*InspectTtcCommand, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}

	commandSettingsSection, err := cli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}

	cmdDesc := cmds.NewCommandDescription(
		"inspect-ttc",
		cmds.WithShort("List all fonts in a TrueType Collection"),
		cmds.WithLong(`
Parse a .ttc file and list all member fonts with their index, name,
SFNT version, and table count.

Examples:
  font-util inspect-ttc fonts.ttc
  font-util inspect-ttc fonts.ttc --output json
  font-util inspect-ttc fonts.ttc --fields index,name
`),
		cmds.WithArguments(
			fields.New(
				"file",
				fields.TypeString,
				fields.WithHelp("Path to the .ttc file"),
				fields.WithIsArgument(true),
			),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)

	return &InspectTtcCommand{CommandDescription: cmdDesc}, nil
}

func (c *InspectTtcCommand) RunIntoGlazeProcessor(
	ctx context.Context,
	vals *values.Values,
	gp middlewares.Processor,
) error {
	s := &InspectTtcSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	if s.File == "" {
		return fmt.Errorf("file is required")
	}

	ttcFile, err := ttc.ParseFile(s.File)
	if err != nil {
		return err
	}

	for _, font := range ttcFile.Fonts {
		sfntLabel := "TrueType"
		ext := ".ttf"
		if font.Header.SFNTVersion == 0x4F54544F {
			sfntLabel = "CFF/OpenType"
			ext = ".otf"
		}

		row := types.NewRow(
			types.MRP("index", font.Index),
			types.MRP("name", font.Name),
			types.MRP("sfnt_version", fmt.Sprintf("0x%08X", font.Header.SFNTVersion)),
			types.MRP("sfnt_type", sfntLabel),
			types.MRP("tables", font.Header.NumTables),
			types.MRP("output_ext", ext),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}

	return nil
}
