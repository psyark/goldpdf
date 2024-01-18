package goldpdf

import (
	"bytes"
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
	pdf         *fpdf.Fpdf
	styler      Styler
	imageLoader imageLoader
}

// drawTextFlow はテキストフローを描画し、そのサイズを返します
// drawが偽のとき、描画は行わずにサイズだけを返します
func (r *Renderer) drawTextFlow(elements FlowElements, draw bool, rs RenderState) (float64, error) {

	height := 0.0
	y := rs.Y
	for i := 0; i < 100 && !elements.IsEmpty(); i++ {
		line, lineHeight := elements.GetLine(r.pdf, rs.W)
		x := rs.X
		if draw {
			for _, e := range line {
				switch e := e.(type) {
				case *TextSpan:
					e.Format.Apply(r.pdf) //TODO 整理
					sw := r.pdf.GetStringWidth(e.Text)
					if e.Format.BackgroundColor != nil {
						cr, cg, cb, ca := e.Format.BackgroundColor.RGBA()
						if ca != 0 {
							r.pdf.SetAlpha(float64(ca)/0xFFFF, "")
							r.pdf.SetFillColor(int(cr>>8), int(cg>>8), int(cb>>8))
							r.pdf.RoundedRect(x, y, sw, e.Format.FontSize, e.Format.Border.Radius, "1234", "F")
						}
					}
					if e.Format.Border.Color != nil && e.Format.Border.Width != 0 {
						cr, cg, cb, ca := e.Format.Border.Color.RGBA()
						if ca != 0 {
							r.pdf.SetLineWidth(e.Format.Border.Width)
							r.pdf.SetAlpha(float64(ca)/0xFFFF, "")
							r.pdf.SetDrawColor(int(cr>>8), int(cg>>8), int(cb>>8))
							r.pdf.RoundedRect(x, y, sw, e.Format.FontSize, e.Format.Border.Radius, "1234", "D")
						}
					}
					e.Format.Apply(r.pdf) // TODO 整理
					r.pdf.Text(x, y+e.Format.FontSize, e.Text)
					x += sw
				case *Image:
					r.pdf.RegisterImageOptionsReader(
						e.Info.Name,
						fpdf.ImageOptions{ImageType: e.Info.Type},
						bytes.NewReader(e.Info.Data),
					)

					r.pdf.ImageOptions(e.Info.Name, x, y, float64(e.Info.Width), float64(e.Info.Height), false, fpdf.ImageOptions{}, 0, "")
					x += float64(e.Info.Width)

				default:
					return 0, fmt.Errorf("unsupported element: %v", e)
				}
			}
		}
		height += lineHeight
		y += lineHeight
	}

	return height, nil
}

// drawBlockNode はブロックノード（またはドキュメントノード）を描画し、そのサイズを返します
// drawが偽のとき、描画は行わずにサイズだけを返します
func (r *Renderer) drawBlockNode(n ast.Node, draw bool, rs RenderState) (float64, error) {
	if n.Type() == ast.TypeInline {
		return 0, fmt.Errorf("drawBlockNode called with inline node: %v > %v", n.Parent().Kind(), n.Kind())
	}

	switch n := n.(type) {
	case *ast.Paragraph, *ast.TextBlock, *ast.Heading, *xast.TableCell: // 内部にインラインがくるやつ
		return r.drawInlineContainer(n, draw, rs)

	case *ast.ThematicBreak:
		return r.drawThematicBreak(n, draw, rs)
	case *ast.Blockquote:
		return r.drawBlockQuote(n, draw, rs)

	default: // 内部にブロックがくるやつ
		return r.drawDefaultBlock(n, draw, rs)
	}
}

