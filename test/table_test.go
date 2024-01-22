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
	xast "github.com/yuin/goldmark/extension/ast"
)

type customStyler struct {
	*goldpdf.DefaultStyler
}

func (s *customStyler) Style(n ast.Node, tf goldpdf.TextFormat) (goldpdf.BlockStyle, goldpdf.TextFormat) {
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

func TestTable(t *testing.T) {
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
				goldpdf.WithStyler(&customStyler{&goldpdf.DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.Black}}),
			),
		),
	)

	if err := markdown.Convert(md, buf); err != nil {
		t.Fatal(err)
	}

	if err := CompareAndOutputResults(buf.Bytes(), "testdata/styled_table.pdf", "testdata/styled_table.png", "testdata/styled_table_got.png", "testdata/styled_table_diff.png"); err != nil {
		t.Fatal(err)
	}
}
