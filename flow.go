package goldpdf

type FlowElement interface {
	size(pdf PDF) (float64, float64)
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

type HardBreak struct{}

func (*HardBreak) size(pdf PDF) (float64, float64) { return 0, 0 }

type Image struct {
	Info *imageInfo
}

func (i *Image) size(pdf PDF) (float64, float64) {
	return float64(i.Info.Width), float64(i.Info.Height)
}
