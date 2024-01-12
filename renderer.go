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

	case *xast.Table:
		return r.renderTable(n, entering)
	case *xast.TableHeader:
		return r.renderTableHeader(n, entering)
	case *xast.TableRow:
		return r.renderTableRow(n, entering)
	case *xast.TableCell:
		return r.renderTableCell(n, entering)

	case *ast.Document, *ast.CodeSpan, *ast.Emphasis, *xast.Strikethrough:
		return ast.WalkContinue, nil // do nothing

	default:
		if entering {
			debugStyle := Style{
				FontSize: 9,
				Color:    color.RGBA{R: 0xFF, A: 0xFF},
			}

			r.pdf.Ln(0)
			r.drawText(
				fmt.Sprintf("%v not implemented", n.Kind().String()),
				"",
				debugStyle,
			)
			r.pdf.Ln(10)
		}
		return ast.WalkContinue, nil
	}
}

// drawText はインラインの背景やボーダーを含むスタイル、折返し、ページ跨ぎを考慮して
// テキストを描画します。
// 内部でfpdf.CellFormatへの複数回の呼び出しを行います
// 全てのテキストの描画はこの関数を通して行ってください
func (r *Renderer) drawText(text string, link string, style Style) {
	style.Apply(r.pdf)
	r.pdf.WriteLinkString(style.FontSize, text, link)
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
