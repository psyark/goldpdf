package goldpdf

import (
	"image/color"

	"github.com/yuin/goldmark/ast"
)

// drawBlockQuote は Blockquoteを描画し、高さを返します
func (r *Renderer) drawBlockQuote(n *ast.Blockquote, draw bool, rs RenderState) (float64, error) {
	h, err := r.renderGenericBlockNode(n, draw, RenderState{X: rs.X + 10, Y: rs.Y, W: rs.W - 10})
	if err != nil {
		return 0, err
	}

	if draw {
		r.pdf.DrawLine(rs.X+2, rs.Y, rs.X+2, rs.Y+h, color.Gray{Y: 0x80}, 4)
	}

	return h, nil
}

// drawThematicBreak は ThematicBreakを描画し、高さを返します
func (r *Renderer) drawThematicBreak(n *ast.ThematicBreak, draw bool, rs RenderState) (float64, error) {
	if draw {
		r.pdf.DrawLine(rs.X, rs.Y+20, rs.X+rs.W, rs.Y+20, color.Gray{Y: 0x80}, 2)
	}
	return 40, nil
}
