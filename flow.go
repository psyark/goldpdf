package goldpdf

import "fmt"

type FlowElement interface {
	size(pdf PDF) (float64, float64)
	drawTo(x, y float64, pdf PDF) error
}

var (
	_ FlowElement = &TextSpan{}
	_ FlowElement = &HardBreak{}
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

type HardBreak struct{}

func (*HardBreak) size(pdf PDF) (float64, float64) { return 0, 0 }

func (h *HardBreak) drawTo(x, y float64, pdf PDF) error {
	return fmt.Errorf("unsupported")
}

type Image struct {
	Info *imageInfo
}

func (i *Image) size(pdf PDF) (float64, float64) {
	return float64(i.Info.Width), float64(i.Info.Height)
}

func (i *Image) drawTo(x, y float64, pdf PDF) error {
	pdf.DrawImage(x, y, i.Info)
	return nil
}
