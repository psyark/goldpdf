package goldpdftest

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/psyark/goldpdf"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"gopkg.in/gographics/imagick.v3/imagick"
)

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

				markdown := goldmark.New(
					goldmark.WithExtensions(
						extension.Strikethrough,
						extension.Table,
					),
					goldmark.WithRenderer(
						goldpdf.New(
							goldpdf.WithStyler(&goldpdf.DefaultStyler{FontFamily: "Arial", FontSize: 12, Color: color.White}),
						),
					),
				)

				if err := markdown.Convert(md, buf); err != nil {
					t.Fatal(err)
				}

				mw := imagick.NewMagickWand()
				defer mw.Destroy()

				if err := mw.ReadImageBlob(buf.Bytes()); err != nil {
					t.Fatal(err)
				}
				if err := mw.SetImageFormat("png"); err != nil {
					t.Fatal(err)
				}

				gotBytes := mw.GetImageBlob()

				wantName := fmt.Sprintf("testdata/%s.png", baseName)
				diffName := fmt.Sprintf("testdata/%s_diff.png", baseName)
				gotName := fmt.Sprintf("testdata/%s_got.png", baseName)
				os.Remove(diffName)
				os.Remove(gotName)

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
					}
				} else {
					if err := os.WriteFile(wantName, gotBytes, 0666); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	}
}
