package shape

import (
	"bytes"
	"unicode/utf8"

	gtfont "github.com/go-text/typesetting/font"
	"github.com/go-text/typesetting/harfbuzz"
	"github.com/go-go-golems/font-util/pkg/fontmetrics"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type Options struct {
	PointSize float64
	Kern      bool
	Liga      bool
}

type Run struct {
	Text          string  `json:"text"`
	Glyphs        []Glyph `json:"glyphs"`
	AdvancePt     float64 `json:"advancePt"`
	MissingGlyphs int     `json:"missingGlyphs"`
	Engine        string  `json:"engine"`
	Note          string  `json:"note,omitempty"`
}

type Glyph struct {
	Rune       string  `json:"rune,omitempty"`
	Cluster    int     `json:"cluster"`
	GlyphID    uint16  `json:"glyphId"`
	XAdvancePt float64 `json:"xAdvancePt"`
	XOffsetPt  float64 `json:"xOffsetPt"`
}

type Shaper struct {
	Font    *sfnt.Font
	Metrics fontmetrics.Metrics
	hbFace  *gtfont.Face
}

func New(f *sfnt.Font, m fontmetrics.Metrics) Shaper { return Shaper{Font: f, Metrics: m} }

func NewWithBytes(data []byte, f *sfnt.Font, m fontmetrics.Metrics) Shaper {
	s := New(f, m)
	if face, err := gtfont.ParseTTF(bytes.NewReader(data)); err == nil {
		s.hbFace = face
	}
	return s
}

func (s Shaper) Shape(text string, opts Options) (Run, error) {
	if s.hbFace != nil {
		return s.shapeHarfbuzz(text, opts)
	}
	return s.shapeSFNT(text, opts)
}

func (s Shaper) shapeHarfbuzz(text string, opts Options) (Run, error) {
	buf := harfbuzz.NewBuffer()
	runes := []rune(text)
	buf.AddRunes(runes, 0, len(runes))
	buf.GuessSegmentProperties()
	hbf := harfbuzz.NewFont(s.hbFace)
	scale := int32(opts.PointSize * 64)
	hbf.XScale, hbf.YScale = scale, scale
	features := []harfbuzz.Feature{}
	if !opts.Kern {
		if f, err := harfbuzz.ParseFeature("kern=0"); err == nil {
			features = append(features, f)
		}
	}
	if !opts.Liga {
		for _, name := range []string{"liga=0", "clig=0", "dlig=0"} {
			if f, err := harfbuzz.ParseFeature(name); err == nil {
				features = append(features, f)
			}
		}
	}
	buf.Shape(hbf, features)
	r := Run{Text: text, Engine: "go-text-harfbuzz", Note: "HarfBuzz-compatible shaping via github.com/go-text/typesetting/harfbuzz."}
	for i, info := range buf.Info {
		pos := buf.Pos[i]
		g := Glyph{Cluster: info.Cluster, GlyphID: uint16(info.Glyph), XAdvancePt: float64(pos.XAdvance) / 64.0, XOffsetPt: float64(pos.XOffset) / 64.0}
		r.Glyphs = append(r.Glyphs, g)
		r.AdvancePt += g.XAdvancePt
		if info.Glyph == 0 {
			r.MissingGlyphs++
		}
	}
	return r, nil
}

func (s Shaper) shapeSFNT(text string, opts Options) (Run, error) {
	var buf sfnt.Buffer
	scale := opts.PointSize / float64(s.Metrics.UnitsPerEm)
	ppem := fixed.Int26_6(s.Metrics.UnitsPerEm << 6)
	r := Run{Text: text, Engine: "sfnt-kern-mvp", Note: "Fallback shaper maps runes to glyphs and applies legacy kern pairs; full GSUB ligature shaping requires HarfBuzz."}
	var prev sfnt.GlyphIndex
	cluster := 0
	for len(text) > 0 {
		ru, size := utf8.DecodeRuneInString(text)
		text = text[size:]
		gid, err := s.Font.GlyphIndex(&buf, ru)
		if err != nil {
			return r, err
		}
		if gid == 0 {
			r.MissingGlyphs++
		}
		kernPt := 0.0
		if opts.Kern && prev != 0 && gid != 0 {
			k, _ := s.Font.Kern(&buf, prev, gid, ppem, font.HintingNone)
			kernPt = float64(k) / 64.0 * scale
		}
		adv, err := s.Font.GlyphAdvance(&buf, gid, ppem, font.HintingNone)
		if err != nil {
			return r, err
		}
		advPt := float64(adv) / 64.0 * scale
		r.Glyphs = append(r.Glyphs, Glyph{Rune: string(ru), Cluster: cluster, GlyphID: uint16(gid), XAdvancePt: advPt, XOffsetPt: kernPt})
		r.AdvancePt += kernPt + advPt
		prev = gid
		cluster++
	}
	return r, nil
}
