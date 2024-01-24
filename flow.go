package goldpdf

import (
	"image"
	"math"
)

type FlowElement interface {
	size(mc MeasureContext) (float64, float64)
	drawTo(page int, x, y float64, rc RenderContext)
}

var (
	_ FlowElement = &TextSpan{}
	_ FlowElement = &Image{}
)

type TextSpan struct {
	Format TextFormat
	Text   string
}

func (s *TextSpan) size(mc MeasureContext) (float64, float64) {
	return mc.GetSpanWidth(s), s.Format.FontSize
}

func (t *TextSpan) drawTo(page int, x, y float64, rc RenderContext) {
	rc.DrawTextSpan(page, x, y, t)
}

type Image struct {
	name      string
	imageType string
	img       image.Image
	data      []byte
}

func (i *Image) size(MeasureContext) (float64, float64) {
	return float64(i.img.Bounds().Dx()), float64(i.img.Bounds().Dy())
}

func (i *Image) drawTo(page int, x float64, y float64, rc RenderContext) {
	rc.DrawImage(page, x, y, i)
}

func getNaturalWidth(elements [][]FlowElement, mc MeasureContext) float64 {
	width := 0.0
	for _, line := range elements {
		lineWidth := 0.0
		for _, e := range line {
			w, _ := e.size(mc)
			lineWidth += w
		}
		width = math.Max(width, lineWidth)
	}
	return width
}

// TODO 二番目の返り値必要？
func splitFirstLine(elements [][]FlowElement, mc MeasureContext, limitWidth float64) (first []FlowElement, rest [][]FlowElement) {
	rest = elements
	width := 0.0

	for len(rest) != 0 && len(rest[0]) != 0 && width < limitWidth {
		switch e := rest[0][0].(type) {
		case *TextSpan:
			if ss := mc.GetSubSpan(e, limitWidth-width); ss.Text == "" {
				return // この行にこれ以上入らない
			} else if ss.Text == e.Text {
				first = append(first, e)
				width += mc.GetSpanWidth(e)
				rest[0] = rest[0][1:]
			} else {
				first = append(first, ss)
				width += mc.GetSpanWidth(ss)
				rest[0][0] = &TextSpan{
					Format: e.Format,
					Text:   string([]rune(e.Text)[len([]rune(ss.Text)):]),
				}
			}

		case *Image:
			// 行が空の場合はlimitWidthを無視
			w, _ := e.size(mc)
			if len(first) == 0 || width+w <= limitWidth {
				first = append(first, e)
				width += w
				rest[0] = rest[0][1:]
			} else {
				return // これ以上入らないので改行
			}
		}
	}

	if len(rest[0]) == 0 {
		rest = rest[1:]
		return
	}

	return
}
