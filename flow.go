package goldpdf

import (
	"image"
)

type FlowElement interface {
	size(pdf PDF) (float64, float64)
	drawTo(x, y float64, pdf PDF) error
}

var (
	_ FlowElement = &TextSpan{}
	_ FlowElement = &Image{}
)

type TextSpan struct {
	Format TextFormat
	Text   string
}

func (s *TextSpan) size(pdf PDF) (float64, float64) {
	return pdf.GetSpanWidth(s), s.Format.FontSize
}

func (t *TextSpan) drawTo(x, y float64, pdf PDF) error {
	pdf.DrawTextSpan(x, y, t)
	return nil
}

type Image struct {
	name      string
	imageType string
	img       image.Image
	data      []byte
}

func (i *Image) size(pdf PDF) (float64, float64) {
	return float64(i.img.Bounds().Dx()), float64(i.img.Bounds().Dy())
}

func (i *Image) drawTo(x float64, y float64, pdf PDF) error {
	pdf.DrawImage(x, y, i)
	return nil
}
