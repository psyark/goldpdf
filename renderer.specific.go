package goldpdf

import (
	"fmt"
	"image/color"
	"math"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

func (r *Renderer) renderListItem(n ast.Node, mc MeasureContext, borderBox HalfBounds) (Rect, error) {
	rect, err := r.renderGenericBlockNode(n, mc, borderBox)
	if err != nil {
		return Rect{}, err
	}

	err = mc.GetRenderContext(func(rc RenderContext) error {
		bs := r.blockStyle(n)
		contentBox := borderBox.Shrink(bs.Border, bs.Padding) // ListItemのコンテンツボックス

		// ListItemの最初のブロックノード
		n2 := n.FirstChild()
		bs2 := r.blockStyle(n2)
		contentBox2 := contentBox.Shrink(bs2.Margin, bs2.Border, bs2.Padding)

		elements, err := r.getFlowElements(n2)
		if err != nil {
			return err
		}
		_, h := elements[0][0].size(mc)

		if list, ok := n.Parent().(*ast.List); ok && list.IsOrdered() {
			ts := &TextElement{
				Format: r.textFormat(n),
				Text:   fmt.Sprintf("%d.", countPrevSiblings(n)+1),
			}
			rc.DrawTextSpan(contentBox.Top.Page, contentBox2.Left-15, contentBox2.Top.Position, ts)
		} else {
			rc.DrawBullet(contentBox.Top.Page, contentBox2.Left-10, contentBox2.Top.Position+h/2, color.Black, 2)
		}

		return nil
	})
	if err != nil {
		return rect, err
	}

	return rect, nil
}

func (r *Renderer) renderTable(n *xast.Table, mc MeasureContext, borderBox HalfBounds) (Rect, error) {
	bs := r.blockStyle(n)

	err := mc.GetRenderContext(func(rc RenderContext) error {
		rect, err := r.renderTable(n, mc, borderBox)
		if err != nil {
			return err
		}
		rc.DrawBox(rect, bs.BackgroundColor, bs.Border)
		return nil
	})
	if err != nil {
		return Rect{}, err
	}

	columnContentWidth := make([]float64, len(n.Alignments))

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		colIndex := 0
		for col := row.FirstChild(); col != nil; col = col.NextSibling() {
			elements, err := r.getFlowElements(col)
			if err != nil {
				return Rect{}, err
			}
			columnContentWidth[colIndex] = math.Max(columnContentWidth[colIndex], elements.Width(mc))
			colIndex++
		}
	}

	contentBox := borderBox.Shrink(bs.Border, bs.Padding)

	totalWidth := 0.0
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

	for _, ccw := range columnContentWidth {
		totalWidth += ccw
	}
	// 列の最大幅がavailableWidthを超過するなら
	if totalWidth > availableWidth {
		// 列の最大幅を均等倍率で縮小する
		for i := range columnContentWidth {
			columnContentWidth[i] *= availableWidth / totalWidth
		}
	}

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch row := row.(type) {
		case *xast.TableHeader, *xast.TableRow:
			bs := r.blockStyle(row)
			rowRect, err := r.renderTableRow(row, mc, contentBox.Shrink(bs.Margin), columnContentWidth)
			if err != nil {
				return Rect{}, err
			}
			contentBox.Top = rowRect.Bottom
			contentBox.Top.Position += bottom(bs.Margin)
		}
	}

	boxBottom := contentBox.Top
	boxBottom.Position += bottom(bs.Border) + bottom(bs.Padding)

	return borderBox.ToRect(boxBottom), nil
}

func (r *Renderer) renderTableRow(n ast.Node, mc MeasureContext, borderBox HalfBounds, columnContentWidth []float64) (Rect, error) {
	switch n.Kind() {
	case xast.KindTableHeader, xast.KindTableRow:
	default:
		return Rect{}, fmt.Errorf("unsupported kind: %v", n.Kind())
	}

	bs := r.blockStyle(n)
	var borderBoxBottom []VerticalCoord

	err := mc.GetRenderContext(func(rc RenderContext) error {
		rowRect, err := r.renderTableRow(n, mc, borderBox, columnContentWidth)
		if err != nil {
			return err
		}

		rc.DrawBox(rowRect, bs.BackgroundColor, bs.Border)

		rowRect.Bottom.Position -= bottom(bs.Border) + bottom(bs.Padding)
		borderBoxBottom = []VerticalCoord{rowRect.Bottom} // ここで各テーブルセルの高さを行と一致させる
		return nil
	})
	if err != nil {
		return Rect{}, err
	}

	contentBox := borderBox.ToRect(borderBox.Top).Shrink(bs.Border, bs.Padding)

	for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
		bs := r.blockStyle(cell)
		contentBox.Left += bs.Margin.Left
		contentBox.Right = contentBox.Left + columnContentWidth[countPrevSiblings(cell)] + horizontal(bs.Border) + horizontal(bs.Padding)

		cellRect, err := r.renderGenericBlockNode(cell, mc, contentBox.ToHalfBounds(), borderBoxBottom...)
		if err != nil {
			return Rect{}, err
		}

		contentBox.Left = contentBox.Right + bs.Margin.Right
		if contentBox.Bottom.LessThan(cellRect.Bottom) {
			contentBox.Bottom = cellRect.Bottom
		}
	}

	contentBox.Bottom.Position += bottom(bs.Border) + bottom(bs.Padding)
	return borderBox.ToRect(contentBox.Bottom), nil
}

func countPrevSiblings(n ast.Node) int {
	c := 0
	for x := n.PreviousSibling(); x != nil; x = x.PreviousSibling() {
		c++
	}
	return c
}
