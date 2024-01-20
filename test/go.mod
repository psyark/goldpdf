module github.com/psyark/goldpdf/test

go 1.19

replace github.com/psyark/goldpdf => ../

require (
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/psyark/goldpdf v0.0.0-00010101000000-000000000000
	github.com/yuin/goldmark v1.6.0
	gopkg.in/gographics/imagick.v3 v3.5.1
)

require (
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/raykov/css-font-parser v0.3.0 // indirect
	github.com/raykov/oksvg v0.0.5 // indirect
	github.com/srwiley/rasterx v0.0.0-20220128185129-2efea2b9ea41 // indirect
	golang.org/x/image v0.0.0-20220321031419-a8550c1d254a // indirect
	golang.org/x/net v0.0.0-20220325170049-de3da57026de // indirect
	golang.org/x/text v0.3.7 // indirect
)
