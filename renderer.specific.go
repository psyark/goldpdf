package goldpdf

import (
	"fmt"
	"image/color"
	"math"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

func (r *Renderer) renderBlockQuote(n *ast.Blockquote, borderBox RenderContext) (float64, error) {
	h, err := r.renderGenericBlockNode(n, borderBox.Extend(6, 0, -6))
	if err != nil {
		return 0, err
	}

	if !borderBox.Preflight {
		borderBox.Target.DrawLine(borderBox.X+3, borderBox.Y, borderBox.X+3, borderBox.Y+h, color.Gray{Y: 0x80}, 6)
	}

	return h, nil
}

func (r *Renderer) renderFencedCodeBlock(n *ast.FencedCodeBlock, borderBox RenderContext) (float64, error) {
	_, tf := r.styler.Style(n, TextFormat{})

	elements := []FlowElement{}
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		ts := &TextSpan{Text: string(line.Value(r.source)), Format: tf}
		elements = append(elements, ts, &HardBreak{})
	}

	return r.renderGenericBlockNode(n, borderBox, elements...)
}

func (r *Renderer) renderListItem(n *ast.ListItem, borderBox RenderContext) (float64, error) {
	h, err := r.renderGenericBlockNode(n, borderBox.Extend(16, 0, -16))
	if err != nil {
		return 0, err
	}

	if !borderBox.Preflight {
		list, ok := n.Parent().(*ast.List)

		// 最初の要素の余白を考慮
		if n.FirstChild() != nil {
			bs, _ := r.styler.Style(n.FirstChild(), TextFormat{})
			borderBox.Y += bs.Margin.Top + bs.Border.Width + bs.Padding.Top
		}

		if ok && list.IsOrdered() {
			index := 0
			for x := n.PreviousSibling(); x != nil; x = x.PreviousSibling() {
				index++
			}

			_, tf := r.styler.Style(n, TextFormat{})
			ts := &TextSpan{
				Format: tf,
				Text:   fmt.Sprintf("%d.", index+1),
			}
			borderBox.Target.DrawTextSpan(borderBox.X, borderBox.Y, ts)
		} else {
			borderBox.Target.DrawBullet(borderBox.X+4, borderBox.Y+h/2, color.Black, 2)
		}
	}

	return h, nil
}

func (r *Renderer) renderTable(n *xast.Table, borderBox RenderContext) (float64, error) {
	// TODO TableRow, TableCellのスタイル
	_, tf := r.styler.Style(n, TextFormat{})

	cellWidths := make([]float64, len(n.Alignments))

	// TODO TableRow, TableCellの余白やボーダー幅の考慮
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		colIndex := 0
		for col := row.FirstChild(); col != nil; col = col.NextSibling() {
			elements := FlowElements{}
			for c := col.FirstChild(); c != nil; c = c.NextSibling() {
				e, err := r.getFlowElements(c, tf)
				if err != nil {
					return 0, err
				}
				elements = append(elements, e...)
			}

			cellWidths[colIndex] = math.Max(
				cellWidths[colIndex],
				borderBox.Target.GetNaturalWidth(elements),
			)
			colIndex++
		}
	}

	totalWidth := 0.0
	for _, w := range cellWidths {
		totalWidth += w
	}
	if totalWidth > borderBox.W {
		for i := range cellWidths {
			cellWidths[i] *= borderBox.W / totalWidth
		}
	}

	height := 0.0
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		colIndex := 0

		rowHeight := 0.0

		cellBox := borderBox
		cellBox.Y = borderBox.Y + height

		for col := row.FirstChild(); col != nil; col = col.NextSibling() {
			elements := FlowElements{}
			for c := col.FirstChild(); c != nil; c = c.NextSibling() {
				e, err := r.getFlowElements(c, tf)
				if err != nil {
					return 0, err
				}
				elements = append(elements, e...)
			}

			cellBox.W = cellWidths[colIndex]

			h, err := r.renderFlowElements(elements, cellBox)
			if err != nil {
				return 0, err
			}
			rowHeight = math.Max(rowHeight, h)

			cellBox.X += cellWidths[colIndex]
			colIndex++
		}

		height += rowHeight
	}

	return height, nil
}

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, borderBox RenderContext) (float64, error) {
	if !borderBox.Preflight {
		borderBox.Target.DrawLine(borderBox.X, borderBox.Y+20, borderBox.X+borderBox.W, borderBox.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}
