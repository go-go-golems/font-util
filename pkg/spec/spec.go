package spec

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type SheetSpec struct {
	Version  int           `yaml:"version" json:"version"`
	Font     string        `yaml:"font" json:"font"`
	Output   string        `yaml:"output" json:"output"`
	Page     PageSpec      `yaml:"page" json:"page"`
	Style    StyleSpec     `yaml:"style" json:"style"`
	Layout   LayoutSpec    `yaml:"layout" json:"layout"`
	Shaping  ShapingSpec   `yaml:"shaping" json:"shaping"`
	Sections []SectionSpec `yaml:"sections" json:"sections"`
}

type PageSpec struct {
	Size        string     `yaml:"size" json:"size"`
	Orientation string     `yaml:"orientation" json:"orientation"`
	Margin      MarginSpec `yaml:"margin" json:"margin"`
}

type MarginSpec struct {
	Top    string `yaml:"top" json:"top"`
	Right  string `yaml:"right" json:"right"`
	Bottom string `yaml:"bottom" json:"bottom"`
	Left   string `yaml:"left" json:"left"`
}

type StyleSpec struct {
	PointSize    string                    `yaml:"point_size" json:"point_size"`
	ModelOpacity float64                   `yaml:"model_opacity" json:"model_opacity"`
	HelperLines  map[string]HelperLineSpec `yaml:"helper_lines" json:"helper_lines"`
	Labels       LabelSpec                 `yaml:"labels" json:"labels"`
}

type HelperLineSpec struct {
	Show  bool     `yaml:"show" json:"show"`
	Width string   `yaml:"width" json:"width"`
	Color string   `yaml:"color" json:"color"`
	Dash  []string `yaml:"dash" json:"dash"`
}

type LabelSpec struct {
	Show     bool   `yaml:"show" json:"show"`
	FontSize string `yaml:"font_size" json:"font_size"`
}

type LayoutSpec struct {
	Mode                    string `yaml:"mode" json:"mode"`
	Columns                 int    `yaml:"columns" json:"columns"`
	RowGap                  string `yaml:"row_gap" json:"row_gap"`
	CellGap                 string `yaml:"cell_gap" json:"cell_gap"`
	ItemGap                 string `yaml:"item_gap" json:"item_gap"`
	PracticeLinesAfterModel int    `yaml:"practice_lines_after_model" json:"practice_lines_after_model"`
	Wrap                    bool   `yaml:"wrap" json:"wrap"`
}

type ShapingSpec struct {
	Direction string          `yaml:"direction" json:"direction"`
	Language  string          `yaml:"language" json:"language"`
	Script    string          `yaml:"script" json:"script"`
	Features  map[string]bool `yaml:"features" json:"features"`
}

type SectionSpec struct {
	Title string    `yaml:"title" json:"title"`
	Rows  []RowSpec `yaml:"rows" json:"rows"`
}

type RowSpec struct {
	Items      []string `yaml:"items" json:"items"`
	BlankLines *int     `yaml:"blank_lines" json:"blank_lines,omitempty"`
	Repeat     int      `yaml:"repeat" json:"repeat"`
	PointSize  string   `yaml:"point_size" json:"point_size,omitempty"`
}

type Resolved struct {
	SheetSpec
	PointSize float64
	Margins   MarginsPt
	RowGap    float64
	CellGap   float64
	ItemGap   float64
	LabelSize float64
}

type MarginsPt struct{ Top, Right, Bottom, Left float64 }

func Load(path string) (SheetSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return SheetSpec{}, err
	}
	var s SheetSpec
	if err := yaml.Unmarshal(b, &s); err != nil {
		return SheetSpec{}, err
	}
	ApplyDefaults(&s)
	return s, nil
}

func ApplyDefaults(s *SheetSpec) {
	if s.Version == 0 {
		s.Version = 1
	}
	if s.Page.Size == "" {
		s.Page.Size = "A4"
	}
	if s.Page.Orientation == "" {
		s.Page.Orientation = "portrait"
	}
	if s.Page.Margin.Top == "" {
		s.Page.Margin.Top = "36pt"
	}
	if s.Page.Margin.Right == "" {
		s.Page.Margin.Right = "36pt"
	}
	if s.Page.Margin.Bottom == "" {
		s.Page.Margin.Bottom = "36pt"
	}
	if s.Page.Margin.Left == "" {
		s.Page.Margin.Left = "36pt"
	}
	if s.Style.PointSize == "" {
		s.Style.PointSize = "54pt"
	}
	if s.Style.ModelOpacity == 0 {
		s.Style.ModelOpacity = 1
	}
	if s.Style.HelperLines == nil {
		s.Style.HelperLines = defaultHelperLines()
	}
	for k, v := range defaultHelperLines() {
		if _, ok := s.Style.HelperLines[k]; !ok {
			s.Style.HelperLines[k] = v
		}
	}
	if s.Style.Labels.FontSize == "" {
		s.Style.Labels.FontSize = "7pt"
	}
	if s.Layout.Mode == "" {
		s.Layout.Mode = "row"
	}
	if s.Layout.Columns == 0 {
		s.Layout.Columns = 1
	}
	if s.Layout.RowGap == "" {
		s.Layout.RowGap = "18pt"
	}
	if s.Layout.CellGap == "" {
		s.Layout.CellGap = "12pt"
	}
	if s.Layout.ItemGap == "" {
		s.Layout.ItemGap = "24pt"
	}
	if s.Layout.PracticeLinesAfterModel == 0 {
		s.Layout.PracticeLinesAfterModel = 2
	}
	if s.Shaping.Direction == "" {
		s.Shaping.Direction = "ltr"
	}
	if s.Shaping.Language == "" {
		s.Shaping.Language = "en"
	}
	if s.Shaping.Script == "" {
		s.Shaping.Script = "latn"
	}
	if s.Shaping.Features == nil {
		s.Shaping.Features = map[string]bool{"kern": true, "liga": true, "clig": true, "calt": true}
	}
}

