package goldpdf

import (
	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, entering bool) (ast.WalkStatus, error) {
	if entering {
		fs := r.currentState().Style.FontSize
		y := r.pdf.GetY()
		lm, _, rm, _ := r.pdf.GetMargins()
		width, _ := r.pdf.GetPageSize()
		r.pdf.Ln(fs)
		r.pdf.Line(lm, y, width-rm, y)
		r.pdf.Ln(fs)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(n *ast.Heading, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(0)
	} else {
		s := r.currentState()
		r.pdf.Ln(s.Style.FontSize)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderFencedCodeBlock(n *ast.FencedCodeBlock, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		s.Style.Apply(r.pdf)

		code := ""
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			code += string(line.Value(r.source))
		}

		r.pdf.Ln(0)
		r.pdf.Write(s.Style.FontSize, string(code))
	} else {
		s := r.currentState()
		r.pdf.Ln(s.Style.FontSize)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderParagraph(n *ast.Paragraph, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(0)
	} else {
		s := r.currentState()
		r.pdf.Ln(s.Style.FontSize)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTextBlock(n *ast.TextBlock, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(0)
	} else {
		s := r.currentState()
		r.pdf.Ln(s.Style.FontSize)
	}
	return ast.WalkContinue, nil
}