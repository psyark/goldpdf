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

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, borderBox RenderContext) (float64, error) {
	if !borderBox.Preflight {
		borderBox.Target.DrawLine(borderBox.X, borderBox.Y+20, borderBox.X+borderBox.W, borderBox.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}
