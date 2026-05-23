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
// Handles TTC files by extracting the first font to a temp file.
func loadFont(path string) (*fontmetrics.Loaded, error) {
	if isTTCFile(path) {
		ttcFile, err := ttc.ParseFile(path)
		if err != nil {
			return nil, fmt.Errorf("parsing TTC: %w", err)
		}
		if len(ttcFile.Fonts) == 0 {
			return nil, fmt.Errorf("TTC contains no fonts")
		}
		tmpFile, err := os.CreateTemp("", "font-util-*.ttf")
		if err != nil {
			return nil, err
		}
		defer func() { _ = os.Remove(tmpFile.Name()) }()
		if err := ttc.ExtractFont(ttcFile.Data, ttcFile.Fonts[0], tmpFile.Name()); err != nil {
			return nil, fmt.Errorf("extracting font from TTC: %w", err)
		}
		_ = tmpFile.Close()
		return fontmetrics.Load(tmpFile.Name())
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
