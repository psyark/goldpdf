package goldpdf

import (
	"image/color"

	"github.com/yuin/goldmark/ast"
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
	bs, tf := r.styler.Style(n, TextFormat{})

	if !borderBox.Preflight {
		h, err := r.renderFencedCodeBlock(n, borderBox.InPreflight())
		if err != nil {
			return 0, err
		}
		borderBox.Target.DrawRect(
			borderBox.X,
			borderBox.Y,
			borderBox.W,
			h,
			bs.BackgroundColor,
			bs.Border,
		)
	}

	elements := []FlowElement{}
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		ts := &TextSpan{Text: string(line.Value(r.source)), Format: tf}
		elements = append(elements, ts, &HardBreak{})
	}

	contentBox := borderBox.Extend(
		bs.Border.Width+bs.Padding.Left,
		bs.Border.Width+bs.Padding.Top,
		-bs.Border.Width*2-bs.Padding.Horizontal(),
	)

	height, err := r.renderFlowElements(elements, contentBox)
	if err != nil {
		return 0, err
	}

	height += bs.Padding.Vertical() + bs.Border.Width*2
	return height, nil
}

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, borderBox RenderContext) (float64, error) {
	if !borderBox.Preflight {
		borderBox.Target.DrawLine(borderBox.X, borderBox.Y+20, borderBox.X+borderBox.W, borderBox.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}
