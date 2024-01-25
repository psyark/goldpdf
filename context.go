package goldpdf

import (
	"bytes"
	"image/color"

	"github.com/jung-kurt/gofpdf"
)

var _ RenderContext = &renderContextImpl{}

// MeasureContext provides a way to measure the dimensions of the drawing element.
type MeasureContext interface {
	GetSpanWidth(span *TextElement) float64
	GetSubSpan(span *TextElement, width float64) *TextElement
	GetPageVerticalBounds(page int) (float64, float64)
	GetRenderContext(fn func(RenderContext) error) error
}

type RenderContext interface {
	MeasureContext
	DrawTextSpan(page int, x, y float64, span *TextElement)
	DrawImage(page int, x, y float64, img *ImageElement)
	DrawBullet(page int, x, y float64, c color.Color, r float64)
	DrawBox(rect Rect, bgColor color.Color, border Border)
}

type renderContextImpl struct {
	fpdf        *gofpdf.Fpdf
	inRendering bool
}

func (p *renderContextImpl) GetSpanWidth(span *TextElement) float64 {
	p.applyTextFormat(span.Format)
	return p.fpdf.GetStringWidth(span.Text)
}

func (p *renderContextImpl) GetSubSpan(span *TextElement, width float64) *TextElement {
	p.applyTextFormat(span.Format)
	width += span.Format.FontSize / 2 // SplitText issue

	lines := p.fpdf.SplitText(span.Text, width)
	if lines[0] == "" {
		return nil
	}

	ss := &TextElement{Text: lines[0], Format: span.Format}
	if p.GetSpanWidth(ss) > width { // SplitText issue
		return nil
	}

	return ss
}

func (p *renderContextImpl) GetPageVerticalBounds(page int) (float64, float64) {
	p.setPage(page)
	_, h := p.fpdf.GetPageSize()
	_, tm, _, bm := p.fpdf.GetMargins()
	return tm, h - bm
}

// GetRenderContext はレンダリングコンテキストの取得をリクエストします
// 取得されたレンダリングコンテキストはコールバック関数に渡されますが、
// このコールバック関数内での新たな GetRenderContextの呼び出しはスキップされます
// このルールにより、ノードの背景やボーダーを描画する際、ノード固有のレンダリング関数を再帰的に呼び出して子孫を加味した高さを計算させることができ、
// 単一の関数がノードのサイズ計算とノードの描画を担当することができるようになります。
func (rc *renderContextImpl) GetRenderContext(fn func(RenderContext) error) error {
	if !rc.inRendering {
		rc.inRendering = true
		defer func() { rc.inRendering = false }()
		return fn(RenderContext(rc))
	}
	return nil
}

func (p *renderContextImpl) DrawTextSpan(page int, x, y float64, span *TextElement) {
	p.setPage(page)
	rect := Rect{
		Left:   x,
		Right:  x + p.GetSpanWidth(span),
		Top:    VerticalCoord{Page: page, Position: y},
		Bottom: VerticalCoord{Page: page, Position: y + span.Format.FontSize},
	}
	p.DrawBox(rect, span.Format.BackgroundColor, span.Format.Border)
	p.applyTextFormat(span.Format)
	p.fpdf.Text(x, y+span.Format.FontSize, span.Text)
}

func (p *renderContextImpl) DrawImage(page int, x, y float64, img *ImageElement) {
	p.setPage(page)
	p.fpdf.RegisterImageOptionsReader(img.name, gofpdf.ImageOptions{ImageType: img.imageType}, bytes.NewReader(img.data))
	w, h := img.size(p)
	p.fpdf.ImageOptions(img.name, x, y, w, h, false, gofpdf.ImageOptions{}, 0, "")
}

func (p *renderContextImpl) DrawBullet(page int, x, y float64, c color.Color, r float64) {
	if _, _, _, ca := c.RGBA(); ca != 0 && r != 0 {
		p.setPage(page)
		p.colorHelper(c, p.fpdf.SetFillColor)
		p.fpdf.Circle(x, y, r, "F")
	}
}

func (p *renderContextImpl) DrawBox(rect Rect, bgColor color.Color, border Border) {
	x := rect.Left
	w := rect.Right - rect.Left

	for page := rect.Top.Page; page <= rect.Bottom.Page; page++ {
		y, b := p.GetPageVerticalBounds(page)
		if page == rect.Top.Page {
			y = rect.Top.Position
		}
		if page == rect.Bottom.Page {
			b = rect.Bottom.Position
		}
		h := b - y
		p.drawBoxInPage(page, x, y, w, h, bgColor, border)
	}
}

func (p *renderContextImpl) drawBoxInPage(page int, x, y, w, h float64, bgColor color.Color, border Border) {
	p.setPage(page)

	var borderRadius float64
	if border, ok := border.(UniformBorder); ok {
		borderRadius = border.Radius
	}

	if bgColor != nil {
		if _, _, _, ca := bgColor.RGBA(); ca != 0 {
			p.colorHelper(bgColor, p.fpdf.SetFillColor)
			p.fpdf.RoundedRect(x, y, w, h, borderRadius, "1234", "F")
		}
	}

	switch border := border.(type) {
	case UniformBorder:
		if border.Color != nil && border.Width != 0 {
			if _, _, _, ca := border.Color.RGBA(); ca != 0 {
				p.fpdf.SetLineWidth(border.Width)
				p.colorHelper(border.Color, p.fpdf.SetDrawColor)
				p.fpdf.RoundedRect(x+border.Width/2, y+border.Width/2, w-border.Width, h-border.Width, border.Radius, "1234", "D")
			}
		}
	case IndividualBorder:
		p.drawEdge(page, x+border.Left.Width/2, y, x+border.Left.Width/2, y+h, border.Left)
		p.drawEdge(page, x, y+border.Top.Width/2, x+w, y+border.Top.Width/2, border.Top)
		p.drawEdge(page, x+w-border.Right.Width/2, y, x+w-border.Right.Width/2, y+h, border.Right)
		p.drawEdge(page, x, y+h-border.Bottom.Width/2, x+w, y+h-border.Bottom.Width/2, border.Bottom)
	}
}

func (p *renderContextImpl) setPage(page int) {
	for page > p.fpdf.PageCount() {
		p.fpdf.AddPage()
	}
	p.fpdf.SetPage(page)
}

func (p *renderContextImpl) drawEdge(page int, x1, y1, x2, y2 float64, edge BorderEdge) {
	p.setPage(page)
	if edge.Color != nil && edge.Width != 0 {
		if _, _, _, ca := edge.Color.RGBA(); ca != 0 {
			p.fpdf.SetLineWidth(edge.Width)
			p.colorHelper(edge.Color, p.fpdf.SetDrawColor)
			p.fpdf.Line(x1, y1, x2, y2)
		}
	}
}

func (p *renderContextImpl) applyTextFormat(format TextFormat) {
	fontStyle := ""
	if format.Bold {
		fontStyle += "B"
	}
	if format.Italic {
		fontStyle += "I"
	}
	if format.Strike {
		fontStyle += "S"
	}
	if format.Underline {
		fontStyle += "U"
	}

	p.fpdf.SetFont(format.FontFamily, fontStyle, format.FontSize)
	p.colorHelper(format.Color, p.fpdf.SetTextColor)
}

func (p *renderContextImpl) colorHelper(c color.Color, fn func(int, int, int)) {
	cr, cg, cb, ca := c.RGBA()
	p.fpdf.SetAlpha(float64(ca)/0xFFFF, "")
	fn(int(cr>>8), int(cg>>8), int(cb>>8))
}
