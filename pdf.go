package goldpdf

import (
	"bytes"
	"image/color"
	"math"

	"github.com/jung-kurt/gofpdf"
)

type PDF interface {
	GetSpanWidth(span *TextSpan) float64
	GetSubSpan(span *TextSpan, width float64) *TextSpan
	GetNaturalWidth(elements []FlowElement) float64
	SplitFirstLine(elements []FlowElement, limitWidth float64) (first []FlowElement, rest []FlowElement)
	DrawTextSpan(x, y float64, span *TextSpan)
	DrawImage(x, y float64, img *imageInfo)
	DrawBullet(x, y float64, c color.Color, r float64)
	DrawLine(x1, y1, x2, y2 float64, c color.Color, w float64)
	DrawBox(x, y, w, h float64, bgColor color.Color, border Border)
}

// TODO RenderContextと統合
// Preflightを呼び出し側が判断する必要がなくなるため

type pdfImpl struct {
	fpdf *gofpdf.Fpdf
}

var _ PDF = &pdfImpl{}

func (p *pdfImpl) GetSpanWidth(span *TextSpan) float64 {
	p.applyTextFormat(span.Format)
	return p.fpdf.GetStringWidth(span.Text)
}

func (p *pdfImpl) GetSubSpan(span *TextSpan, width float64) *TextSpan {
	p.applyTextFormat(span.Format)
	lines := p.fpdf.SplitText(span.Text, width)
	return &TextSpan{Text: lines[0], Format: span.Format}
}

func (p *pdfImpl) GetNaturalWidth(elements []FlowElement) float64 {
	width := 0.0

	lineWidth := 0.0
	for _, e := range elements {
		switch e := e.(type) {
		case *TextSpan:
			lineWidth += p.GetSpanWidth(e)
		case *Image:
			lineWidth += float64(e.Info.Width)
		case *HardBreak:
			width = math.Max(width, lineWidth)
			lineWidth = 0
		}
	}

	return math.Max(width, lineWidth)
}

func (pdf *pdfImpl) SplitFirstLine(elements []FlowElement, limitWidth float64) (first []FlowElement, rest []FlowElement) {
	if len(elements) == 0 {
		return nil, nil
	}

	rest = elements
	width := 0.0

	for len(rest) != 0 && width < limitWidth {
		switch e := rest[0].(type) {
		case *TextSpan:
			if sw := pdf.GetSpanWidth(e); sw <= limitWidth-width {
				first = append(first, e)
				width += sw
				rest = rest[1:]
			} else {
				// 折返し
				ss := pdf.GetSubSpan(e, limitWidth-width)
				if ss.Text == "" {
					return // この行にこれ以上入らない
				}
				first = append(first, ss)
				width += pdf.GetSpanWidth(ss)
				rest[0] = &TextSpan{
					Format: e.Format,
					Text:   string([]rune(e.Text)[len([]rune(ss.Text)):]),
				}
			}

		case *Image:
			// 行が空の場合はlimitWidthを無視
			if len(first) == 0 || width+float64(e.Info.Width) <= limitWidth {
				first = append(first, e)
				width += float64(e.Info.Width)
				rest = rest[1:]
			} else {
				return // これ以上入らないので改行
			}
		case *HardBreak:
			rest = rest[1:]
			return
		}
	}

	return
}

func (p *pdfImpl) DrawTextSpan(x, y float64, span *TextSpan) {
	sw := p.GetSpanWidth(span)
	p.DrawBox(x, y, sw, span.Format.FontSize, span.Format.BackgroundColor, span.Format.Border)
	p.applyTextFormat(span.Format)
	p.fpdf.Text(x, y+span.Format.FontSize, span.Text)
}

func (p *pdfImpl) DrawImage(x, y float64, img *imageInfo) {
	p.fpdf.RegisterImageOptionsReader(img.Name, gofpdf.ImageOptions{ImageType: img.Type}, bytes.NewReader(img.Data))
	p.fpdf.ImageOptions(img.Name, x, y, float64(img.Width), float64(img.Height), false, gofpdf.ImageOptions{}, 0, "")
}

func (p *pdfImpl) DrawBullet(x, y float64, c color.Color, r float64) {
	if _, _, _, ca := c.RGBA(); ca != 0 && r != 0 {
		p.colorHelper(c, p.fpdf.SetFillColor)
		p.fpdf.Circle(x, y, r, "F")
	}
}

func (p *pdfImpl) DrawLine(x1, y1, x2, y2 float64, c color.Color, w float64) {
	if _, _, _, ca := c.RGBA(); ca != 0 && w != 0 {
		p.fpdf.SetLineWidth(w)
		p.colorHelper(c, p.fpdf.SetDrawColor)
		p.fpdf.Line(x1, y1, x2, y2)
	}
}

func (p *pdfImpl) DrawBox(x, y, w, h float64, bgColor color.Color, border Border) {
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
		p.drawEdge(x+border.Left.Width/2, y, x+border.Left.Width/2, y+h, border.Left)
		p.drawEdge(x, y+border.Top.Width/2, x+w, y+border.Top.Width/2, border.Top)
		p.drawEdge(x+w-border.Right.Width/2, y, x+w-border.Right.Width/2, y+h, border.Right)
		p.drawEdge(x, y+h-border.Bottom.Width/2, x+w, y+h-border.Bottom.Width/2, border.Bottom)
	}
}
func (p *pdfImpl) drawEdge(x1, y1, x2, y2 float64, edge BorderEdge) {
	if edge.Color != nil && edge.Width != 0 {
		if _, _, _, ca := edge.Color.RGBA(); ca != 0 {
			p.fpdf.SetLineWidth(edge.Width)
			p.colorHelper(edge.Color, p.fpdf.SetDrawColor)
			p.fpdf.Line(x1, y1, x2, y2)
		}
	}
}

func (p *pdfImpl) applyTextFormat(format TextFormat) {
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

func (p *pdfImpl) colorHelper(c color.Color, fn func(int, int, int)) {
	cr, cg, cb, ca := c.RGBA()
	p.fpdf.SetAlpha(float64(ca)/0xFFFF, "")
	fn(int(cr>>8), int(cg>>8), int(cb>>8))
}
