package layout

import (
	"testing"

	"github.com/go-go-golems/font-util/pkg/fontmetrics"
	"github.com/go-go-golems/font-util/pkg/shape"
	"github.com/go-go-golems/font-util/pkg/spec"
	"golang.org/x/image/font/gofont/goregular"
)

func TestRowWrapAlternatesModelAndBlankRows(t *testing.T) {
	loaded, err := fontmetrics.LoadBytes(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	s := spec.Starter("font.ttf", "out.pdf")
	s.Layout.Wrap = true
	s.Layout.ItemGap = "24pt"
	s.Sections = []spec.SectionSpec{{Title: "wrap", Rows: []spec.RowSpec{{Items: []string{"A", "V", "AV", "To", "fi", "office", "O", "B", "a", "e", "g", "8"}, BlankLines: intPtr(1)}}}}
	r, err := spec.Resolve(s)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(r, loaded.Metrics, shape.NewWithBytes(loaded.Bytes, loaded.Font, loaded.Metrics))
	if err != nil {
		t.Fatal(err)
	}
	rows := doc.Pages[0].Rows
	if len(rows) != 4 {
		t.Fatalf("rows=%d want 4", len(rows))
	}
	if !rows[0].Model || rows[1].Model || !rows[2].Model || rows[3].Model {
		t.Fatalf("expected model, blank, model, blank; got %+v", []bool{rows[0].Model, rows[1].Model, rows[2].Model, rows[3].Model})
	}
}

func TestCellsChunkIntoModelAndBlankRows(t *testing.T) {
	loaded, err := fontmetrics.LoadBytes(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	s := spec.Starter("font.ttf", "out.pdf")
	s.Layout.Mode = "cells"
	s.Layout.Columns = 2
	s.Sections = []spec.SectionSpec{{Title: "cells", Rows: []spec.RowSpec{{Items: []string{"AV", "To", "fi"}, BlankLines: intPtr(1)}}}}
	r, err := spec.Resolve(s)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(r, loaded.Metrics, shape.NewWithBytes(loaded.Bytes, loaded.Font, loaded.Metrics))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Pages) != 1 {
		t.Fatalf("pages=%d", len(doc.Pages))
	}
	rows := doc.Pages[0].Rows
	if len(rows) != 4 {
		t.Fatalf("rows=%d want 4", len(rows))
	}
	if got := len(rows[0].Items); got != 2 {
		t.Fatalf("first model items=%d", got)
	}
	if got := len(rows[2].Items); got != 1 {
		t.Fatalf("second model items=%d", got)
	}
}

func intPtr(v int) *int { return &v }
