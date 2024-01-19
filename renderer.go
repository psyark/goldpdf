package goldpdf

import (
	"fmt"
	"image/color"
	"io"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
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

func (rc RenderContext) InPreflight() RenderContext {
	rc.Preflight = true
	return rc
}

// renderBlockNode draws a block node (or document node) inside a borderBox
// and returns the height of its border box.
func (r *Renderer) renderBlockNode(n ast.Node, borderBox RenderContext) (float64, error) {
	if n.Type() == ast.TypeInline {
		return 0, fmt.Errorf("renderBlockNode has been called with an inline node: %v > %v", n.Parent().Kind(), n.Kind())
	}

	switch n := n.(type) {
	case *ast.Blockquote:
		return r.renderBlockQuote(n, borderBox)
	case *ast.FencedCodeBlock:
		return r.renderFencedCodeBlock(n, borderBox)
	case *ast.ListItem:
		return r.renderListItem(n, borderBox)
	case *ast.ThematicBreak:
		return r.renderThematicBreak(n, borderBox)
	case *xast.Table:
		return r.renderTable(n, borderBox)
	default:
		return r.renderGenericBlockNode(n, borderBox)
	}
}

// renderGenericBlockNode provides basic rendering for all block nodes
// except specific block nodes.
func (r *Renderer) renderGenericBlockNode(n ast.Node, borderBox RenderContext, additionalElements ...FlowElement) (float64, error) {
	bs, tf := r.styler.Style(n, TextFormat{})

	if !borderBox.Preflight {
		h, err := r.renderGenericBlockNode(n, borderBox.InPreflight(), additionalElements...)
		if err != nil {
			return 0, err
		}
		borderBox.Target.DrawRect(
			borderBox.X,
			borderBox.Y,
			borderBox.W,
			h,
			bs.BackgroundColor,
			bs.Border,
		)
	}

	contentBox := borderBox.Extend(
		bs.Border.Width+bs.Padding.Left,
		bs.Border.Width+bs.Padding.Top,
		-bs.Border.Width*2-bs.Padding.Horizontal(),
	)

	height := bs.Border.Width + bs.Padding.Top
	elements := additionalElements

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Type() {
		case ast.TypeBlock:
			bs2, _ := r.styler.Style(c, TextFormat{})

			// TODO マージンの相殺
			height += bs2.Margin.Top
			if h, err := r.renderBlockNode(c, contentBox.Extend(0, height, 0)); err != nil {
				return 0, err
			} else {
				borderBox.Y += h
				height += h
			}
			height += bs2.Margin.Bottom
		case ast.TypeInline:
			if e, err := r.getFlowElements(c, tf); err != nil {
				return 0, err
			} else {
				elements = append(elements, e...)
			}
		}
	}

	if len(elements) != 0 {
		if h, err := r.renderFlowElements(elements, contentBox); err != nil {
			return 0, err
		} else {
			height += h
		}
	}

	height += bs.Padding.Bottom + bs.Border.Width
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
