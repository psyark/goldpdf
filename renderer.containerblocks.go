package goldpdf

import (
	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderBlockquote(n *ast.Blockquote, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.currentState().XMin += 10
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(n *ast.List, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(n *ast.ListItem, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.SetAlpha(1, "")
		r.pdf.SetFillColor(0x80, 0x80, 0x80)
		r.pdf.Circle(r.currentState().XMin+10, r.pdf.GetY()+5, 2, "F")
		r.currentState().XMin += 20
	}
	return ast.WalkContinue, nil
}
