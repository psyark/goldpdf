package goldpdf

import (
	"image/color"

	"github.com/yuin/goldmark/ast"
)

// drawBlockQuote は Blockquoteを描画し、高さを返します
func (r *Renderer) drawBlockQuote(n *ast.Blockquote, rc RenderContext) (float64, error) {
	h, err := r.renderGenericBlockNode(n, rc.Extend(10, 0, -10))
	if err != nil {
		return 0, err
	}

	if !rc.Preflight {
		rc.Target.DrawLine(rc.X+2, rc.Y, rc.X+2, rc.Y+h, color.Gray{Y: 0x80}, 4)
	}

	return h, nil
}

// drawThematicBreak は ThematicBreakを描画し、高さを返します
func (r *Renderer) drawThematicBreak(n *ast.ThematicBreak, rc RenderContext) (float64, error) {
	if !rc.Preflight {
		rc.Target.DrawLine(rc.X, rc.Y+20, rc.X+rc.W, rc.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}
