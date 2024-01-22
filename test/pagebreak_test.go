package goldpdftest

import (
	"bytes"
	"image/color"
	"os"
	"testing"

	"github.com/psyark/goldpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
)

type pageBreakStyler struct {
	*goldpdf.DefaultStyler
}

func (s *pageBreakStyler) Style(n ast.Node, tf goldpdf.TextFormat) (goldpdf.BlockStyle, goldpdf.TextFormat) {
	bs, tf := s.DefaultStyler.Style(n, tf)
	switch n.(type) {
	case *ast.List:
		bs.Padding = goldpdf.Spacing{Left: 20, Top: 20, Right: 20, Bottom: 20}
		if _, ok := n.Parent().(*ast.Document); ok {
			bs.BackgroundColor = color.RGBA{R: 0xFF, G: 0xF4, B: 0xF4, A: 0xFF}
			bs.Border = goldpdf.IndividualBorder{
				Left: goldpdf.BorderEdge{Color: color.RGBA{R: 0xFF, G: 0xCC, B: 0xCC, A: 0xFF}, Width: 10},
			}
		} else {
			bs.BackgroundColor = color.RGBA{R: 0xF4, G: 0xFF, B: 0xF4, A: 0xFF}
			bs.Border = goldpdf.IndividualBorder{
				Left: goldpdf.BorderEdge{Color: color.RGBA{R: 0xCC, G: 0xFF, B: 0xCC, A: 0xFF}, Width: 10},
			}
		}
		bs.Padding.Left = 20
	case *ast.ListItem:
		bs.Margin.Bottom = 20
		bs.Border = goldpdf.UniformBorder{Color: color.Gray{Y: 0xDD}, Width: 1, Radius: 10}
		bs.BackgroundColor = color.White
		bs.Padding = goldpdf.Spacing{Left: 20, Right: 20, Top: 20, Bottom: 20}
	}
	return bs, tf
}

func TestPageBreak(t *testing.T) {
	md, err := os.ReadFile("testdata/pagebreak.md")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)

	markdown := goldmark.New(
		goldmark.WithExtensions(
			extension.Strikethrough,
			extension.Table,
		),
		goldmark.WithRenderer(
			goldpdf.New(
				goldpdf.WithStyler(&pageBreakStyler{&goldpdf.DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.Black}}),
			),
		),
	)

	if err := markdown.Convert(md, buf); err != nil {
		t.Fatal(err)
	}

	if err := CompareAndOutputResults(buf.Bytes(), "testdata/pagebreak.pdf", "testdata/pagebreak.png", "testdata/pagebreak_got.png", "testdata/pagebreak_diff.png"); err != nil {
		t.Fatal(err)
	}
}
