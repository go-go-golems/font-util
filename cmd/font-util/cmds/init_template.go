package cmds

import (
	"context"
	"os"

	"github.com/go-go-golems/font-util/pkg/spec"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"gopkg.in/yaml.v3"
)

type InitTemplateCommand struct {
	*cmds.CommandDescription
}

type InitTemplateSettings struct {
	Font   string `glazed:"font"`
	Out    string `glazed:"out"`
	PdfOut string `glazed:"pdf-out"`
}

func NewInitTemplateCommand() (*InitTemplateCommand, error) {
	cmdDesc := cmds.NewCommandDescription(
		"init-template",
		cmds.WithShort("Create a starter YAML template for typography practice"),
		cmds.WithLong(`
Create a starter YAML template for typography copy-practice sheets.
The generated template includes default sections for kerning pairs,
ligatures, and free practice rows.

Examples:
  font-util init-template --font ./font.otf --out practice.yaml --pdf-out practice.pdf
`),
		cmds.WithFlags(
			fields.New(
				"font",
				fields.TypeString,
				fields.WithDefault("./font.otf"),
				fields.WithHelp("Path to the font file"),
			),
			fields.New(
				"out",
				fields.TypeString,
				fields.WithDefault("practice.yaml"),
				fields.WithHelp("YAML output path"),
			),
			fields.New(
				"pdf-out",
				fields.TypeString,
				fields.WithDefault("practice.pdf"),
				fields.WithHelp("Default PDF output path in the template"),
			),
		),
	)

	return &InitTemplateCommand{CommandDescription: cmdDesc}, nil
}

func (c *InitTemplateCommand) RunIntoGlazeProcessor(
	_ context.Context,
	vals *values.Values,
	_ middlewares.Processor,
) error {
	s := &InitTemplateSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}

	tmpl := spec.Starter(s.Font, s.PdfOut)
	b, err := yaml.Marshal(tmpl)
	if err != nil {
		return err
	}

	if err := os.WriteFile(s.Out, b, 0644); err != nil {
		return err
	}

	return nil
}
