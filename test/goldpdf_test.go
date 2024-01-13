package goldpdftest

import (
	"fmt"
	"image/color"
	"log"
	"os"
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
	md, err := os.ReadFile("example.md")
	if err != nil {
		panic(err)
	}

	out, err := os.OpenFile("example.pdf", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	defer out.Close()

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

	if err := markdown.Convert(md, out); err != nil {
		panic(err)
	}

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImage("example.pdf"); err != nil {
		t.Fatal(err)
	}
	if err := mw.SetImageFormat("png"); err != nil {
		t.Fatal(err)
	}
	if err := mw.SetResolution(150, 150); err != nil {
		log.Fatal("failed at SetResolution", err)
	}

	n := mw.GetNumberImages()
	for i := 0; i < int(n); i++ {
		if !mw.SetIteratorIndex(i) {
			break
		}

		pw := imagick.NewPixelWand()
		pw.SetColor("white")
		mw.SetImageBackgroundColor(pw)
		mw.SetBackgroundColor(pw)

		if err := mw.WriteImage(fmt.Sprintf("example_%02d.png", i)); err != nil {
			t.Fatal(err)
		}
	}
}
