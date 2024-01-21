package goldpdf

import (
	"fmt"
	"image/color"
	"math"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

type columnFormat struct {
	contentWidth float64
}

func (r *Renderer) renderListItem(n *ast.ListItem, borderBox RenderContext) (float64, error) {
	h, err := r.renderGenericBlockNode(n, borderBox, nil)
	if err != nil {
		return 0, err
	}

	err = borderBox.Preflight(func() error {
		list, ok := n.Parent().(*ast.List)

		// 最初の要素の余白を考慮
		if n.FirstChild() != nil {
			bs, _ := r.style(n.FirstChild())
			borderBox.Y += top(bs.Margin) + top(bs.Border) + top(bs.Padding)
		}

		if ok && list.IsOrdered() {
			_, tf := r.style(n)
			ts := &TextSpan{
				Format: tf,
				Text:   fmt.Sprintf("%d.", countPrevSiblings(n)+1),
			}
			borderBox.Target.DrawTextSpan(borderBox.X, borderBox.Y, ts)
		} else {
			borderBox.Target.DrawBullet(borderBox.X+4, borderBox.Y+h/2, color.Black, 2)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return h, nil
}

func (r *Renderer) renderTable(n *xast.Table, borderBox RenderContext) (float64, error) {
	bs, _ := r.style(n)

	err := borderBox.Preflight(func() error {
		h, err := r.renderTable(n, borderBox)
		if err != nil {
			return err
		}
		borderBox.Target.DrawBox(borderBox.X, borderBox.Y, borderBox.W, h, bs.BackgroundColor, bs.Border)
		return nil
	})
	if err != nil {
		return 0, err
	}

	columnFormats := make([]columnFormat, len(n.Alignments))

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		colIndex := 0
		for col := row.FirstChild(); col != nil; col = col.NextSibling() {
			elements, err := r.getFlowElements(col)
			if err != nil {
				return 0, err
			}

			contentWidth := borderBox.Target.GetNaturalWidth(elements)
			columnFormats[colIndex].contentWidth = math.Max(columnFormats[colIndex].contentWidth, contentWidth)
			colIndex++
		}
	}

	contentBox := borderBox.Shrink(bs.Border, bs.Padding)

	totalWidth := 0.0
	availableWidth := contentBox.W
	if row := n.FirstChild(); row != nil {
		// TableHeaderの水平成分を減らす
		bs, _ := r.style(row)
		availableWidth -= horizontal(bs.Margin) + horizontal(bs.Border) + horizontal(bs.Padding)
		if col := row.FirstChild(); col != nil {
			// TableCellの水平成分を減らす
			bs, _ := r.style(col)
			availableWidth -= (horizontal(bs.Margin) + horizontal(bs.Border) + horizontal(bs.Padding)) * float64(len(n.Alignments))
		}
	}

	for _, cf := range columnFormats {
		totalWidth += cf.contentWidth
	}
	// 列の最大幅がavailableWidthを超過するなら
	if totalWidth > availableWidth {
		// 列の最大幅を均等倍率で縮小する
		for i := range columnFormats {
			columnFormats[i].contentWidth *= availableWidth / totalWidth
		}
	}

	height := 0.0

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch row := row.(type) {
		case *xast.TableHeader, *xast.TableRow:
			bs, _ := r.style(row)
			h, err := r.renderTableRow(row, contentBox.MoveDown(height).Shrink(bs.Margin), columnFormats)
			if err != nil {
				return 0, err
			}
			height += h
		}
	}
	height += horizontal(bs.Border) + horizontal(bs.Padding)

	return height, nil
}

func (r *Renderer) renderTableRow(n ast.Node, borderBox RenderContext, columnFormats []columnFormat) (float64, error) {
	switch n.Kind() {
	case xast.KindTableHeader, xast.KindTableRow:
	default:
		return 0, fmt.Errorf("unsupported kind: %v", n.Kind())
	}

	bs, _ := r.style(n)

	options := &rgbnOption{}
	err := borderBox.Preflight(func() error {
		h, err := r.renderTableRow(n, borderBox, columnFormats)
		if err != nil {
			return err
		}

		borderBox.Target.DrawBox(borderBox.X, borderBox.Y, borderBox.W, h, bs.BackgroundColor, bs.Border)
		options.forceHeight = h - vertical(bs.Border) - vertical(bs.Padding)
		return nil
	})
	if err != nil {
		return 0, err
	}

	contentBox := borderBox.Shrink(bs.Border, bs.Padding)
	height := 0.0

	for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
		tf, _ := r.style(cell)
		contentBox.W = columnFormats[countPrevSiblings(cell)].contentWidth + horizontal(tf.Border) + horizontal(tf.Padding)

		h, err := r.renderGenericBlockNode(cell, contentBox, options)
		if err != nil {
			return 0, err
		}

		height = math.Max(height, h)
		contentBox.X += contentBox.W
	}

	height += vertical(bs.Border) + vertical(bs.Padding)
	return height, nil
}

func countPrevSiblings(n ast.Node) int {
	c := 0
	for x := n.PreviousSibling(); x != nil; x = x.PreviousSibling() {
		c++
	}
	return c
}
