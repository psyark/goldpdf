package goldpdf

import (
	"math"

	xast "github.com/yuin/goldmark/extension/ast"
)

type TableLayout func(r *Renderer, n *xast.Table, mc MeasureContext, borderBox HalfBounds) ([]float64, error)

var (
	_ TableLayout = TableLayoutEvenly
	_ TableLayout = TableLayoutAutoFilled
	_ TableLayout = TableLayoutAutoCompact
)

// TableLayoutEvenly is a TableLayout that expands to fill the table so that each column is of equal width,
// without considering the column contents
func TableLayoutEvenly(r *Renderer, n *xast.Table, mc MeasureContext, borderBox HalfBounds) ([]float64, error) {
	availableWidth := getAvailableCellContentWidths(r, n, borderBox)
	columnContentWidth := make([]float64, len(n.Alignments))
	for i := range columnContentWidth {
		columnContentWidth[i] = availableWidth / float64(len(n.Alignments))
	}
	return columnContentWidth, nil
}

// TableLayoutAutoFilled is a TableLayout that expands the column width to fill the table
// by a ratio based on the width of the column content
func TableLayoutAutoFilled(r *Renderer, n *xast.Table, mc MeasureContext, borderBox HalfBounds) ([]float64, error) {
	return tableLayoutAuto(r, n, mc, borderBox, true)
}

// TableLayoutAutoCompact determines the width of a column in proportion to the width of the column's content,
// but does not expand beyond the width required for the content
func TableLayoutAutoCompact(r *Renderer, n *xast.Table, mc MeasureContext, borderBox HalfBounds) ([]float64, error) {
	return tableLayoutAuto(r, n, mc, borderBox, false)
}

func tableLayoutAuto(r *Renderer, n *xast.Table, mc MeasureContext, borderBox HalfBounds, filled bool) ([]float64, error) {
	columnContentWidth := make([]float64, len(n.Alignments))

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		colIndex := 0
		for col := row.FirstChild(); col != nil; col = col.NextSibling() {
			elements, err := r.getFlowElements(col)
			if err != nil {
				return nil, err
			}
			columnContentWidth[colIndex] = math.Max(columnContentWidth[colIndex], elements.Width(mc))
			colIndex++
		}
	}

	availableWidth := getAvailableCellContentWidths(r, n, borderBox)

	totalWidth := 0.0
	for _, ccw := range columnContentWidth {
		totalWidth += ccw
	}

	// 列の最大幅がavailableWidthを超過するなら
	if filled || totalWidth > availableWidth {
		// 列の最大幅を均等倍率で縮小する
		for i := range columnContentWidth {
			columnContentWidth[i] *= availableWidth / totalWidth
		}
	}

	return columnContentWidth, nil
}

func getAvailableCellContentWidths(r *Renderer, n *xast.Table, borderBox HalfBounds) float64 {
	bs := r.blockStyle(n)
	contentBox := borderBox.Shrink(bs.Border, bs.Padding)

	availableWidth := contentBox.Right - contentBox.Left
	if row := n.FirstChild(); row != nil {
		// TableHeaderの水平成分を減らす
		bs := r.blockStyle(row)
		availableWidth -= horizontal(bs.Margin) + horizontal(bs.Border) + horizontal(bs.Padding)
		if col := row.FirstChild(); col != nil {
			// TableCellの水平成分を減らす
			bs := r.blockStyle(col)
			availableWidth -= (horizontal(bs.Margin) + horizontal(bs.Border) + horizontal(bs.Padding)) * float64(len(n.Alignments))
		}
	}
	return availableWidth
}
