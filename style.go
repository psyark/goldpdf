package goldpdf

type Style struct {
	FontSize float64
}

type Styles struct {
	Paragraph              Style
	H1, H2, H3, H4, H5, H6 Style
}
