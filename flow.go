package goldpdf

import (
	"math"
	"strings"
)

// InlineElement は PDFに描画されるインラインの要素であり、テキストか画像の2種類があります
type InlineElement interface {
	String() string
	size(mc MeasureContext) (float64, float64)
	drawTo(rc RenderContext, page int, x, y float64)
}

var (
	_ InlineElement = &TextElement{}
	_ InlineElement = &LineBreakElement{}
	_ InlineElement = &ImageElement{}
)

// TextElement は、単一のテキストフォーマットが設定された改行を含まないテキストを持つインライン要素です
type TextElement struct {
	Format TextFormat
	Text   string
}

func (s *TextElement) size(mc MeasureContext) (float64, float64) {
	return mc.GetTextWidth(s), s.Format.FontSize
}

func (t *TextElement) drawTo(rc RenderContext, page int, x, y float64) {
	rc.DrawText(page, x, y, t)
}

func (t *TextElement) String() string {
	return t.Text
}

// LineBreakElement は、改行を表すインライン要素です
type LineBreakElement struct {
	Format TextFormat
}

func (s *LineBreakElement) size(mc MeasureContext) (float64, float64) {
	return 0, s.Format.FontSize
}

func (t *LineBreakElement) drawTo(rc RenderContext, page int, x, y float64) {
	panic("line break")
}
func (t *LineBreakElement) String() string {
	return "\\n"
}

// ImageElement は、単一の画像です
type ImageElement struct {
	Name          string
	ImageType     string // see ImageType of fpdf.ImageOptions
	Width, Height float64
	Bytes         []byte
}

func (i *ImageElement) size(MeasureContext) (float64, float64) {
	return i.Width, i.Height
}

func (i *ImageElement) drawTo(rc RenderContext, page int, x float64, y float64) {
	rc.DrawImage(page, x, y, i)
}
func (t *ImageElement) String() string {
	return "[image]"
}

func getLineSize(mc MeasureContext, line []InlineElement) (float64, float64) {
	var width, height float64
	for _, e := range line {
		w, h := e.size(mc)
		width += w
		height = math.Max(height, h)
	}
	return width, height
}

func wrapElements(mc MeasureContext, limitWidth float64, elements []InlineElement) [][]InlineElement {
	result := [][]InlineElement{{}}
	rest := append([]InlineElement{}, elements...)

	lineWidth := 0.0
	for len(rest) != 0 {
		switch e := rest[0].(type) {
		case *TextElement:
			if e.Text == "" { // 幅0のテキストなら現在の行に追加
				result[len(result)-1] = append(result[len(result)-1], e)
				rest = rest[1:]
			} else if ss := mc.GetSubText(e, limitWidth-lineWidth); ss == nil {
				if lineWidth == 0 { // width == 0 なのにこれ以上入らないなら終了して無限ループを回避
					// TODO 1文字入れる
					return result
				}

				// この行にこれ以上入らない
				lineWidth = 0
				result = append(result, []InlineElement{})
				// Remove spaces immediately after line breaks
				e.Text = strings.TrimPrefix(e.Text, " ")
			} else {
				result[len(result)-1] = append(result[len(result)-1], ss)
				lineWidth += mc.GetTextWidth(ss)
				if ss.Text == e.Text {
					rest = rest[1:]
				} else {
					rest[0] = &TextElement{Format: e.Format, Text: strings.TrimPrefix(e.Text, ss.Text)}
				}
			}

		case *ImageElement:
			w, _ := e.size(mc)

			if lineWidth+w > limitWidth && lineWidth != 0 {
				// 幅が飛び出る場合は改行する
				// ただし現在の行が空の場合は（無限ループを防ぐため）limitWidthを無視する
				result = append(result, []InlineElement{})
				lineWidth = 0
			}

			result[len(result)-1] = append(result[len(result)-1], e)
			lineWidth += w
			rest = rest[1:]

		case *LineBreakElement:
			result = append(result, []InlineElement{})
			lineWidth = 0
			rest = rest[1:]

		default:
			panic(e)
		}
	}

	return result
}

// TODO リネーム
func GetNaturalWidth(mc MeasureContext, elements []InlineElement) float64 {
	var lineWidth, maxWidth float64
	for _, element := range elements {
		if _, ok := element.(*LineBreakElement); ok {
			if maxWidth < lineWidth {
				maxWidth = lineWidth
			}
			lineWidth = 0
		} else {
			w, _ := element.size(mc)
			lineWidth += w
		}
	}
	if maxWidth < lineWidth {
		maxWidth = lineWidth
	}
	return maxWidth
}
