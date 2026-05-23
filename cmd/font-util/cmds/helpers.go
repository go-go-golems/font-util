package cmds

import (
	"os"
	"strings"

	"github.com/go-go-golems/font-util/pkg/fontmetrics"
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
// For TTC files, uses fontmetrics.LoadFromTTC (opentype.ParseCollection +
// in-memory extraction for correct OS/2 metrics).
// fontIndex selects which font in a TTC to use (0-based).
func loadFont(path string, fontIndex int) (*fontmetrics.Loaded, error) {
	if isTTCFile(path) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return fontmetrics.LoadFromTTC(data, fontIndex)
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
