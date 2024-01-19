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

	return r.renderGenericBlockNode(n, borderBox.WithFlowElements(elements))
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
			_, tf := r.styler.Style(n, TextFormat{})
			ts := &TextSpan{
				Format: tf,
				Text:   fmt.Sprintf("%d.", countPrevSiblings(n)+1),
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
	bf, tf := r.styler.Style(n, TextFormat{})

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

	contentBox := borderBox.Extend(
		bf.Border.Width,
		bf.Border.Width,
		-bf.Border.Width*2,
	)

	totalWidth := 0.0
	availableWidth := contentBox.W
	if row := n.FirstChild(); row != nil {
		bf, _ := r.styler.Style(row, TextFormat{})
		availableWidth -= bf.Border.Width*2 + bf.Padding.Horizontal()
		if col := row.FirstChild(); col != nil {
			bf, _ := r.styler.Style(col, TextFormat{})
			availableWidth -= (bf.Border.Width*2 + bf.Padding.Horizontal()) * float64(len(n.Alignments))
		}
	}

	for _, w := range cellWidths {
		totalWidth += w
	}
	if totalWidth > availableWidth {
		for i := range cellWidths {
			cellWidths[i] *= availableWidth / totalWidth
		}
	}

	height := 0.0
	for row := n.FirstChild(); row != nil; row = row.NextSibling() {
		switch row := row.(type) {
		case *xast.TableHeader, *xast.TableRow:
			h, err := r.renderTableRow(row, borderBox.Extend(0, height, 0), cellWidths)
			if err != nil {
				return 0, err
			}
			height += h
		}
	}

	return height, nil
}

func (r *Renderer) renderTableRow(n ast.Node, borderBox RenderContext, cellWidths []float64) (float64, error) {
	colIndex := 0
	height := 0.0

	cellBox := borderBox

	for col := n.FirstChild(); col != nil; col = col.NextSibling() {
		tf, _ := r.styler.Style(col, TextFormat{})

		cellBox.W = cellWidths[colIndex] + tf.Border.Width*2 + tf.Padding.Horizontal()

		h, err := r.renderBlockNode(col, cellBox)
		if err != nil {
			return 0, err
		}

		height = math.Max(height, h)

		cellBox.X += cellBox.W
		colIndex++
	}

	return height, nil
}

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, borderBox RenderContext) (float64, error) {
	if !borderBox.Preflight {
		borderBox.Target.DrawLine(borderBox.X, borderBox.Y+20, borderBox.X+borderBox.W, borderBox.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}

func countPrevSiblings(n ast.Node) int {
	c := 0
	for x := n.PreviousSibling(); x != nil; x = x.PreviousSibling() {
		c++
	}
	return c
}
