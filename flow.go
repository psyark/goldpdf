package goldpdf

type FlowElement interface {
	isFlowElement()
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

func (*TextSpan) isFlowElement() {}

type HardBreak struct{}

func (*HardBreak) isFlowElement() {}

type Image struct {
	Info *imageInfo
}

func (*Image) isFlowElement() {}
