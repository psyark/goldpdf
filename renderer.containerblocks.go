package goldpdf

import (
	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderBlockquote(n *ast.Blockquote, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.currentState().XMin += 10
		r.pdf.SetX(r.currentState().XMin)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(n *ast.List, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(n *ast.ListItem, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Circle(r.currentState().XMin+10, r.pdf.GetY()+5, 2, "F")
		r.currentState().XMin += 20
		r.pdf.SetX(r.currentState().XMin)
	}
	return ast.WalkContinue, nil
}
