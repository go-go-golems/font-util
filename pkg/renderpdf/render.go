package renderpdf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/go-go-golems/font-util/pkg/fontmetrics"
	"github.com/go-go-golems/font-util/pkg/layout"
	"github.com/go-go-golems/font-util/pkg/shape"
	"github.com/go-go-golems/font-util/pkg/spec"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func Render(doc layout.Document, rs spec.Resolved, metrics fontmetrics.Metrics, font *sfnt.Font, out string) error {
	pdf := fpdf.New(unitOrient(rs.Page.Orientation), "pt", rs.Page.Size, "")
	pdf.SetMargins(rs.Margins.Left, rs.Margins.Top, rs.Margins.Right)
	pdf.SetAutoPageBreak(false, rs.Margins.Bottom)
	for _, p := range doc.Pages {
		pdf.AddPage()
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(90, 90, 90)
		pdf.Text(rs.Margins.Left, rs.Margins.Top-12, fmt.Sprintf("%s - page %d", metrics.FontName, p.Number))
		for _, row := range p.Rows {
			drawRow(pdf, row, rs, metrics, font)
		}
	}
	return pdf.OutputFileAndClose(out)
}

func unitOrient(o string) string {
	if o == "landscape" {
		return "L"
	}
	return "P"
}

func drawRow(pdf *fpdf.Fpdf, row layout.Row, rs spec.Resolved, m fontmetrics.Metrics, font *sfnt.Font) {
	scale := row.PointSize / float64(m.UnitsPerEm)
	lines := []struct {
		key, label string
		y          float64
	}{
		{"ascender", "asc", row.BaselineY - float64(m.Ascender)*scale},
		{"cap_height", "cap", row.BaselineY - float64(m.CapHeight)*scale},
		{"x_height", "x", row.BaselineY - float64(m.XHeight)*scale},
		{"baseline", "base", row.BaselineY},
		{"descender", "desc", row.BaselineY - float64(m.Descender)*scale},
	}
	for _, l := range lines {
		st := rs.Style.HelperLines[l.key]
		if !st.Show {
			continue
		}
		r, g, b := parseHex(st.Color)
		pdf.SetDrawColor(r, g, b)
		w, _ := spec.ParseLengthPt(st.Width)
		if w <= 0 {
			w = .4
		}
		pdf.SetLineWidth(w)
		if len(st.Dash) >= 2 {
			on, _ := spec.ParseLengthPt(st.Dash[0])
			off, _ := spec.ParseLengthPt(st.Dash[1])
			pdf.SetDashPattern([]float64{on, off}, 0)
		} else {
			pdf.SetDashPattern(nil, 0)
		}
		pdf.Line(row.LeftX, l.y, row.RightX, l.y)
		if rs.Style.Labels.Show {
			pdf.SetFont("Helvetica", "", rs.LabelSize)
			pdf.SetTextColor(r, g, b)
			pdf.Text(row.LeftX-24, l.y+2, l.label)
		}
	}
	pdf.SetDashPattern(nil, 0)
	if row.Model {
		pdf.SetFillColor(0, 0, 0)
		for _, it := range row.Items {
			drawRun(pdf, font, m, row.PointSize, it.X, row.BaselineY, it.Run.Glyphs)
		}
	}
}

func drawRun(pdf *fpdf.Fpdf, font *sfnt.Font, m fontmetrics.Metrics, pointSize, x, baseline float64, glyphs []shape.Glyph) {
	cursor := x
	for _, g := range glyphs {
		drawGlyph(pdf, font, m, pointSize, g.GlyphID, cursor+g.XOffsetPt, baseline)
		cursor += g.XAdvancePt
	}
}

func drawGlyph(pdf *fpdf.Fpdf, font *sfnt.Font, m fontmetrics.Metrics, pointSize float64, gid uint16, originX, originY float64) {
	if font == nil || gid == 0 {
		return
	}
	var buf sfnt.Buffer
	segments, err := font.LoadGlyph(&buf, sfnt.GlyphIndex(gid), fixed.Int26_6(m.UnitsPerEm<<6), nil)
	if err != nil || len(segments) == 0 {
		return
	}
	scale := pointSize / float64(m.UnitsPerEm)
	var current point
	hasPath := false
	convert := func(p fixed.Point26_6) point {
		// sfnt.LoadGlyph returns coordinates in the same Y-down convention used by
		// golang.org/x/image rasterization examples: negative Y values are above
		// the baseline for Latin fonts. PDF page coordinates are also Y-down in
		// fpdf's public API, so add the scaled glyph Y to the baseline origin.
		return point{originX + float64(p.X)/64.0*scale, originY + float64(p.Y)/64.0*scale}
	}
	for _, seg := range segments {
		switch seg.Op {
		case sfnt.SegmentOpMoveTo:
			if hasPath {
				pdf.ClosePath()
			}
			p := convert(seg.Args[0])
			pdf.MoveTo(p.X, p.Y)
			current = p
			hasPath = true
		case sfnt.SegmentOpLineTo:
			p := convert(seg.Args[0])
			pdf.LineTo(p.X, p.Y)
			current = p
		case sfnt.SegmentOpQuadTo:
			c := convert(seg.Args[0])
			end := convert(seg.Args[1])
			c1, c2 := quadToCubic(current, c, end)
			pdf.CurveBezierCubicTo(c1.X, c1.Y, c2.X, c2.Y, end.X, end.Y)
			current = end
		case sfnt.SegmentOpCubeTo:
			c1 := convert(seg.Args[0])
			c2 := convert(seg.Args[1])
			end := convert(seg.Args[2])
			pdf.CurveBezierCubicTo(c1.X, c1.Y, c2.X, c2.Y, end.X, end.Y)
			current = end
		}
	}
	if hasPath {
		pdf.ClosePath()
		pdf.DrawPath("f*")
	}
}

type point struct{ X, Y float64 }

func quadToCubic(p0, p1, p2 point) (point, point) {
	return point{p0.X + 2.0/3.0*(p1.X-p0.X), p0.Y + 2.0/3.0*(p1.Y-p0.Y)}, point{p2.X + 2.0/3.0*(p1.X-p2.X), p2.Y + 2.0/3.0*(p1.Y-p2.Y)}
}

func parseHex(s string) (int, int, int) {
	s = strings.TrimPrefix(strings.TrimSpace(s), "#")
	if len(s) != 6 {
		return 0, 0, 0
	}
	r, _ := strconv.ParseInt(s[0:2], 16, 0)
	g, _ := strconv.ParseInt(s[2:4], 16, 0)
	b, _ := strconv.ParseInt(s[4:6], 16, 0)
	return int(r), int(g), int(b)
}
