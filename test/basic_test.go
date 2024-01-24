package goldpdftest

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/psyark/goldpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/gographics/imagick.v3/imagick"
)

var (
	//go:embed "testdata/ipaexg.ttf"
	IpaexgBytes []byte

	//go:embed "testdata/Noto_Sans/static/NotoSans-Regular.ttf"
	NotoSansRegular []byte
	//go:embed "testdata/Noto_Sans/static/NotoSans-Italic.ttf"
	NotoSansRegularItalic []byte
	//go:embed "testdata/Noto_Sans/static/NotoSans-Bold.ttf"
	NotoSansBold []byte
	//go:embed "testdata/Noto_Sans/static/NotoSans-BoldItalic.ttf"
	NotoSansBoldItalic []byte
)

func TestMain(m *testing.M) {
	imagick.Initialize()
	defer imagick.Terminate()

	m.Run()
}

func TestBasic(t *testing.T) {
	entries, err := os.ReadDir("testdata/basic")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		entry := entry

		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			baseName := strings.TrimSuffix(entry.Name(), ".md")

			t.Run(entry.Name(), func(t *testing.T) {
				md, err := os.ReadFile(fmt.Sprintf("testdata/basic/%s", entry.Name()))
				if err != nil {
					t.Fatal(err)
				}

				buf := bytes.NewBuffer(nil)

				var options []goldpdf.Option
				if strings.HasPrefix(baseName, "ja_") {
					fontFamily := "Ipaexg"
					options = []goldpdf.Option{
						goldpdf.WithStyler(&goldpdf.DefaultStyler{
							FontFamily: fontFamily,
							FontSize:   12,
							Color:      color.Black,
						}),
						goldpdf.WithPDFProvider(func() *gofpdf.Fpdf {
							f := gofpdf.New("P", "pt", "A4", "")
							f.AddUTF8FontFromBytes(fontFamily, "", IpaexgBytes)
							f.AddUTF8FontFromBytes(fontFamily, "B", IpaexgBytes)
							f.AddUTF8FontFromBytes(fontFamily, "I", IpaexgBytes)
							f.AddUTF8FontFromBytes(fontFamily, "BI", IpaexgBytes)
							return f
						}),
					}
				} else {
					fontFamily := "NotoSans"
					options = []goldpdf.Option{
						goldpdf.WithStyler(&goldpdf.DefaultStyler{
							FontFamily: fontFamily,
							FontSize:   12,
							Color:      color.Black,
						}),
						goldpdf.WithPDFProvider(func() *gofpdf.Fpdf {
							f := gofpdf.New("P", "pt", "A4", "")
							f.AddUTF8FontFromBytes(fontFamily, "", NotoSansRegular)
							f.AddUTF8FontFromBytes(fontFamily, "B", NotoSansBold)
							f.AddUTF8FontFromBytes(fontFamily, "I", NotoSansRegularItalic)
							f.AddUTF8FontFromBytes(fontFamily, "BI", NotoSansBoldItalic)
							return f
						}),
					}

				}
				markdown := goldmark.New(
					goldmark.WithExtensions(
						extension.Strikethrough,
						extension.Table,
					),
					goldmark.WithRenderer(
						goldpdf.New(options...),
					),
				)

				if err := markdown.Convert(md, buf); err != nil {
					t.Fatal(err)
				}

				pdfBytes := buf.Bytes()
				pdfName := fmt.Sprintf("testdata/basic/%s.pdf", baseName)
				wantName := fmt.Sprintf("testdata/basic/%s.png", baseName)
				gotName := fmt.Sprintf("testdata/basic/%s_got.png", baseName)
				diffName := fmt.Sprintf("testdata/basic/%s_diff.png", baseName)

				if err := CompareAndOutputResults(pdfBytes, pdfName, wantName, gotName, diffName); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}