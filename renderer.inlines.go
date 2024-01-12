package goldpdf

import (
	"bytes"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderText(n *ast.Text, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.drawText(string(n.Text(r.source)))
	}
	return ast.WalkSkipChildren, nil
}

func (r *Renderer) renderLink(n *ast.Link, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Link = string(n.Destination)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderAutoLink(n *ast.AutoLink, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Link = string(n.URL(r.source))
		r.drawText(s.Link)
	}
	return ast.WalkSkipChildren, nil
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
		}
	}
	return ast.WalkSkipChildren, nil
}
