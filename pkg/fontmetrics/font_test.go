package fontmetrics

import (
	"os"
	"path/filepath"
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

func TestLoadFromTTC(t *testing.T) {
	ttcPath := filepath.Join("..", "..", "GillSans.ttc")
	if _, err := os.Stat(ttcPath); os.IsNotExist(err) {
		t.Skip("Skipping: GillSans.ttc not found")
	}

	data, err := os.ReadFile(ttcPath)
	if err != nil {
		t.Fatal(err)
	}

	// Load font 0 via LoadFromTTC
	loaded0, err := LoadFromTTC(data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if loaded0.Metrics.FontName == "" {
		t.Error("expected non-empty font name for font 0")
	}
	if loaded0.Metrics.Source != "os2" {
		t.Errorf("expected os2 source, got %q", loaded0.Metrics.Source)
	}

	// Load font 7 (Light)
	loaded7, err := LoadFromTTC(data, 7)
	if err != nil {
		t.Fatal(err)
	}
	if loaded7.Metrics.FontName == "" {
		t.Error("expected non-empty font name for font 7")
	}

	// Out of range
	_, err = LoadFromTTC(data, 99)
	if err == nil {
		t.Error("expected error for out-of-range font index")
	}

	t.Logf("Font 0: %s (source=%s, xHeight=%d, capHeight=%d)",
		loaded0.Metrics.FontName, loaded0.Metrics.Source, loaded0.Metrics.XHeight, loaded0.Metrics.CapHeight)
	t.Logf("Font 7: %s (source=%s)", loaded7.Metrics.FontName, loaded7.Metrics.Source)
}
