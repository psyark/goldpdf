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

func (r *Renderer) renderListItem(n ast.Node, mc MeasureContext, borderBox Rect) (Rect, error) {
	rect, err := r.renderGenericBlockNode(n, mc, borderBox, false)
	if err != nil {
		return rect, err
	}

	err = mc.GetRenderContext(func(rc RenderContext) error {
		bs := r.blockStyle(n)
		contentBox := borderBox.Shrink(bs.Border, bs.Padding) // ListItemのコンテンツボックス

		// ListItemの最初のブロックノード
		n2 := n.FirstChild()
		bs2 := r.blockStyle(n2)
		contentBox2 := contentBox.Shrink(bs2.Margin, bs2.Border, bs2.Padding)

		elements := r.getFlowElements(n2)
		_, h := elements[0][0].size(mc)

		if list, ok := n.Parent().(*ast.List); ok && list.IsOrdered() {
			ts := &TextSpan{
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

func (r *Renderer) renderTable(n *xast.Table, mc MeasureContext, borderBox Rect) (Rect, error) {
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

	columnFormats := make([]columnFormat, len(n.Alignments))

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		colIndex := 0
		for col := row.FirstChild(); col != nil; col = col.NextSibling() {
			elements := r.getFlowElements(col)
			contentWidth := getNaturalWidth(elements, mc)
			columnFormats[colIndex].contentWidth = math.Max(columnFormats[colIndex].contentWidth, contentWidth)
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

	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch row := row.(type) {
		case *xast.TableHeader, *xast.TableRow:
			bs := r.blockStyle(row)
			rowRect, err := r.renderTableRow(row, mc, contentBox.Shrink(bs.Margin), columnFormats)
			if err != nil {
				return Rect{}, err
			}
			contentBox.Top = rowRect.Bottom
			contentBox.Top.Position += vertical(bs.Margin)
		}
	}

	borderBox.Bottom = contentBox.Top
	borderBox.Bottom.Position += horizontal(bs.Border) + horizontal(bs.Padding)

	return borderBox, nil
}

func (r *Renderer) renderTableRow(n ast.Node, mc MeasureContext, borderBox Rect, columnFormats []columnFormat) (Rect, error) {
	switch n.Kind() {
	case xast.KindTableHeader, xast.KindTableRow:
	default:
		return Rect{}, fmt.Errorf("unsupported kind: %v", n.Kind())
	}

	bs := r.blockStyle(n)

	err := mc.GetRenderContext(func(rc RenderContext) error {
		rowRect, err := r.renderTableRow(n, mc, borderBox, columnFormats)
		if err != nil {
			return err
		}

		rc.DrawBox(rowRect, bs.BackgroundColor, bs.Border)
		borderBox.Bottom = rowRect.Bottom // ここで各テーブルセルの高さを行と一致させる
		return nil
	})
	if err != nil {
		return Rect{}, err
	}

	contentBox := borderBox.Shrink(bs.Border, bs.Padding)

	for cell := n.FirstChild(); cell != nil; cell = cell.NextSibling() {
		bs := r.blockStyle(cell)
		contentBox.Left += bs.Margin.Left
		contentBox.Right = contentBox.Left + columnFormats[countPrevSiblings(cell)].contentWidth + horizontal(bs.Border) + horizontal(bs.Padding)

		cellRect, err := r.renderGenericBlockNode(cell, mc, contentBox, true)
		if err != nil {
			return Rect{}, err
		}

		contentBox.Left = contentBox.Right + bs.Margin.Right
		if contentBox.Bottom.LessThan(cellRect.Bottom) {
			contentBox.Bottom = cellRect.Bottom
		}
	}

	borderBox.Bottom = contentBox.Bottom
	borderBox.Bottom.Position += vertical(bs.Border) + vertical(bs.Padding)
	return borderBox, nil
}

func countPrevSiblings(n ast.Node) int {
	c := 0
	for x := n.PreviousSibling(); x != nil; x = x.PreviousSibling() {
		c++
	}
	return c
}
