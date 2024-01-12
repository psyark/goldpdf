package goldpdf

import (
	"bytes"

	"github.com/go-pdf/fpdf"
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

func (r *Renderer) renderAutoLink(n *ast.AutoLink, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Link = string(n.URL(r.source))
		s.Style.Color = r.styles.LinkColor
		s.Style.Apply(r.pdf)
		r.pdf.WriteLinkString(s.Style.FontSize, s.Link, s.Link)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderImage(n *ast.Image, entering bool) (ast.WalkStatus, error) {
	if entering {
		if info := r.imageLoader.load(string(n.Destination)); info != nil {
			r.pdf.RegisterImageOptionsReader(
				info.Name,
				fpdf.ImageOptions{ImageType: info.Type},
				bytes.NewReader(info.Data),
			)

			x, y := r.pdf.GetXY()
			r.pdf.ImageOptions(info.Name, x, y, float64(info.Width), float64(info.Height), false, fpdf.ImageOptions{}, 0, "")
			// TODO: コンテンツ領域の高さを親に伝達
			return ast.WalkSkipChildren, nil
		}
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
