package goldpdf

import (
	"math"
)

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

type FlowElements []FlowElement

func (f *FlowElements) IsEmpty() bool {
	return len(*f) == 0
}

func (f *FlowElements) GetLine(pdf PDF, limitWidth float64) (line FlowElements, height float64) {
	if f.IsEmpty() {
		return nil, 0
	}

	width := 0.0

	for !f.IsEmpty() && width < limitWidth {
		switch e := (*f)[0].(type) {
		case *TextSpan:
			if sw := pdf.GetSpanWidth(e); sw <= limitWidth-width {
				line = append(line, e)
				width += sw
				height = math.Max(height, e.Format.FontSize)
				*f = (*f)[1:]
			} else {
				// 折返し
				ss := pdf.GetSubSpan(e, limitWidth-width)
				if ss.Text == "" {
					return // この行にこれ以上入らない
				}
				line = append(line, ss)
				width += pdf.GetSpanWidth(ss)
				height = math.Max(height, e.Format.FontSize)
				(*f)[0] = &TextSpan{
					Format: e.Format,
					Text:   string([]rune(e.Text)[len([]rune(ss.Text)):]),
				}
			}

		case *Image:
			// 行が空の場合はlimitWidthを無視
			if line.IsEmpty() || width+float64(e.Info.Width) <= limitWidth {
				line = append(line, e)
				height = math.Max(height, float64(e.Info.Height))
				width += float64(e.Info.Width)
				*f = (*f)[1:]
			} else {
				return // これ以上入らないので改行
			}
		case *HardBreak:
			*f = (*f)[1:]
			return
		}
	}

	return
}
