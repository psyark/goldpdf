package goldpdf

import (
	"bytes"
	"image/color"
	"math"

	"github.com/go-pdf/fpdf"
)

type PDF interface {
	GetSpanWidth(span *TextSpan) float64
	GetSubSpan(span *TextSpan, width float64) *TextSpan
	GetNaturalWidth(elements FlowElements) float64
	DrawTextSpan(x, y float64, span *TextSpan)
	DrawImage(x, y float64, img *imageInfo)
	DrawBullet(x, y float64, c color.Color, r float64)
	DrawLine(x1, y1, x2, y2 float64, c color.Color, w float64)
	DrawRect(x, y, w, h float64, bgColor color.Color, border UniformBorder)
}

type pdfImpl struct {
	fpdf *fpdf.Fpdf
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

func (p *pdfImpl) GetNaturalWidth(elements FlowElements) float64 {
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

func (p *pdfImpl) DrawTextSpan(x, y float64, span *TextSpan) {
	sw := p.GetSpanWidth(span)
	p.DrawRect(x, y, sw, span.Format.FontSize, span.Format.BackgroundColor, span.Format.Border)
	p.applyTextFormat(span.Format)
	p.fpdf.Text(x, y+span.Format.FontSize, span.Text)
}

func (p *pdfImpl) DrawImage(x, y float64, img *imageInfo) {
	p.fpdf.RegisterImageOptionsReader(img.Name, fpdf.ImageOptions{ImageType: img.Type}, bytes.NewReader(img.Data))
	p.fpdf.ImageOptions(img.Name, x, y, float64(img.Width), float64(img.Height), false, fpdf.ImageOptions{}, 0, "")
}

func (p *pdfImpl) DrawBullet(x, y float64, c color.Color, r float64) {
	if _, _, _, ca := c.RGBA(); ca != 0 && r != 0 {
		p.fpdf.SetFillColor(p.colorHelper(c))
		p.fpdf.Circle(x, y, r, "F")
	}
}

func (p *pdfImpl) DrawLine(x1, y1, x2, y2 float64, c color.Color, w float64) {
	if _, _, _, ca := c.RGBA(); ca != 0 && w != 0 {
		p.fpdf.SetLineWidth(w)
		p.fpdf.SetDrawColor(p.colorHelper(c))
		p.fpdf.Line(x1, y1, x2, y2)
	}
}

func (p *pdfImpl) DrawRect(x, y, w, h float64, bgColor color.Color, border UniformBorder) {
	if bgColor != nil {
		if _, _, _, ca := bgColor.RGBA(); ca != 0 {
			p.fpdf.SetFillColor(p.colorHelper(bgColor))
			p.fpdf.RoundedRect(x, y, w, h, border.Radius, "1234", "F")
		}
	}
	if border.Color != nil && border.Width != 0 {
		if _, _, _, ca := border.Color.RGBA(); ca != 0 {
			p.fpdf.SetLineWidth(border.Width)
			p.fpdf.SetDrawColor(p.colorHelper(border.Color))
			p.fpdf.RoundedRect(x+border.Width/2, y+border.Width/2, w-border.Width, h-border.Width, border.Radius, "1234", "D")
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
	p.fpdf.SetTextColor(p.colorHelper(format.Color))
}

func (p *pdfImpl) colorHelper(c color.Color) (int, int, int) {
	cr, cg, cb, ca := c.RGBA()
	p.fpdf.SetAlpha(float64(ca)/0xFFFFF, "")
	return int(cr >> 8), int(cg >> 8), int(cb >> 8)
}
