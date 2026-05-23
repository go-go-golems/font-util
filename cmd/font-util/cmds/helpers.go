package cmds

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/font-util/pkg/fontmetrics"
	"github.com/go-go-golems/font-util/pkg/ttc"
)

// isTTCFile checks if a file has the TTC magic bytes.
func isTTCFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil || len(data) < 4 {
		return false
	}
	return string(data[0:4]) == "ttcf"
}

// loadFont loads a font for rendering/inspection.
// Handles TTC files by extracting the specified font in-memory (no temp files).
// fontIndex selects which font in a TTC to use (0-based); -1 means use the first.
func loadFont(path string, fontIndex int) (*fontmetrics.Loaded, error) {
	if isTTCFile(path) {
		ttcFile, err := ttc.ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("parsing TTC: %w", err)
		}
		if len(ttcFile.Fonts) == 0 {
			return nil, fmt.Errorf("TTC contains no fonts")
		}
		idx := fontIndex
		if idx < 0 {
			idx = 0
		}
		if idx >= len(ttcFile.Fonts) {
			return nil, fmt.Errorf("font index %d out of range (TTC has %d fonts)", idx, len(ttcFile.Fonts))
		}
		ttfBytes, err := ttc.ExtractFontBytes(ttcFile.Data, ttcFile.Fonts[idx])
		if err != nil {
			return nil, fmt.Errorf("extracting font %d from TTC: %w", idx, err)
		}
		return fontmetrics.LoadBytes(ttfBytes)
	}
	return fontmetrics.Load(path)
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}
