package goldpdf

import (
	"image/color"
	"math"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

type BlockStyle struct {
	Margin          Spacing
	Padding         Spacing
	BackgroundColor color.Color
	Border          Border
	TextAlign       xast.Alignment
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

type Styler interface {
	Style(ast.Node, TextFormat) (BlockStyle, TextFormat)
}

var _ Styler = &DefaultStyler{}

type DefaultStyler struct {
	FontFamily string
	FontSize   float64
	Color      color.Color
}

func (s *DefaultStyler) Style(n ast.Node, tf TextFormat) (BlockStyle, TextFormat) {
	bs := BlockStyle{TextAlign: xast.AlignNone}

	switch n := n.(type) {
	case *ast.Document:
		tf.FontFamily = s.FontFamily
		tf.FontSize = s.FontSize
		tf.Color = s.Color
	case *ast.Heading:
		tf.FontSize = s.FontSize * math.Pow(1.15, float64(7-n.Level))
		bs.Margin = Spacing{Top: tf.FontSize / 2, Bottom: tf.FontSize / 2}
	case *ast.Paragraph:
		bs.Margin = Spacing{Top: tf.FontSize / 2, Bottom: tf.FontSize / 2}
	case *ast.Blockquote:
		bs.Padding = Spacing{Left: 10}
		bs.Margin = Spacing{Top: tf.FontSize / 2, Bottom: tf.FontSize / 2}
		bs.Border = IndividualBorder{
			Left: BorderEdge{Color: color.Gray{Y: 0x80}, Width: 6},
		}
	case *ast.List:
		bs.Margin = Spacing{Top: tf.FontSize / 2, Bottom: tf.FontSize / 2}
	case *ast.ListItem:
		bs.Padding = Spacing{Left: 16}
	case *ast.Link, *ast.AutoLink:
		tf.Color = color.RGBA{B: 0xFF, A: 0xFF}
		tf.Underline = true
	case *ast.Emphasis:
		switch n.Level {
		case 2:
			tf.Bold = true
		default:
			tf.Italic = true
		}
	case *ast.CodeSpan:
		tf.BackgroundColor = color.Gray{Y: 0xF2}
		tf.Border = UniformBorder{Width: 0.5, Color: color.Gray{Y: 0x99}, Radius: 3}
	case *ast.FencedCodeBlock:
		bs.BackgroundColor = color.Gray{Y: 0xF2}
		bs.Margin = Spacing{Top: 10, Bottom: 10}
		bs.Border = UniformBorder{Width: 0.5, Color: color.Gray{Y: 0x99}, Radius: 3}
		bs.Padding = Spacing{Top: 10, Left: 10, Bottom: 10, Right: 10}
	case *ast.ThematicBreak:
		bs.Margin = Spacing{Top: 19, Bottom: 19}
		bs.Border = IndividualBorder{
			Top: BorderEdge{Width: 2, Color: color.Gray{Y: 0x80}},
		}
	case *xast.Strikethrough:
		tf.Strike = true
	case *xast.Table:
		bs.Margin = Spacing{Top: 10, Bottom: 10}
	case *xast.TableHeader:
		tf.Bold = true
	case *xast.TableRow:
		edgeColor := color.Gray{Y: uint8(255 * 0.8)}
		if _, ok := n.PreviousSibling().(*xast.TableHeader); ok {
			edgeColor.Y = uint8(255 * 0.2)
		}
		bs.Border = IndividualBorder{
			Top: BorderEdge{Color: edgeColor, Width: 0.5},
		}
	case *xast.TableCell:
		bs.Padding = Spacing{Left: 10, Top: 10, Right: 10, Bottom: 10}
		colIndex := countPrevSiblings(n)
		if tr := n.Parent(); tr != nil {
			if table, ok := tr.Parent().(*xast.Table); ok {
				bs.TextAlign = table.Alignments[colIndex]
			}
		}
	}
	return bs, tf
}
