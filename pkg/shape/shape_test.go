package shape

import (
	"testing"

	"github.com/go-go-golems/font-util/pkg/fontmetrics"
	"golang.org/x/image/font/gofont/goregular"
)

func TestHarfbuzzShaperRuns(t *testing.T) {
	loaded, err := fontmetrics.LoadBytes(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	sh := NewWithBytes(loaded.Bytes, loaded.Font, loaded.Metrics)
	run, err := sh.Shape("office", Options{PointSize: 54, Kern: true, Liga: true})
	if err != nil {
		t.Fatal(err)
	}
	if run.Engine != "go-text-harfbuzz" {
		t.Fatalf("engine=%s", run.Engine)
	}
	if len(run.Glyphs) == 0 || run.AdvancePt <= 0 {
		t.Fatalf("bad run: %+v", run)
	}
}
