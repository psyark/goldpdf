package goldpdftest

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/jung-kurt/gofpdf"
	"github.com/psyark/goldpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/gographics/imagick.v3/imagick"
)

//go:embed "testdata/ipaexg.ttf"
var IpaexgBytes []byte

func TestMain(m *testing.M) {
	imagick.Initialize()
	defer imagick.Terminate()

	m.Run()
}

func TestConvert(t *testing.T) {
	entries, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		entry := entry
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			baseName := strings.TrimSuffix(entry.Name(), ".md")

			t.Run(entry.Name(), func(t *testing.T) {
				md, err := os.ReadFile(fmt.Sprintf("testdata/%s", entry.Name()))
				if err != nil {
					t.Fatal(err)
				}

				buf := bytes.NewBuffer(nil)

				options := []goldpdf.Option{}
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

				pdfName := fmt.Sprintf("testdata/%s.pdf", baseName)
				pdfBytes := buf.Bytes()
				os.WriteFile(pdfName, pdfBytes, 0666)

				got, err := capturePDF(pdfBytes, color.White)
				if err != nil {
					t.Fatal(err)
				}

				buf.Reset()
				if err := png.Encode(buf, got); err != nil {
					t.Fatal(err)
				}

				gotBytes := buf.Bytes()

				wantName := fmt.Sprintf("testdata/%s.png", baseName)
				diffName := fmt.Sprintf("testdata/%s_diff.png", baseName)
				gotName := fmt.Sprintf("testdata/%s_got.png", baseName)

				if wantBytes, err := os.ReadFile(wantName); !os.IsNotExist(err) {
					diff, err := CompareImages(wantBytes, gotBytes)
					if err != nil {
						t.Fatal(err)
					}
					if diff != nil {
						f, err := os.OpenFile(diffName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
						if err != nil {
							t.Fatal(err)
						}

						defer f.Close()
						if err := png.Encode(f, diff); err != nil {
							t.Fatal(err)
						}

						if err := os.WriteFile(gotName, gotBytes, 0666); err != nil {
							t.Fatal(err)
						}
						t.Fatal("mismatch")
					} else {
						os.Remove(diffName)
						os.Remove(gotName)
					}
				} else {
					os.Remove(diffName)
					os.Remove(gotName)
					if err := os.WriteFile(wantName, gotBytes, 0666); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}

func capturePDF(pdfBytes []byte, bgColor color.Color) (image.Image, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(pdfBytes); err != nil {
		return nil, err
	}
	if err := mw.SetImageFormat("png"); err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(mw.GetImageBlob()))
	if err != nil {
		return nil, err
	}

	bg := image.NewRGBA(img.Bounds())
	draw.Draw(bg, bg.Rect, image.NewUniform(bgColor), image.Point{}, draw.Over)
	draw.Draw(bg, bg.Rect, img, image.Point{}, draw.Over)

	return bg, nil
}
