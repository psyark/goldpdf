package goldpdf

import (
	"image/color"

	"github.com/go-pdf/fpdf"
)

type Style struct {
	Color      color.Color
	FontSize   float64
	FontFamily string
	Bold       bool
	Italic     bool
	Strike     bool
}

func (s Style) Apply(pdf *fpdf.Fpdf) {
	fontStyle := ""
	if s.Bold {
		fontStyle += "B"
	}
	if s.Italic {
		fontStyle += "I"
	}
	if s.Strike {
		fontStyle += "S"
	}
	pdf.SetFont(s.FontFamily, fontStyle, s.FontSize)
	cr, cg, cb, _ := s.Color.RGBA()
	pdf.SetTextColor(int(cr>>8), int(cg>>8), int(cb>>8))
}

type Styles struct {
	Paragraph              Style
	H1, H2, H3, H4, H5, H6 Style
	LinkColor              color.Color
	CodeSpan               Style
	CodeBlock              Style
}
