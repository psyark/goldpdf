package goldpdf

import (
	"image/color"
	"math"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

type Spaces struct {
	Left, Top, Right, Bottom float64
}

type BlockStyle struct {
	Margin          Spaces
	Padding         Spaces
	BackgroundColor color.Color
}

type TextFormat struct {
	Color           color.Color
	BackgroundColor color.Color
	FontSize        float64
	FontFamily      string
	Bold            bool
	Italic          bool
	Strike          bool
	Underline       bool
	Border          Border
}

type Border struct {
	Width  float64
	Color  color.Color
	Radius float64
}

func (s TextFormat) Apply(pdf fpdf.Pdf) {
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
	pdf.SetAlpha(1, "")
}

type Styler interface {
	Style(ast.Node, TextFormat) (BlockStyle, TextFormat)
}

type DefaultStyler struct {
	FontFamily string
	FontSize   float64
	Color      color.Color
}

func (s *DefaultStyler) Style(n ast.Node, format TextFormat) (BlockStyle, TextFormat) {
	style := BlockStyle{}

	if format.FontFamily == "" {
		format.FontFamily = s.FontFamily
	}
	if format.FontSize == 0 {
		format.FontSize = s.FontSize
	}
	if format.Color == nil {
		format.Color = s.Color
	}

	switch n := n.(type) {
	case *ast.Heading:
		format.FontSize = s.FontSize * math.Pow(1.15, float64(7-n.Level))
		style.Margin = Spaces{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
	case *ast.Paragraph:
		style.Margin = Spaces{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
	case *ast.Blockquote:
		style.Padding = Spaces{Left: 10}
		style.Margin = Spaces{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
	case *ast.Link, *ast.AutoLink:
		format.Color = color.RGBA{B: 0xFF, A: 0xFF}
		format.Underline = true
	case *ast.Emphasis:
		switch n.Level {
		case 2:
			format.Bold = true
		default:
			format.Italic = true
		}
	case *ast.FencedCodeBlock, *ast.CodeSpan:
		format.Color = color.Black
		format.BackgroundColor = color.Gray{Y: 0xF2}
		format.Border = Border{Width: 0.5, Color: color.Gray{Y: 0x99}, Radius: 3}
	case *xast.Strikethrough:
		format.Strike = true
	case *xast.TableHeader:
		format.BackgroundColor = color.Gray{Y: 0x80}
	case *xast.TableCell:
		format.Border = Border{Width: 1, Color: s.Color}
	}
	return style, format
}
