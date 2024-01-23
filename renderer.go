package goldpdf

import (
	"fmt"
	"image/color"
	"io"

	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type PDFProvider func() *gofpdf.Fpdf

type Renderer struct {
	source      []byte
	pdfProvider PDFProvider
	styler      Styler
	imageLoader ImageLoader
}

func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	if n.Type() != ast.TypeDocument {
		return fmt.Errorf("called with a node other than Document: %s", n.Kind())
	}

	fpdf := r.pdfProvider()
	fpdf.AddPage()

	r.source = source

	lm, tm, rm, _ := fpdf.GetMargins()
	pw, _ := fpdf.GetPageSize()

	rect := Rect{
		Left:  lm,
		Right: pw - rm,
		Top:   VerticalCoord{Page: 0, Position: tm},
	}
	if _, err := r.renderBlockNode(n, &renderContextImpl{fpdf: fpdf}, rect); err != nil {
		return err
	}

	return fpdf.Output(w)
}

func (r *Renderer) blockStyleTextFormat(n ast.Node) (BlockStyle, TextFormat) {
	ancestors := []ast.Node{}
	for p := n; p != nil; p = p.Parent() {
		ancestors = append(ancestors, p)
	}

	var bs BlockStyle
	var tf TextFormat
	for i := range ancestors {
		bs, tf = r.styler.Style(ancestors[len(ancestors)-i-1], tf)
	}
	return bs, tf
}

func (r *Renderer) blockStyle(n ast.Node) BlockStyle {
	bs, _ := r.blockStyleTextFormat(n)
	return bs
}

func (r *Renderer) textFormat(n ast.Node) TextFormat {
	_, tf := r.blockStyleTextFormat(n)
	return tf
}

// AddOptions does nothing
func (r *Renderer) AddOptions(options ...renderer.Option) {}

type Option func(*Renderer)

func New(options ...Option) renderer.Renderer {
	r := &Renderer{
		pdfProvider: func() *gofpdf.Fpdf { return gofpdf.New(gofpdf.OrientationPortrait, "pt", "A4", ".") },
		styler:      &DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.Black},
		imageLoader: &DefaultImageLoader{},
	}
	for _, option := range options {
		option(r)
	}
	return r
}

func WithPDFProvider(pdfProvider PDFProvider) Option {
	return func(r *Renderer) { r.pdfProvider = pdfProvider }
}

func WithStyler(styler Styler) Option {
	return func(r *Renderer) { r.styler = styler }
}

func WithImageLoader(imageLoader ImageLoader) Option {
	return func(r *Renderer) { r.imageLoader = imageLoader }
}
