package fontmetrics

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/go-go-golems/font-util/pkg/ttc"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type Metrics struct {
	FontName   string `json:"fontName"`
	UnitsPerEm int    `json:"unitsPerEm"`
	GlyphCount int    `json:"glyphCount"`
	Ascender   int    `json:"ascender"`
	Descender  int    `json:"descender"`
	LineGap    int    `json:"lineGap"`
	XHeight    int    `json:"xHeight"`
	CapHeight  int    `json:"capHeight"`
	Source     string `json:"source"`
}

type Loaded struct {
	Bytes   []byte
	Font    *sfnt.Font
	Metrics Metrics
}

func Load(path string) (*Loaded, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadBytes(b)
}

func LoadBytes(b []byte) (*Loaded, error) {
	f, err := sfnt.Parse(b)
	if err != nil {
		return nil, err
	}
	m, err := Extract(b, f)
	if err != nil {
		return nil, err
	}
	return &Loaded{Bytes: b, Font: f, Metrics: m}, nil
}

// LoadFromTTC loads a specific font from a TrueType Collection (.ttc).
// fontIndex is 0-based.
// Uses opentype.ParseCollection for the sfnt.Font, then extracts
// the font's raw bytes via ttc.ExtractFontBytes so that OS/2 metrics
// are read from the correct font (not the first font in the collection).
func LoadFromTTC(data []byte, fontIndex int) (*Loaded, error) {
	coll, err := opentype.ParseCollection(data)
	if err != nil {
		return nil, fmt.Errorf("parsing TTC: %w", err)
	}
	if fontIndex < 0 || fontIndex >= coll.NumFonts() {
		return nil, fmt.Errorf("font index %d out of range (TTC has %d fonts)", fontIndex, coll.NumFonts())
	}

	f, err := coll.Font(fontIndex)
	if err != nil {
		return nil, fmt.Errorf("loading font %d from collection: %w", fontIndex, err)
	}

	// Parse the TTC with our own parser to get the FontEntry for extraction.
	// This gives us the raw standalone TTF bytes so parseOS2 reads the
	// correct font's OS/2 table.
	ttcFile, err := ttc.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing TTC for extraction: %w", err)
	}
	if fontIndex >= len(ttcFile.Fonts) {
		return nil, fmt.Errorf("font index %d out of range", fontIndex)
	}
	fontBytes, err := ttc.ExtractFontBytes(data, ttcFile.Fonts[fontIndex])
	if err != nil {
		return nil, fmt.Errorf("extracting font %d bytes: %w", fontIndex, err)
	}

	m, err := Extract(fontBytes, f)
	if err != nil {
		return nil, err
	}

	return &Loaded{Bytes: fontBytes, Font: f, Metrics: m}, nil
}

func Extract(data []byte, f *sfnt.Font) (Metrics, error) {
	var buf sfnt.Buffer
	upem := int(f.UnitsPerEm())
	ppem := fixed.Int26_6(f.UnitsPerEm())
	fm, err := f.Metrics(&buf, ppem, font.HintingNone)
	if err != nil {
		return Metrics{}, err
	}
	name, _ := f.Name(&buf, sfnt.NameIDFull)
	m := Metrics{FontName: name, UnitsPerEm: upem, GlyphCount: f.NumGlyphs(), Ascender: fixedToInt(fm.Ascent), Descender: -fixedToInt(fm.Descent), LineGap: fixedToInt(fm.Height - fm.Ascent - fm.Descent), Source: "sfnt+fallback"}
	if os2, ok := parseOS2(data); ok {
		if os2.TypoAscender != 0 {
			m.Ascender = int(os2.TypoAscender)
		}
		if os2.TypoDescender != 0 {
			m.Descender = int(os2.TypoDescender)
		}
		m.LineGap = int(os2.TypoLineGap)
		if os2.XHeight != 0 {
			m.XHeight = int(os2.XHeight)
			m.Source = "os2"
		}
		if os2.CapHeight != 0 {
			m.CapHeight = int(os2.CapHeight)
			m.Source = "os2"
		}
	}
	if m.XHeight == 0 {
		m.XHeight = glyphTop(f, 'x', upem)
	}
	if m.CapHeight == 0 {
		m.CapHeight = glyphTop(f, 'H', upem)
	}
	if m.XHeight == 0 {
		m.XHeight = upem / 2
	}
	if m.CapHeight == 0 {
		m.CapHeight = upem * 7 / 10
	}
	return m, nil
}

func fixedToInt(v fixed.Int26_6) int { return int(v.Round()) }

func glyphTop(f *sfnt.Font, r rune, upem int) int {
	var buf sfnt.Buffer
	gid, err := f.GlyphIndex(&buf, r)
	if err != nil || gid == 0 {
		return 0
	}
	bounds, _, err := f.GlyphBounds(&buf, gid, fixed.Int26_6(upem<<6), font.HintingNone)
	if err != nil {
		return 0
	}
	return fixedToInt(bounds.Max.Y)
}

type os2Metrics struct{ TypoAscender, TypoDescender, TypoLineGap, XHeight, CapHeight int16 }

func parseOS2(data []byte) (os2Metrics, bool) {
	if len(data) < 12 {
		return os2Metrics{}, false
	}
	num := int(binary.BigEndian.Uint16(data[4:6]))
	for i := 0; i < num; i++ {
		o := 12 + i*16
		if o+16 > len(data) {
			return os2Metrics{}, false
		}
		if string(data[o:o+4]) != "OS/2" {
			continue
		}
		off := int(binary.BigEndian.Uint32(data[o+8 : o+12]))
		ln := int(binary.BigEndian.Uint32(data[o+12 : o+16]))
		if off < 0 || ln < 0 || off+ln > len(data) || ln < 74 {
			return os2Metrics{}, false
		}
		t := data[off : off+ln]
		m := os2Metrics{TypoAscender: i16(t, 68), TypoDescender: i16(t, 70), TypoLineGap: i16(t, 72)}
		if len(t) >= 90 {
			m.XHeight = i16(t, 86)
			m.CapHeight = i16(t, 88)
		}
		return m, true
	}
	return os2Metrics{}, false
}

func i16(b []byte, off int) int16 {
	if off+2 > len(b) {
		return 0
	}
	return int16(binary.BigEndian.Uint16(b[off : off+2]))
}

func Scale(m Metrics, pointSize float64) float64 { return pointSize / float64(m.UnitsPerEm) }

func Validate(m Metrics) error {
	if m.UnitsPerEm <= 0 {
		return fmt.Errorf("invalid units per em: %d", m.UnitsPerEm)
	}
	return nil
}
