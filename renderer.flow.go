package goldpdf

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// getFlowElements retrieves the FlowElement belonging to the specified node.
// Belonging means "a descendant inline node of the node and not a descendant of a child block node of the node."
func (r *Renderer) getFlowElements(n ast.Node) ([]InlineElement, error) {
	elements := []InlineElement{}

	switch n := n.(type) {
	case *ast.CodeBlock, *ast.FencedCodeBlock:
		tf := r.textFormat(n)
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			str := string(line.Value(r.source))

			text := &TextElement{Text: strings.TrimSuffix(str, "\n"), Format: tf}
			elements = append(elements, text)
			if strings.HasSuffix(str, "\n") {
				elements = append(elements, &LineBreakElement{Format: tf})
			}
		}
	case *ast.AutoLink:
		tf := r.textFormat(n)
		text := &TextElement{Format: tf, Text: string(n.URL(r.source))}
		elements = append(elements, text)
	case *ast.Text:
		tf := r.textFormat(n)
		text := &TextElement{Format: tf, Text: string(n.Text(r.source))}
		elements = append(elements, text)
		if n.HardLineBreak() {
			elements = append(elements, &LineBreakElement{Format: tf})
		}
	case *ast.Image:
		img, err := r.imageLoader.LoadImage(string(n.Destination))
		if err != nil {
			return nil, err
		}
		if img != nil {
			// If the image can be retrieved, ignore descendants (alt text).
			elements = append(elements, img)
			return elements, nil
		}
	case *ast.RawHTML:
		html := ""
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			html += string(seg.Value(r.source))
		}
		if html == "<br>" {
			tf := r.textFormat(n)
			elements = append(elements, &LineBreakElement{Format: tf})
		}
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Type() == ast.TypeInline {
			e, err := r.getFlowElements(c)
			if err != nil {
				return nil, err
			}
			elements = append(elements, e...)
		}
	}

	return elements, nil
}

// renderInlineElements draws inline elements inside the contentBox and returns a content box with the actual drawn height.
func (r *Renderer) renderInlineElements(elements []InlineElement, mc MeasureContext, contentBox HalfBounds, align xast.Alignment) (Rect, error) {
	result := contentBox.ToRect(contentBox.Top)

	for i, line := range wrapElements(mc, contentBox.Width(), elements) {
		lineWidth, lineHeight := getLineSize(mc, line)

		pageTop, pageBottom := mc.GetPageVerticalBounds(contentBox.Top.Page)
		if contentBox.Top.Position+lineHeight > pageBottom {
			contentBox.Top.Page++
			contentBox.Top.Position = pageTop
		}

		if i == 0 {
			result.Top = contentBox.Top
		}

		result.Bottom = contentBox.Top
		result.Bottom.Position += lineHeight

		err := mc.GetRenderContext(func(rc RenderContext) error {
			x := contentBox.Left
			y := contentBox.Top.Position

			switch align {
			case xast.AlignRight:
				x += contentBox.Width() - lineWidth
			case xast.AlignCenter:
				x += (contentBox.Width() - lineWidth) / 2
			}

			for _, e := range line {
				w, h := e.size(mc)
				e.drawTo(rc, contentBox.Top.Page, x, y+lineHeight-h)
				x += w
			}
			return nil
		})
		if err != nil {
			return Rect{}, err
		}

		contentBox.Top.Position += lineHeight
	}

	return result, nil
}
