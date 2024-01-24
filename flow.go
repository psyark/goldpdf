package goldpdf

import (
	"image"
	"math"
)

// InlineElement は PDFに描画されるインラインの要素であり、テキストか画像の2種類があります
type InlineElement interface {
	size(mc MeasureContext) (float64, float64)
	drawTo(rc RenderContext, page int, x, y float64)
}

var (
	_ InlineElement = &TextElement{}
	_ InlineElement = &ImageElement{}
)

// TextElement は、単一のテキストフォーマットが設定されたテキストを持つインライン要素です
type TextElement struct {
	Format TextFormat
	Text   string
}

func (s *TextElement) size(mc MeasureContext) (float64, float64) {
	return mc.GetSpanWidth(s), s.Format.FontSize
}

func (t *TextElement) drawTo(rc RenderContext, page int, x, y float64) {
	rc.DrawTextSpan(page, x, y, t)
}

// ImageElement は、単一の画像です
type ImageElement struct {
	name      string
	imageType string
	img       image.Image
	data      []byte
}

func (i *ImageElement) size(MeasureContext) (float64, float64) {
	return float64(i.img.Bounds().Dx()), float64(i.img.Bounds().Dy())
}

func (i *ImageElement) drawTo(rc RenderContext, page int, x float64, y float64) {
	rc.DrawImage(page, x, y, i)
}

type InlineElementsLines [][]InlineElement

func (l *InlineElementsLines) AddLine(line ...InlineElement) {
	*l = append(*l, line)
}
func (l *InlineElementsLines) AppendToLastLine(e ...InlineElement) {
	if len(*l) == 0 {
		l.AddLine()
	}
	i := len(*l) - 1
	(*l)[i] = append((*l)[i], e...)
}

func (l InlineElementsLines) Width(mc MeasureContext) float64 {
	width := 0.0
	for _, line := range l {
		lineWidth := 0.0
		for _, e := range line {
			w, _ := e.size(mc)
			lineWidth += w
		}
		width = math.Max(width, lineWidth)
	}
	return width
}

func (l InlineElementsLines) Wrap(mc MeasureContext, width float64) InlineElementsLines {
	result := InlineElementsLines{}
	for _, line := range l {
		for _, wrappedLine := range wrapLine(mc, width, line) {
			result.AddLine(wrappedLine...)
		}
	}
	return result
}

func wrapLine(mc MeasureContext, limitWidth float64, line []InlineElement) InlineElementsLines {
	result := InlineElementsLines{}

	width := 0.0
	rest := append([]InlineElement{}, line...)
	for len(rest) != 0 {
		switch e := rest[0].(type) {
		case *TextElement:
			if ss := mc.GetSubSpan(e, limitWidth-width); ss.Text == "" {
				result.AddLine()
				width = 0
				continue // この行にこれ以上入らない
			} else if ss.Text == e.Text {
				// TODO 下と共通化
				result.AppendToLastLine(ss)
				width += mc.GetSpanWidth(ss)
				rest = rest[1:]
			} else {
				result.AppendToLastLine(ss)
				width += mc.GetSpanWidth(ss)
				rest[0] = &TextElement{
					Format: e.Format,
					Text:   string([]rune(e.Text)[len([]rune(ss.Text)):]),
				}
			}

		case *ImageElement:
			// 行が空の場合はlimitWidthを無視
			w, _ := e.size(mc)
			if width == 0 || width+w <= limitWidth {
				result.AppendToLastLine(e)
				width += w
				rest = rest[1:]
			} else {
				result.AddLine()
				width = 0
				continue // これ以上入らないので改行
			}
		}
	}

	return result
}
