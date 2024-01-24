package goldpdf

import (
	"math"
	"strings"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// getFlowElements retrieves the FlowElement belonging to the specified node.
// Belonging means "a descendant inline node of the node and not a descendant of a child block node of the node."
// The result is a slice of a slice of a FlowElement, where the outer slice represents a break by a HardLineBreak.
func (r *Renderer) getFlowElements(n ast.Node) [][]FlowElement {
	elements := [][]FlowElement{}

	addElementsToLastLine := func(e ...FlowElement) {
		if len(elements) == 0 {
			elements = append(elements, []FlowElement{})
		}
		elements[len(elements)-1] = append(elements[len(elements)-1], e...)
	}

	switch n := n.(type) {
	case *ast.CodeBlock, *ast.FencedCodeBlock:
		tf := r.textFormat(n)
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			ts := &TextSpan{Text: strings.TrimRight(string(line.Value(r.source)), "\n"), Format: tf}
			addElementsToLastLine(ts)
			elements = append(elements, []FlowElement{}) // HardLineBreak
		}
	case *ast.AutoLink:
		tf := r.textFormat(n)
		ts := &TextSpan{Format: tf, Text: string(n.URL(r.source))}
		addElementsToLastLine(ts)
	case *ast.Text:
		tf := r.textFormat(n)
		ts := &TextSpan{Format: tf, Text: string(n.Text(r.source))}
		addElementsToLastLine(ts)
		if n.HardLineBreak() {
			elements = append(elements, []FlowElement{}) // HardLineBreak
		}
	case *ast.Image:
		img := r.imageLoader.LoadImage(string(n.Destination))
		if img != nil {
			// If the image can be retrieved, ignore descendants (alt text).
			addElementsToLastLine(img)
			return elements
		} else {
			e := r.getFlowElements(n)
			addElementsToLastLine(e[0]...)
			elements = append(elements, e[1:]...)
		}
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Type() == ast.TypeInline {
			e := r.getFlowElements(c)
			addElementsToLastLine(e[0]...)
			elements = append(elements, e[1:]...)
		}
	}

	return elements
}

// renderFlowElements draws a text flow inside the contentBox and returns a content box with the actual drawn height.
func (r *Renderer) renderFlowElements(elements [][]FlowElement, mc MeasureContext, contentBox Rect, align xast.Alignment) (Rect, error) {
	height := 0.0
	for len(elements) != 0 {
		line, rest := splitFirstLine(elements, mc, contentBox.Width())
		if len(line) == 0 {
			break
		}

		var lineWidth, lineHeight float64
		for _, e := range line {
			w, h := e.size(mc)
			lineWidth += w
			lineHeight = math.Max(lineHeight, h)
		}

		elements = rest

		err := mc.GetRenderContext(func(rc RenderContext) error {
			x := contentBox.Left
			y := contentBox.Top.Position + height

			switch align {
			case xast.AlignRight:
				x += contentBox.Width() - lineWidth
			case xast.AlignCenter:
				x += (contentBox.Width() - lineWidth) / 2
			}

			for _, e := range line {
				w, h := e.size(mc)
				e.drawTo(contentBox.Top.Page, x, y+lineHeight-h, rc)
				x += w
			}
			return nil
		})
		if err != nil {
			return Rect{}, err
		}

		height += lineHeight
	}

	contentBox.Bottom = VerticalCoord{
		Page:     contentBox.Top.Page,
		Position: contentBox.Top.Position + height,
	}
	return contentBox, nil
}