func defaultHelperLines() map[string]HelperLineSpec {
	return map[string]HelperLineSpec{
		"baseline":   {Show: true, Width: "0.8pt", Color: "#222222"},
		"x_height":   {Show: true, Width: "0.4pt", Color: "#6798d0", Dash: []string{"3pt", "3pt"}},
		"cap_height": {Show: true, Width: "0.4pt", Color: "#d06767", Dash: []string{"6pt", "3pt"}},
		"ascender":   {Show: false, Width: "0.3pt", Color: "#aaaaaa", Dash: []string{"2pt", "2pt"}},
		"descender":  {Show: false, Width: "0.3pt", Color: "#aaaaaa", Dash: []string{"2pt", "2pt"}},
	}
}

func Resolve(s SheetSpec) (Resolved, error) {
	ApplyDefaults(&s)
	if err := Validate(s); err != nil {
		return Resolved{}, err
	}
	ps, err := ParseLengthPt(s.Style.PointSize)
	if err != nil {
		return Resolved{}, fmt.Errorf("style.point_size: %w", err)
	}
	mt, err := ParseLengthPt(s.Page.Margin.Top)
	if err != nil {
		return Resolved{}, err
	}
	mr, err := ParseLengthPt(s.Page.Margin.Right)
	if err != nil {
		return Resolved{}, err
	}
	mb, err := ParseLengthPt(s.Page.Margin.Bottom)
	if err != nil {
		return Resolved{}, err
	}
	ml, err := ParseLengthPt(s.Page.Margin.Left)
	if err != nil {
		return Resolved{}, err
	}
	rg, err := ParseLengthPt(s.Layout.RowGap)
	if err != nil {
		return Resolved{}, err
	}
	cg, err := ParseLengthPt(s.Layout.CellGap)
	if err != nil {
		return Resolved{}, err
	}
	ig, err := ParseLengthPt(s.Layout.ItemGap)
	if err != nil {
		return Resolved{}, err
	}
	ls, err := ParseLengthPt(s.Style.Labels.FontSize)
	if err != nil {
		return Resolved{}, err
	}
	return Resolved{SheetSpec: s, PointSize: ps, Margins: MarginsPt{mt, mr, mb, ml}, RowGap: rg, CellGap: cg, ItemGap: ig, LabelSize: ls}, nil
}

func Validate(s SheetSpec) error {
	if s.Version != 1 {
		return fmt.Errorf("unsupported version %d", s.Version)
	}
	if strings.TrimSpace(s.Font) == "" {
		return errors.New("font is required")
	}
	if strings.TrimSpace(s.Output) == "" {
		return errors.New("output is required")
	}
	if s.Layout.Mode != "row" && s.Layout.Mode != "cells" {
		return fmt.Errorf("layout.mode must be row or cells")
	}
	if s.Layout.Columns < 1 {
		return errors.New("layout.columns must be >= 1")
	}
	if len(s.Sections) == 0 {
		return errors.New("at least one section is required")
	}
	for i, sec := range s.Sections {
		if len(sec.Rows) == 0 {
			return fmt.Errorf("sections[%d] has no rows", i)
		}
		for j, row := range sec.Rows {
			blank := s.Layout.PracticeLinesAfterModel
			if row.BlankLines != nil {
				blank = *row.BlankLines
			}
			if len(row.Items) == 0 && blank < 1 {
				return fmt.Errorf("sections[%d].rows[%d] emits no items and no blank lines", i, j)
			}
		}
	}
	return nil
}

func ParseLengthPt(v string) (float64, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0, errors.New("empty length")
	}
	units := []struct {
		suffix string
		mul    float64
	}{{"pt", 1}, {"in", 72}, {"mm", 72 / 25.4}, {"cm", 72 / 2.54}}
	for _, u := range units {
		if strings.HasSuffix(v, u.suffix) {
			n, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimSuffix(v, u.suffix)), 64)
			if err != nil {
				return 0, err
			}
			return n * u.mul, nil
		}
	}
	return strconv.ParseFloat(v, 64)
}

func Starter(font, output string) SheetSpec {
	if output == "" {
		output = "practice.pdf"
	}
	s := SheetSpec{Version: 1, Font: font, Output: output, Sections: []SectionSpec{{Title: "Kerning and ligatures", Rows: []RowSpec{{Items: []string{"A", "V", "AV", "To", "fi", "office"}}}}, {Title: "Free practice", Rows: []RowSpec{{Items: nil, BlankLines: intPtr(4)}}}}}
	ApplyDefaults(&s)
	// Quick-start sheets should never silently draw off the right edge.
	// YAML users can still edit the generated template if stricter overflow
	// behavior is added later.
	s.Layout.Wrap = true
	return s
}

func intPtr(v int) *int { return &v }
