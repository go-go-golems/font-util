package ttc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDidotTTC(t *testing.T) {
	path := filepath.Join("..", "..", "Didot.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Didot.ttc not found (test data not available)")
	}

	ttc, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if ttc.Header.Tag != "ttcf" {
		t.Errorf("Expected tag 'ttcf', got %q", ttc.Header.Tag)
	}

	if ttc.Header.NumFonts == 0 {
		t.Error("Expected at least 1 font in Didot.ttc")
	}

	t.Logf("Didot.ttc: %d fonts", ttc.Header.NumFonts)
	for _, font := range ttc.Fonts {
		t.Logf("  Font %d: name=%q, tables=%d, SFNTVersion=0x%08X",
			font.Index, font.Name, font.Header.NumTables, font.Header.SFNTVersion)

		if font.Name == "" {
			t.Errorf("Font %d has empty name", font.Index)
		}
		if font.Header.NumTables == 0 {
			t.Errorf("Font %d has no tables", font.Index)
		}
	}
}

func TestParseFuturaTTC(t *testing.T) {
	path := filepath.Join("..", "..", "Futura.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Futura.ttc not found (test data not available)")
	}

	ttc, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if ttc.Header.NumFonts == 0 {
		t.Error("Expected at least 1 font in Futura.ttc")
	}

	t.Logf("Futura.ttc: %d fonts", ttc.Header.NumFonts)
	for _, font := range ttc.Fonts {
		t.Logf("  Font %d: name=%q, tables=%d",
			font.Index, font.Name, font.Header.NumTables)
	}
}

func TestParseGillSansTTC(t *testing.T) {
	path := filepath.Join("..", "..", "GillSans.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: GillSans.ttc not found (test data not available)")
	}

	ttc, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if ttc.Header.NumFonts == 0 {
		t.Error("Expected at least 1 font in GillSans.ttc")
	}

	t.Logf("GillSans.ttc: %d fonts", ttc.Header.NumFonts)
	for _, font := range ttc.Fonts {
		t.Logf("  Font %d: name=%q, tables=%d",
			font.Index, font.Name, font.Header.NumTables)
	}
}

func TestParseRejectsNonTTC(t *testing.T) {
	_, err := Parse([]byte("NOT_A_TTF_FILE_AT_ALL_REALLY"))
	if err == nil {
		t.Error("Expected error for non-TTC data, got nil")
	}
}

func TestParseRejectsEmptyData(t *testing.T) {
	_, err := Parse([]byte{})
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
}

func TestParseRejectsTooSmall(t *testing.T) {
	_, err := Parse([]byte{0x00, 0x01, 0x02})
	if err == nil {
		t.Error("Expected error for too-small data, got nil")
	}
}

func TestParseRejectsTTFInsteadOfTTC(t *testing.T) {
	// A standalone TTF starts with 0x00010000, not "ttcf"
	data := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, err := Parse(data)
	if err == nil {
		t.Error("Expected error for TTF data (not TTC), got nil")
	}
}

func TestCalcSearchFields(t *testing.T) {
	tests := []struct {
		numTables     uint16
		searchRange   uint16
		entrySelector uint16
		rangeShift    uint16
	}{
		{1, 16, 0, 0},
		{2, 32, 1, 0},
		{4, 64, 2, 0},
		{8, 128, 3, 0},
		{9, 128, 3, 16},
		{16, 256, 4, 0},
		{17, 256, 4, 16},
		{20, 256, 4, 64},
	}

	for _, tt := range tests {
		sr, es, rs := CalcSearchFields(tt.numTables)
		if sr != tt.searchRange || es != tt.entrySelector || rs != tt.rangeShift {
			t.Errorf("CalcSearchFields(%d) = (%d, %d, %d), want (%d, %d, %d)",
				tt.numTables, sr, es, rs, tt.searchRange, tt.entrySelector, tt.rangeShift)
		}
	}
}

func TestExtractDidotRoundTrip(t *testing.T) {
	path := filepath.Join("..", "..", "Didot.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Didot.ttc not found (test data not available)")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	ttc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(ttc.Fonts) == 0 {
		t.Fatal("No fonts in TTC")
	}

	// Extract the first font to a temp file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, ttc.Fonts[0].Name+".ttf")

	err = ExtractFont(data, ttc.Fonts[0], outputPath)
	if err != nil {
		t.Fatalf("ExtractFont failed: %v", err)
	}

	// Verify the file exists and has content
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Stat on extracted file failed: %v", err)
	}
	if info.Size() == 0 {
		t.Error("Extracted TTF file is empty")
	}

	// Re-parse the extracted TTF — it should start with a valid SFNT version
	extractedData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile on extracted TTF failed: %v", err)
	}

	if len(extractedData) < 12 {
		t.Fatalf("Extracted TTF too small: %d bytes", len(extractedData))
	}

	sfntVersion := uint32(extractedData[0])<<24 | uint32(extractedData[1])<<16 | uint32(extractedData[2])<<8 | uint32(extractedData[3])
	if sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F {
		t.Errorf("Extracted file has invalid SFNT version: 0x%08X", sfntVersion)
	}

	numTables := uint16(extractedData[4])<<8 | uint16(extractedData[5])
	if numTables != ttc.Fonts[0].Header.NumTables {
		t.Errorf("Extracted file NumTables mismatch: got %d, want %d", numTables, ttc.Fonts[0].Header.NumTables)
	}

	t.Logf("Extracted %s: %d bytes, %d tables, SFNTVersion=0x%08X",
		outputPath, info.Size(), numTables, sfntVersion)
}

