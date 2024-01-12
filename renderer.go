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
	source      []byte
	pdf         *fpdf.Fpdf
	states      []*State
	styler      Styler
	imageLoader imageLoader
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
		// depth := len(r.states)
		// fmt.Println(strings.Repeat("  ", depth), n.Kind(), n.Type())
		newState := &State{
			Node:  n,
			Style: Style{},
		}
		if n.Type() == ast.TypeDocument {
			newState.Style = r.styler.Style(newState.Style, n)
		} else {
			newState.Style = r.styler.Style(r.currentState().Style, n)
			newState.Link = r.currentState().Link
		}
		r.states = append(r.states, newState)
	} else {
		defer func() {
			r.states = r.states[:len(r.states)-1]
		}()
	}

	switch n := n.(type) {
	case *ast.Document:
		return r.renderDocument(n, entering)
	case *ast.Heading:
		return r.renderHeading(n, entering)
	case *ast.FencedCodeBlock:
		return r.renderFencedCodeBlock(n, entering)
	case *ast.Paragraph:
		return r.renderParagraph(n, entering)
	case *ast.Text:
		return r.renderText(n, entering)
	case *ast.ThematicBreak:
		return r.renderThematicBreak(n, entering)
	case *ast.Blockquote:
		return r.renderBlockquote(n, entering)
	case *ast.List:
		return r.renderList(n, entering)
	case *ast.ListItem:
		return r.renderListItem(n, entering)
	case *ast.TextBlock:
		return r.renderTextBlock(n, entering)
	case *ast.Link:
		return r.renderLink(n, entering)
	case *ast.AutoLink:
		return r.renderAutoLink(n, entering)
	case *ast.Image:
		return r.renderImage(n, entering)

	case *ast.CodeSpan, *ast.Emphasis, *xast.Strikethrough:
		return ast.WalkContinue, nil // do nothing

	default:
		if entering {
			r.pdf.Ln(10)
			r.pdf.SetFont("", "", 10)
			r.pdf.SetTextColor(255, 0, 0)
			r.pdf.Write(10, fmt.Sprintf("%v not implemented", n.Kind().String()))
			r.pdf.SetTextColor(0, 0, 0)
			r.pdf.Ln(10)
		}
		return ast.WalkContinue, nil
	}
}

func (r *Renderer) renderDocument(n *ast.Document, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}

func (r *Renderer) currentState() *State {
	return r.states[len(r.states)-1]
}

// AddOptions does nothing
func (r *Renderer) AddOptions(options ...renderer.Option) {
}

func New() renderer.Renderer {
	pdf := fpdf.New(fpdf.OrientationPortrait, "pt", "A4", ".")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 16)

	return &Renderer{
		pdf:    pdf,
		states: []*State{},
		styler: &DefaultStyler{FontFamily: "", FontSize: 12, Color: color.Black},
	}
}
