package goldpdf

import (
	"image/color"
	"math"

	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

type BlockStyle struct {
	Margin          Spacing
	Padding         Spacing
	BackgroundColor color.Color
	Border          Border
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
	Border          UniformBorder
}

func (s TextFormat) Apply(pdf gofpdf.Fpdf) {
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
	Style(ast.Node) (BlockStyle, TextFormat)
}

var _ Styler = &DefaultStyler{}

type DefaultStyler struct {
	FontFamily string
	FontSize   float64
	Color      color.Color
}

func (s *DefaultStyler) Style(n ast.Node) (BlockStyle, TextFormat) {
	ancestors := []ast.Node{}
	for p := n; p != nil; p = p.Parent() {
		ancestors = append(ancestors, p)
	}

	var bs BlockStyle
	var tf TextFormat
	for i := range ancestors {
		bs, tf = s.style(ancestors[len(ancestors)-i-1], tf)
	}
	return bs, tf
}

func (s *DefaultStyler) style(n ast.Node, format TextFormat) (BlockStyle, TextFormat) {
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
		style.Margin = Spacing{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
	case *ast.Paragraph:
		style.Margin = Spacing{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
	case *ast.Blockquote:
		style.Padding = Spacing{Left: 10}
		style.Margin = Spacing{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
		style.Border = IndividualBorder{
			Left: BorderEdge{Color: color.Gray{Y: 0x80}, Width: 6},
		}
	case *ast.List:
		style.Margin = Spacing{Top: format.FontSize / 2, Bottom: format.FontSize / 2}
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
	case *ast.CodeSpan:
		format.BackgroundColor = color.Gray{Y: 0xF2}
		format.Border = UniformBorder{Width: 0.5, Color: color.Gray{Y: 0x99}, Radius: 3}
	case *ast.FencedCodeBlock:
		style.BackgroundColor = color.Gray{Y: 0xF2}
		style.Margin = Spacing{Top: 10, Bottom: 10}
		style.Border = UniformBorder{Width: 0.5, Color: color.Gray{Y: 0x99}, Radius: 3}
		style.Padding = Spacing{Top: 10, Left: 10, Bottom: 10, Right: 10}
	case *xast.Strikethrough:
		format.Strike = true
	case *xast.Table:
		style.Margin = Spacing{Top: 10, Bottom: 10}
	case *xast.TableHeader:
		format.Bold = true
	case *xast.TableRow:
		edgeColor := color.Gray{Y: uint8(255 * 0.8)}
		if _, ok := n.PreviousSibling().(*xast.TableHeader); ok {
			edgeColor.Y = uint8(255 * 0.2)
		}
		style.Border = IndividualBorder{
			Top: BorderEdge{Color: edgeColor, Width: 0.5},
		}
	case *xast.TableCell:
		style.Padding = Spacing{Left: 10, Top: 10, Right: 10, Bottom: 10}
	}
	return style, format
}
