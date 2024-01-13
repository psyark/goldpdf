package goldpdf

import (
	"github.com/yuin/goldmark/ast"
)

func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := r.currentState()
		y := r.pdf.GetY()

		r.pdf.SetDrawColor(0x80, 0x80, 0x80)
		r.pdf.SetLineWidth(2)

		r.pdf.Line(s.XMin, y+10, s.XMax, y+10)
		r.pdf.Ln(20)
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
		code := ""
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			code += string(line.Value(r.source))
		}

		// TODO ボーダー・背景をインラインではなくブロックレベルで描画
		r.pdf.Ln(0)
		r.drawText(code)
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
