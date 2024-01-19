package goldpdf

import (
	"image/color"

	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderBlockQuote(n *ast.Blockquote, rc RenderContext) (float64, error) {
	h, err := r.renderGenericBlockNode(n, rc.Extend(10, 0, -10))
	if err != nil {
		return 0, err
	}

	if !rc.Preflight {
		rc.Target.DrawLine(rc.X+2, rc.Y, rc.X+2, rc.Y+h, color.Gray{Y: 0x80}, 4)
	}

	return h, nil
}

func (r *Renderer) renderFencedCodeBlock(n *ast.FencedCodeBlock, rc RenderContext) (float64, error) {
	bs, tf := r.styler.Style(n, TextFormat{})
	elements := []FlowElement{}

	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		ts := &TextSpan{Text: string(line.Value(r.source)), Format: tf}
		elements = append(elements, ts, &HardBreak{})
	}

	height := bs.Margin.Vertical() + bs.Padding.Vertical() + bs.Border.Width*2
	borderBox := rc.Extend(
		bs.Margin.Left,
		bs.Margin.Top,
		-bs.Margin.Horizontal(),
	)
	contentBox := borderBox.Extend(
		bs.Border.Width+bs.Padding.Left,
		bs.Border.Width+bs.Padding.Top,
		-bs.Border.Width*2-bs.Padding.Horizontal(),
	)

	if len(elements) != 0 {
		if !rc.Preflight {
			if contentHeight, err := r.renderFlowElements(elements, contentBox.InPreflight()); err != nil {
				return 0, err
			} else {
				rc.Target.DrawRect(borderBox.X, borderBox.Y, borderBox.W, contentHeight+bs.Padding.Vertical()+bs.Border.Width*2, bs.BackgroundColor, bs.Border)
			}
		}

		if contentHeight, err := r.renderFlowElements(elements, contentBox); err != nil {
			return 0, err
		} else {
			height += contentHeight
		}
	}

	return height, nil
}

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, rc RenderContext) (float64, error) {
	if !rc.Preflight {
		rc.Target.DrawLine(rc.X, rc.Y+20, rc.X+rc.W, rc.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}
