module github.com/psyark/goldpdf/test

go 1.19

replace github.com/psyark/goldpdf => ../

require (
	github.com/psyark/goldpdf v0.0.0-00010101000000-000000000000
	github.com/yuin/goldmark v1.6.0
	gopkg.in/gographics/imagick.v3 v3.5.1
)

require github.com/go-pdf/fpdf v0.9.0 // indirect
