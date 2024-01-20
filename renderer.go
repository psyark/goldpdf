package goldpdf

import (
	"fmt"
	"image/color"
	"io"

	"github.com/jung-kurt/gofpdf"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type Renderer struct {
	pdfProvider PDFProvider
	source      []byte
	styler      Styler
	imageLoader imageLoader
}

func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	if n.Type() != ast.TypeDocument {
		return fmt.Errorf("想定しないノード")
	}

	fpdf := r.pdfProvider()
	fpdf.AddPage()

	r.source = source

	lm, tm, rm, _ := fpdf.GetMargins()
	pw, _ := fpdf.GetPageSize()

	rc := RenderContext{X: lm, Y: tm, W: pw - lm - rm, Target: &pdfImpl{fpdf: fpdf}}

	if _, err := r.renderBlockNode(n, rc); err != nil {
		return err
	}
	return fpdf.Output(w)
}

// AddOptions does nothing
func (r *Renderer) AddOptions(options ...renderer.Option) {
}

type Option func(*Renderer)

func New(options ...Option) renderer.Renderer {
	r := &Renderer{
		pdfProvider: func() *gofpdf.Fpdf { return gofpdf.New(gofpdf.OrientationPortrait, "pt", "A4", ".") },
		styler:      &DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.Black},
	}
	for _, option := range options {
		option(r)
	}
	return r
}

type PDFProvider func() *gofpdf.Fpdf

func WithPDFProvider(pdfProvider PDFProvider) Option {
	return func(r *Renderer) { r.pdfProvider = pdfProvider }
}

func WithStyler(styler Styler) Option {
	return func(r *Renderer) { r.styler = styler }
}
