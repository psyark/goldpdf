package goldpdf

import "github.com/yuin/goldmark/ast"

func (r *Renderer) renderBlockquote(n *ast.Blockquote, entering bool) (ast.WalkStatus, error) {
	lm, _, _, _ := r.pdf.GetMargins()
	if entering {
		r.pdf.SetLeftMargin(lm + 10)
	} else {
		r.pdf.SetLeftMargin(lm - 10)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderList(n *ast.List, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) renderListItem(n *ast.ListItem, entering bool) (ast.WalkStatus, error) {
	lm, _, _, _ := r.pdf.GetMargins()
	if entering {
		r.pdf.Circle(lm+10, r.pdf.GetY()+5, 2, "F")
		r.pdf.SetLeftMargin(lm + 20)
	} else {
		r.pdf.SetLeftMargin(lm - 20)
	}
	return ast.WalkContinue, nil
}