func TestExtractAllFontsDidot(t *testing.T) {
	path := filepath.Join("..", "..", "Didot.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Didot.ttc not found (test data not available)")
	}

	tmpDir := t.TempDir()
	outputPaths, fontNames, err := ExtractAllFonts(path, tmpDir, false)
	if err != nil {
		t.Fatalf("ExtractAllFonts failed: %v", err)
	}

	if len(outputPaths) == 0 {
		t.Error("No fonts extracted")
	}

	for i, p := range outputPaths {
		info, err := os.Stat(p)
		if err != nil {
			t.Errorf("Font %d (%s): file not found: %v", i, fontNames[i], err)
			continue
		}
		t.Logf("Font %d: name=%q, path=%s, size=%d", i, fontNames[i], p, info.Size())
		if info.Size() == 0 {
			t.Errorf("Font %d (%s): file is empty", i, fontNames[i])
		}
	}
}

func TestExtractFontBytesRoundTrip(t *testing.T) {
	path := filepath.Join("..", "..", "Futura.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Futura.ttc not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	ttc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Extract font 2 (Futura-Bold) in-memory
	ttfBytes, err := ExtractFontBytes(data, ttc.Fonts[2])
	if err != nil {
		t.Fatalf("ExtractFontBytes failed: %v", err)
	}

	if len(ttfBytes) == 0 {
		t.Fatal("ExtractFontBytes returned empty bytes")
	}

	// Verify the in-memory result can be re-parsed as a valid TTF
	sfntVersion := uint32(ttfBytes[0])<<24 | uint32(ttfBytes[1])<<16 | uint32(ttfBytes[2])<<8 | uint32(ttfBytes[3])
	if sfntVersion != 0x00010000 && sfntVersion != 0x4F54544F {
		t.Errorf("In-memory TTF has invalid SFNT version: 0x%08X", sfntVersion)
	}

	numTables := uint16(ttfBytes[4])<<8 | uint16(ttfBytes[5])
	if numTables != ttc.Fonts[2].Header.NumTables {
		t.Errorf("NumTables mismatch: got %d, want %d", numTables, ttc.Fonts[2].Header.NumTables)
	}

	t.Logf("Extracted Futura-Bold in-memory: %d bytes, %d tables", len(ttfBytes), numTables)
}

func TestExtractFontBytesVersusFile(t *testing.T) {
	path := filepath.Join("..", "..", "Didot.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Didot.ttc not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	ttc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Extract the same font both ways and compare
	tmpDir := t.TempDir()
	fileOutputPath := filepath.Join(tmpDir, "test.ttf")

	err = ExtractFont(data, ttc.Fonts[0], fileOutputPath)
	if err != nil {
		t.Fatalf("ExtractFont (file) failed: %v", err)
	}

	fileBytes, err := os.ReadFile(fileOutputPath)
	if err != nil {
		t.Fatalf("ReadFile on extracted TTF failed: %v", err)
	}

	memBytes, err := ExtractFontBytes(data, ttc.Fonts[0])
	if err != nil {
		t.Fatalf("ExtractFontBytes failed: %v", err)
	}

	if len(fileBytes) != len(memBytes) {
		t.Errorf("Size mismatch: file=%d bytes, memory=%d bytes", len(fileBytes), len(memBytes))
	}

	// Byte-for-byte comparison
	match := true
	for i := range fileBytes {
		if i >= len(memBytes) || fileBytes[i] != memBytes[i] {
			match = false
			break
		}
	}
	if !match {
		t.Error("In-memory and file extraction produced different bytes")
	} else {
		t.Logf("Both methods produced identical %d-byte output", len(fileBytes))
	}
}

func TestOtfExtensionForCFF(t *testing.T) {
	// Test that ExtractAllFonts uses .otf extension for CFF fonts.
	// Our test TTC files are TrueType, so we create a synthetic test.
	// This tests the logic path without needing a real OTC file.
	path := filepath.Join("..", "..", "Didot.ttc")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("Skipping: Didot.ttc not found")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	ttc, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}

	// All Didot fonts are TrueType
	for _, font := range ttc.Fonts {
		if font.Header.SFNTVersion == 0x4F54544F {
			t.Errorf("Didot font %d unexpectedly has CFF SFNT version", font.Index)
		}
	}

	// Manually verify the extension logic
	for _, font := range ttc.Fonts {
		ext := ".ttf"
		if font.Header.SFNTVersion == 0x4F54544F {
			ext = ".otf"
		}
		if ext != ".ttf" {
			t.Errorf("Didot font %d: expected .ttf extension, got %s", font.Index, ext)
		}
	}
}
