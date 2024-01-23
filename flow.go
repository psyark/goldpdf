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

func SplitFirstLine(pdf PDF, elements [][]FlowElement, limitWidth float64) (first []FlowElement, rest [][]FlowElement) {
	rest = elements
	width := 0.0

	for len(rest) != 0 && len(rest[0]) != 0 && width < limitWidth {
		switch e := rest[0][0].(type) {
		case *TextSpan:
			if sw := pdf.GetSpanWidth(e); sw <= limitWidth-width {
				first = append(first, e)
				width += sw
				rest[0] = rest[0][1:]
			} else {
				// 折返し
				ss := pdf.GetSubSpan(e, limitWidth-width)
				if ss.Text == "" {
					return // この行にこれ以上入らない
				}
				first = append(first, ss)
				width += pdf.GetSpanWidth(ss)
				rest[0][0] = &TextSpan{
					Format: e.Format,
					Text:   string([]rune(e.Text)[len([]rune(ss.Text)):]),
				}
			}

		case *Image:
			// 行が空の場合はlimitWidthを無視
			w, _ := e.size(pdf)
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
