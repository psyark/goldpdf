package goldpdf

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

func TestWrapElements(t *testing.T) {
	fpdf := gofpdf.New("P", "pt", "A4", "")
	mc := &renderContextImpl{fpdf: fpdf}

	text := &TextElement{
		Format: TextFormat{FontSize: 10, FontFamily: "Arial", Color: color.Black},
		Text:   "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
	}

	result := ""
	for _, line := range wrapElements(mc, 200, []InlineElement{text}) {
		result += fmt.Sprintf("%s\n", line)
	}

	expected := `[Lorem ipsum dolor sit amet, consectetur]
[ adipiscing elit, sed do eiusmod tempor]
[ incididunt ut labore et dolore magna aliqua.]
[ Ut enim ad minim veniam, quis nostrud]
[ exercitation ullamco laboris nisi ut aliquip ex]
[ ea commodo consequat. Duis aute irure]
[ dolor in reprehenderit in voluptate velit esse]
[ cillum dolore eu fugiat nulla pariatur.]
[ Excepteur sint occaecat cupidatat non]
[ proident, sunt in culpa qui officia deserunt]
[ mollit anim id est laborum.]
`

	if result != expected {
		t.Errorf("WrapElements() = %v, want %v", result, expected)
	}
}
