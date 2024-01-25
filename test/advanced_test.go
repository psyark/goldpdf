package goldpdftest

import (
	"bytes"
	_ "embed"
	"image/color"
	"os"
	"testing"

	"github.com/psyark/goldpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	xast "github.com/yuin/goldmark/extension/ast"
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

func TestAdvancedPageBreak(t *testing.T) {
	md, err := os.ReadFile("testdata/advanced/pagebreak.md")
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

	err = CompareAndOutputResults(
		buf.Bytes(),
		"testdata/advanced/pagebreak.pdf",
		"testdata/advanced/pagebreak.png",
		"testdata/advanced/pagebreak_got.png",
		"testdata/advanced/pagebreak_diff.png",
	)
	if err != nil {
		t.Fatal(err)
	}
}

type tableStyler struct{ *goldpdf.DefaultStyler }

func (s *tableStyler) Style(n ast.Node, tf goldpdf.TextFormat) (goldpdf.BlockStyle, goldpdf.TextFormat) {
	bs, tf := s.DefaultStyler.Style(n, tf)
	switch n.(type) {
	case *xast.Table:
		bs.Border = goldpdf.UniformBorder{Width: 1, Color: color.Black, Radius: 5}
		bs.Padding = goldpdf.Spacing{Left: 10, Right: 10, Top: 10, Bottom: 10}
		bs.BackgroundColor = color.RGBA{R: 0xFF, A: 0xFF}
	case *xast.TableRow, *xast.TableHeader:
		bs.Border = goldpdf.UniformBorder{Width: 1, Color: color.Black, Radius: 5}
		bs.Padding = goldpdf.Spacing{Left: 10, Right: 10, Top: 10, Bottom: 10}
		bs.BackgroundColor = color.RGBA{G: 0xFF, A: 0xFF}
	case *xast.TableCell:
		bs.Border = goldpdf.UniformBorder{Width: 1, Color: color.Black, Radius: 5}
		bs.Padding = goldpdf.Spacing{Left: 10, Right: 10, Top: 10, Bottom: 10}
		bs.BackgroundColor = color.RGBA{B: 0xFF, A: 0xFF}
	}
	return bs, tf
}

func TestAdvancedStyledTable(t *testing.T) {
	md, err := os.ReadFile("testdata/basic/table.md")
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
				goldpdf.WithStyler(&tableStyler{&goldpdf.DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.Black}}),
			),
		),
	)

	if err := markdown.Convert(md, buf); err != nil {
		t.Fatal(err)
	}

	err = CompareAndOutputResults(
		buf.Bytes(),
		"testdata/advanced/styled_table.pdf",
		"testdata/advanced/styled_table.png",
		"testdata/advanced/styled_table_got.png",
		"testdata/advanced/styled_table_diff.png",
	)
	if err != nil {
		t.Fatal(err)
	}
}
