package goldpdf

import (
	"os"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

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
		goldmark.WithRenderer(New()),
	)

	if err := markdown.Convert(md, out); err != nil {
		panic(err)
	}
}
