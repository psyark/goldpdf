package goldpdf

import (
	"image/color"
	"math"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

type Style struct {
	Color      color.Color
	FontSize   float64
	FontFamily string
	Bold       bool
	Italic     bool
	Strike     bool
	Underline  bool
}

func (s Style) Apply(pdf *fpdf.Fpdf) {
	fontStyle := ""
	if s.Bold {
		fontStyle += "B"
	}
	if s.Italic {
		fontStyle += "I"
	}
	if s.Strike {
		fontStyle += "S"
	}
	if s.Underline {
		fontStyle += "U"
	}
	pdf.SetFont(s.FontFamily, fontStyle, s.FontSize)
	cr, cg, cb, _ := s.Color.RGBA()
	pdf.SetTextColor(int(cr>>8), int(cg>>8), int(cb>>8))
}

type Styler interface {
	Style(Style, ast.Node) Style
}

type DefaultStyler struct {
	FontFamily string
	FontSize   float64
	Color      color.Color
}

func (s *DefaultStyler) Style(current Style, n ast.Node) Style {
	switch n := n.(type) {
	case *ast.Document:
		current.FontFamily = s.FontFamily
		current.FontSize = s.FontSize
		current.Color = s.Color
	case *ast.Heading:
		current.FontSize = s.FontSize * math.Pow(1.15, float64(7-n.Level))
	case *ast.Link:
		current.Color = color.RGBA{B: 0xFF, A: 0xFF}
		current.Underline = true
	case *ast.Emphasis:
		switch n.Level {
		case 2:
			current.Bold = true
		default:
			current.Italic = true
		}
	case *ast.CodeBlock, *ast.CodeSpan:
		current.Color = color.RGBA{R: 0x99, G: 0x99, B: 0, A: 255}
	case *xast.Strikethrough:
		current.Strike = true
	}
	return current
}
