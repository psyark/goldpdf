package goldpdf

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/go-pdf/fpdf"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
)

type Renderer struct {
	source []byte
	pdf    *fpdf.Fpdf
	states []State
	depth  int
	styles Styles
}

func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.source = source
	if n.Type() == ast.TypeDocument {
		if err := ast.Walk(n, r.walk); err != nil {
			return err
		}
	}
	return r.pdf.Output(w)
}

func (r *Renderer) walk(n ast.Node, entering bool) (ast.WalkStatus, error) {
	if entering {
		fmt.Println(strings.Repeat("  ", r.depth), n.Kind())
		r.depth++
	} else {
		r.depth--
	}

	switch n := n.(type) {
	case *ast.Document:
		return r.renderDocument(n, entering)
	case *ast.Heading:
		return r.renderHeading(n, entering)
	case *ast.Paragraph:
		return r.renderParagraph(n, entering)
	case *ast.Text:
		return r.renderText(n, entering)
	case *ast.ThematicBreak:
		return r.renderThematicBreak(n, entering)

	default:
		if entering {
			r.pdf.Ln(10)
			r.pdf.SetTextColor(255, 0, 0)
			r.pdf.Write(10, fmt.Sprintf("%v not implemented", n.Kind().String()))
			r.pdf.SetTextColor(0, 0, 0)
			r.pdf.Ln(10)
		}
		return ast.WalkContinue, nil
	}
}

func (r *Renderer) renderDocument(n *ast.Document, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pushState(State{Style: r.styles.Paragraph})
	} else {
		r.popState()
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderHeading(n *ast.Heading, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := State{}
		switch n.Level {
		case 1:
			s.Style = r.styles.H1
		case 2:
			s.Style = r.styles.H2
		case 3:
			s.Style = r.styles.H3
		case 4:
			s.Style = r.styles.H4
		case 5:
			s.Style = r.styles.H5
		case 6:
			s.Style = r.styles.H6
		}
		r.pushState(s)
		r.pdf.Ln(0)
	} else {
		s := r.popState()
		r.pdf.Ln(s.Style.FontSize / 2)
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderParagraph(n *ast.Paragraph, entering bool) (ast.WalkStatus, error) {
	if entering {
		s := State{Style: r.styles.Paragraph}
		r.pushState(s)
		r.pdf.Ln(0)
	} else {
		s := r.popState()
		r.pdf.Ln(s.Style.FontSize)
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderText(n *ast.Text, entering bool) (ast.WalkStatus, error) {
	if entering {
		fs := r.currentState().Style.FontSize
		r.pdf.SetFontSize(fs)
		r.pdf.Write(fs, string(n.Text(r.source)))
	}
	return ast.WalkContinue, nil
}
func (r *Renderer) renderThematicBreak(n *ast.ThematicBreak, entering bool) (ast.WalkStatus, error) {
	if entering {
		fs := r.currentState().Style.FontSize
		y := r.pdf.GetY()
		lm, _, rm, _ := r.pdf.GetMargins()
		width, _ := r.pdf.GetPageSize()
		r.pdf.Ln(fs)
		r.pdf.Line(lm, y, width-rm, y)
		r.pdf.Ln(fs)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) pushState(state State) {
	r.states = append(r.states, state)
}
func (r *Renderer) popState() State {
	var s State
	s, r.states = r.states[len(r.states)-1], r.states[:len(r.states)-1]
	return s
}
func (r *Renderer) currentState() State {
	return r.states[len(r.states)-1]
}

// AddOptions does nothing
func (r *Renderer) AddOptions(options ...renderer.Option) {
}

func New() renderer.Renderer {
	pdf := fpdf.New(fpdf.OrientationPortrait, "mm", "A4", ".")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 16)

	return &Renderer{
		pdf:    pdf,
		states: []State{},
		styles: Styles{
			Paragraph: Style{FontSize: 12 * math.Pow(1.15, 0)},
			H1:        Style{FontSize: 12 * math.Pow(1.15, 6)},
			H2:        Style{FontSize: 12 * math.Pow(1.15, 5)},
			H3:        Style{FontSize: 12 * math.Pow(1.15, 4)},
			H4:        Style{FontSize: 12 * math.Pow(1.15, 3)},
			H5:        Style{FontSize: 12 * math.Pow(1.15, 2)},
			H6:        Style{FontSize: 12 * math.Pow(1.15, 1)},
		},
	}
}