func (r *Renderer) drawInlineContainer(n ast.Node, draw bool, rs RenderState) (float64, error) {
	// TODO: default style
	tf := TextFormat{
		Color:      color.Black,
		FontSize:   12,
		FontFamily: "Arial",
	}

	switch n := n.(type) {
	case *ast.Paragraph, *ast.Heading, *ast.TextBlock, *xast.TableCell:
	default:
		return 0, fmt.Errorf("unsupported kind: %v", n.Kind().String())
	}

	bs, tf := r.styler.Style(n, tf)

	rs2 := rs
	rs2.Y += bs.Margin.Top
	elements := r.getFlowElements(n, tf)
	// if draw {
	// 	height, _ := r.drawTextFlow(elements, false, rs2)
	// 	// ここで背景を描画
	// 	r.pdf.SetAlpha(1, "")
	// 	r.pdf.SetLineWidth(0.5)
	// 	r.pdf.SetDrawColor(0x00, 0x80, 0x80)
	// 	r.pdf.SetFillColor(0xEE, 0xFF, 0xFF)
	// 	r.pdf.SetTextColor(0x00, 0x66, 0x66)
	// 	r.pdf.Rect(rs.X, rs.Y, rs.W, height+bs.Margin.Top+bs.Margin.Bottom, "DF")

	// 	// debug
	// 	r.pdf.SetFont("Arial", "B", 8)
	// 	t := fmt.Sprintf("[%s]", n.Kind().String())
	// 	r.pdf.Text(rs.X+rs.W-r.pdf.GetStringWidth(t)-2, rs.Y+10, t)
	// }
	height, err := r.drawTextFlow(elements, draw, rs2)
	return height + bs.Margin.Top + bs.Margin.Bottom, err
}

func (r *Renderer) drawDefaultBlock(n ast.Node, draw bool, rs RenderState) (float64, error) {
	var height float64
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if h, err := r.drawBlockNode(c, draw, rs); err != nil {
			return 0, err
		} else {
			height += h
			rs.Y += h
		}
	}
	return height, nil
}

func (r *Renderer) drawBlockQuote(n *ast.Blockquote, draw bool, rs RenderState) (float64, error) {
	rs2 := rs
	rs2.X += 10
	rs2.W -= 10

	if draw {
		h, err := r.drawDefaultBlock(n, false, rs2)
		if err != nil {
			return 0, err
		}

		r.pdf.SetAlpha(1, "")
		r.pdf.SetLineWidth(4)
		r.pdf.SetDrawColor(0, 0, 0)
		r.pdf.Line(rs.X+2, rs.Y, rs.X+2, rs.Y+h)
	}

	return r.drawDefaultBlock(n, draw, rs2)
}

func (r *Renderer) drawThematicBreak(n *ast.ThematicBreak, draw bool, rs RenderState) (float64, error) {
	if draw {
		r.pdf.SetAlpha(1, "")
		r.pdf.SetDrawColor(0x80, 0x80, 0x80)
		r.pdf.SetLineWidth(2)
		r.pdf.Line(rs.X, rs.Y+20, rs.X+rs.W, rs.Y+20)
	}
	return 40, nil
}

func (r *Renderer) getFlowElements(n ast.Node, tf TextFormat) []FlowElement {
	elements := []FlowElement{}

	_, tf = r.styler.Style(n, tf)

	if n.Type() == ast.TypeInline {
		switch n := n.(type) {
		case *ast.Text:
			elements = append(elements, &TextSpan{Format: tf, Text: string(n.Text(r.source))})
			if n.HardLineBreak() {
				elements = append(elements, &HardBreak{})
			}
		case *ast.Emphasis, *ast.Link, *ast.CodeSpan, *xast.Strikethrough:
		case *ast.Image:
			info := r.imageLoader.load(string(n.Destination))
			if info != nil {
				// TODO リンク切れ
				elements = append(elements, &Image{Info: info})
			}
		case *ast.AutoLink:
			ts := &TextSpan{
				Text:   string(n.URL(r.source)),
				Format: tf,
			}
			elements = append(elements, ts)
		default:
			fmt.Println("🍣", n.Kind())
		}
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		e := r.getFlowElements(c, tf)
		elements = append(elements, e...)
	}
	return elements
}

func (r *Renderer) Render(w io.Writer, source []byte, n ast.Node) error {
	r.source = source
	if n.Type() == ast.TypeDocument {
		r.pdf = r.pdfProvider()
		r.pdf.AddPage()

		lm, tm, rm, _ := r.pdf.GetMargins()
		w, _ := r.pdf.GetPageSize()
		rs := RenderState{X: lm, Y: tm, W: w - lm - rm}

		if _, err := r.drawBlockNode(n, true, rs); err != nil {
			return err
		}
	}
	return r.pdf.Output(w)
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
