package fontmetrics

import (
	"testing"

	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/sfnt"
)

func TestExtractGoRegular(t *testing.T) {
	f, err := sfnt.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	m, err := Extract(goregular.TTF, f)
	if err != nil {
		t.Fatal(err)
	}
	if m.UnitsPerEm <= 0 || m.GlyphCount <= 0 {
		t.Fatalf("bad metrics: %+v", m)
	}
	if m.XHeight <= 0 || m.CapHeight <= 0 {
		t.Fatalf("missing guide metrics: %+v", m)
	}
}
