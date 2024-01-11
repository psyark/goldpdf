package goldpdf

import (
	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

func (r *Renderer) renderText(n *ast.Text, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Style.Apply(r.pdf)
		if s.Link != "" {
			r.pdf.WriteLinkString(s.Style.FontSize, string(n.Text(r.source)), s.Link)
		} else {
			r.pdf.Write(s.Style.FontSize, string(n.Text(r.source)))
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderEmphasis(n *ast.Emphasis, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		switch n.Level {
		case 2:
			s.Style.Bold = true
		default:
			s.Style.Italic = true
		}
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderLink(n *ast.Link, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Link = string(n.Destination)
		s.Style.Color = r.styles.LinkColor
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderStrikethrough(n *xast.Strikethrough, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Style.Strike = true
	}
	return ast.WalkContinue, nil
}
