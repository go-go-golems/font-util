package layout

import (
	"fmt"

	"github.com/go-go-golems/font-util/pkg/fontmetrics"
	"github.com/go-go-golems/font-util/pkg/shape"
	"github.com/go-go-golems/font-util/pkg/spec"
)

type Document struct {
	PageSizeName string              `json:"pageSizeName"`
	PageWidth    float64             `json:"pageWidth"`
	PageHeight   float64             `json:"pageHeight"`
	Pages        []Page              `json:"pages"`
	Metrics      fontmetrics.Metrics `json:"metrics"`
}

type Page struct {
	Number int   `json:"number"`
	Rows   []Row `json:"rows"`
}

type Row struct {
	Section   string  `json:"section"`
	BaselineY float64 `json:"baselineY"`
	LeftX     float64 `json:"leftX"`
	RightX    float64 `json:"rightX"`
	PointSize float64 `json:"pointSize"`
	Model     bool    `json:"model"`
	Items     []Item  `json:"items"`
}

type Item struct {
	Text string    `json:"text"`
	X    float64   `json:"x"`
	Run  shape.Run `json:"run"`
}

func Build(rs spec.Resolved, m fontmetrics.Metrics, sh shape.Shaper) (Document, error) {
	w, h := pageSize(rs.Page.Size, rs.Page.Orientation)
	doc := Document{PageSizeName: rs.Page.Size, PageWidth: w, PageHeight: h, Metrics: m, Pages: []Page{{Number: 1}}}
	scale := rs.PointSize / float64(m.UnitsPerEm)
	above := max(float64(m.Ascender), float64(m.CapHeight)) * scale
	below := -float64(m.Descender) * scale
	if below < 0 {
		below = 0
	}
	rowHeight := above + below + rs.RowGap
	y := rs.Margins.Top + above
	left, right := rs.Margins.Left, w-rs.Margins.Right
	addRow := func(row Row) { doc.Pages[len(doc.Pages)-1].Rows = append(doc.Pages[len(doc.Pages)-1].Rows, row) }
	newPageIfNeeded := func() {
		if y+below > h-rs.Margins.Bottom {
			doc.Pages = append(doc.Pages, Page{Number: len(doc.Pages) + 1})
			y = rs.Margins.Top + above
		}
	}
	for _, sec := range rs.Sections {
		for _, rowSpec := range sec.Rows {
			pt := rs.PointSize
			if rowSpec.PointSize != "" {
				parsed, err := spec.ParseLengthPt(rowSpec.PointSize)
				if err != nil {
					return doc, err
				}
				pt = parsed
			}
			opts := shape.Options{PointSize: pt, Kern: rs.Shaping.Features["kern"], Liga: rs.Shaping.Features["liga"]}
			blanks := rs.Layout.PracticeLinesAfterModel
			if rowSpec.BlankLines != nil {
				blanks = *rowSpec.BlankLines
			}
			if len(rowSpec.Items) > 0 {
				if rs.Layout.Mode == "cells" {
					for _, chunkItems := range chunk(rowSpec.Items, rs.Layout.Columns) {
						newPageIfNeeded()
						items, err := placeCellItems(chunkItems, left, right, rs.Layout.Columns, rs.CellGap, sh, opts)
						if err != nil {
							return doc, fmt.Errorf("section %q: %w", sec.Title, err)
						}
						addRow(Row{Section: sec.Title, BaselineY: y, LeftX: left, RightX: right, PointSize: pt, Model: true, Items: items})
						y += rowHeight
						for i := 0; i < blanks; i++ {
							newPageIfNeeded()
							addRow(Row{Section: sec.Title, BaselineY: y, LeftX: left, RightX: right, PointSize: pt})
							y += rowHeight
						}
					}
					continue
				}
				rowChunks, err := placeRowItems(rowSpec.Items, left, right, rs.ItemGap, rs.Layout.Wrap, sh, opts)
				if err != nil {
					return doc, fmt.Errorf("section %q: %w", sec.Title, err)
				}
				for _, items := range rowChunks {
					newPageIfNeeded()
					addRow(Row{Section: sec.Title, BaselineY: y, LeftX: left, RightX: right, PointSize: pt, Model: true, Items: items})
					y += rowHeight
					for i := 0; i < blanks; i++ {
						newPageIfNeeded()
						addRow(Row{Section: sec.Title, BaselineY: y, LeftX: left, RightX: right, PointSize: pt})
						y += rowHeight
					}
				}
				continue
			}
			if blanks == 0 {
				blanks = 1
			}
			for i := 0; i < blanks; i++ {
				newPageIfNeeded()
				addRow(Row{Section: sec.Title, BaselineY: y, LeftX: left, RightX: right, PointSize: pt})
				y += rowHeight
			}
		}
	}
	return doc, nil
}

func placeCellItems(texts []string, left, right float64, columns int, cellGap float64, sh shape.Shaper, opts shape.Options) ([]Item, error) {
	items := make([]Item, 0, len(texts))
	if columns < 1 {
		columns = 1
	}
	cellW := (right - left - float64(columns-1)*cellGap) / float64(columns)
	for i, t := range texts {
		r, err := sh.Shape(t, opts)
		if err != nil {
			return nil, err
		}
		cellX := left + float64(i)*(cellW+cellGap)
		x := cellX + max(0, (cellW-r.AdvancePt)/2)
		items = append(items, Item{Text: t, X: x, Run: r})
	}
	return items, nil
}

func placeRowItems(texts []string, left, right, gap float64, wrap bool, sh shape.Shaper, opts shape.Options) ([][]Item, error) {
	var rows [][]Item
	var current []Item
	x := left
	flush := func() {
		if len(current) > 0 {
			rows = append(rows, current)
			current = nil
		}
		x = left
	}
	for _, t := range texts {
		r, err := sh.Shape(t, opts)
		if err != nil {
			return nil, err
		}
		if wrap && len(current) > 0 && x+r.AdvancePt > right {
			flush()
		}
		current = append(current, Item{Text: t, X: x, Run: r})
		x += r.AdvancePt + gap
	}
	flush()
	return rows, nil
}

func chunk(items []string, n int) [][]string {
	if n < 1 {
		n = 1
	}
	var chunks [][]string
	for len(items) > 0 {
		end := n
		if len(items) < end {
			end = len(items)
		}
		chunks = append(chunks, items[:end])
		items = items[end:]
	}
	return chunks
}

func pageSize(name, orient string) (float64, float64) {
	var w, h float64
	switch name {
	case "Letter", "letter":
		w, h = 612, 792
	default:
		w, h = 595.28, 841.89
	}
	if orient == "landscape" {
		return h, w
	}
	return w, h
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
