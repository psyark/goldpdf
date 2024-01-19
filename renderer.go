package goldpdf

import (
	"fmt"
	"image/color"
	"io"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type Renderer struct {
	pdfProvider PDFProvider
	source      []byte
	styler      Styler
	imageLoader imageLoader
}

type RenderContext struct {
	X, Y, W   float64
	Preflight bool
	Target    PDF
}

func (rc RenderContext) Extend(dx, dy, dw float64) RenderContext {
	rc.X += dx
	rc.Y += dy
	rc.W += dw
	return rc
}

// renderBlockNode はブロックノード（またはドキュメントノード）を描画し、その高さを返します
// drawが偽のとき、描画は行わずにサイズだけを返します
func (r *Renderer) renderBlockNode(n ast.Node, rc RenderContext) (float64, error) {
	if n.Type() == ast.TypeInline {
		return 0, fmt.Errorf("renderBlockNode has been called with an inline node: %v > %v", n.Parent().Kind(), n.Kind())
	}

	switch n := n.(type) {
	case *ast.ThematicBreak:
		return r.drawThematicBreak(n, rc)
	case *ast.Blockquote:
		return r.drawBlockQuote(n, rc)
	default:
		return r.renderGenericBlockNode(n, rc)
	}
}

// renderGenericBlockNode はブロックノードに対する基本的なレンダリングを提供します
func (r *Renderer) renderGenericBlockNode(n ast.Node, rc RenderContext) (float64, error) {
	bs, tf := r.styler.Style(n, TextFormat{})
	rc.Y += bs.Margin.Top
	height := bs.Margin.Top
	elements := []FlowElement{}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Type() {
		case ast.TypeBlock:
			if h, err := r.renderBlockNode(c, rc); err != nil {
				return 0, err
			} else {
				rc.Y += h
				height += h
			}
		case ast.TypeInline:
			if e, err := r.getFlowElements(c, tf); err != nil {
				return 0, err
			} else {
				elements = append(elements, e...)
			}
		}
	}

	if len(elements) != 0 {
		if h, err := r.renderFlowElements(elements, rc); err != nil {
			return 0, err
		} else {
			height += h
		}
	}

	height += bs.Margin.Bottom
	return height, nil
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
		pdfProvider: func() *fpdf.Fpdf { return fpdf.New(fpdf.OrientationPortrait, "pt", "A4", ".") },
		styler:      &DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.Black},
	}
	for _, option := range options {
		option(r)
	}
	return r
}

type PDFProvider func() *fpdf.Fpdf

func WithPDFProvider(pdfProvider PDFProvider) Option {
	return func(r *Renderer) { r.pdfProvider = pdfProvider }
}

func WithStyler(styler Styler) Option {
	return func(r *Renderer) { r.styler = styler }
}
