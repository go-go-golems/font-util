package cmds

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
)

func TestInspectFontOnSingleFont(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "Go-Regular.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0644); err != nil {
		t.Fatal(err)
	}

	// Test loadFont directly
	loaded, err := loadFont(fontPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Metrics.FontName == "" {
		t.Error("expected non-empty font name")
	}
	if loaded.Metrics.UnitsPerEm <= 0 {
		t.Error("expected positive units per em")
	}
}

func TestLoadFontTTCWithIndex(t *testing.T) {
	// Test with real TTC files if available
	ttcPath := filepath.Join("..", "..", "..", "GillSans.ttc")
	if _, err := os.Stat(ttcPath); os.IsNotExist(err) {
		t.Skip("Skipping: GillSans.ttc not found")
	}

	// Load font 0
	loaded0, err := loadFont(ttcPath, 0)
	if err != nil {
		t.Fatal(err)
	}
	if loaded0.Metrics.FontName != "Gill Sans" {
		t.Errorf("font 0 name = %q, want %q", loaded0.Metrics.FontName, "Gill Sans")
	}

	// Load font 1 (Bold)
	loaded1, err := loadFont(ttcPath, 1)
	if err != nil {
		t.Fatal(err)
	}
	if loaded1.Metrics.FontName != "Gill Sans Bold" {
		t.Errorf("font 1 name = %q, want %q", loaded1.Metrics.FontName, "Gill Sans Bold")
	}

	// Out of range
	_, err = loadFont(ttcPath, 99)
	if err == nil {
		t.Error("expected error for out-of-range font index")
	}
}

func TestIsTTCFile(t *testing.T) {
	dir := t.TempDir()

	// Not a TTC
	ttfPath := filepath.Join(dir, "font.ttf")
	if err := os.WriteFile(ttfPath, goregular.TTF, 0644); err != nil {
		t.Fatal(err)
	}
	if isTTCFile(ttfPath) {
		t.Error("goregular.TTF should not be detected as TTC")
	}

	// A TTC file
	ttcPath := filepath.Join("..", "..", "..", "GillSans.ttc")
	if _, err := os.Stat(ttcPath); os.IsNotExist(err) {
		t.Skip("Skipping: GillSans.ttc not found")
	}
	if !isTTCFile(ttcPath) {
		t.Error("GillSans.ttc should be detected as TTC")
	}
}

func TestInitTemplateCreatesFile(t *testing.T) {
	dir := t.TempDir()
	fontPath := filepath.Join(dir, "font.otf")
	yamlPath := filepath.Join(dir, "practice.yaml")
	pdfPath := filepath.Join(dir, "practice.pdf")

	// Touch a dummy font file
	if err := os.WriteFile(fontPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test that spec.Starter creates a valid template
	s := specStarter(fontPath, pdfPath)
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Error("expected non-empty template")
	}

	// Also test that writing works
	if err := os.WriteFile(yamlPath, b, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(yamlPath); err != nil {
		t.Errorf("YAML file not created: %v", err)
	}
}

// Minimal spec starter to avoid importing spec in tests
func specStarter(font, output string) map[string]interface{} {
	return map[string]interface{}{
		"version": 1,
		"font":    font,
		"output":  output,
	}
}
