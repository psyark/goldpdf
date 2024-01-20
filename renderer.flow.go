package goldpdf

import (
	"fmt"
	"math"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// getFlowElements は指定されたインラインノードとその全ての子孫ノードをFlowElementのフラットな配列として返します
func (r *Renderer) getFlowElements(n ast.Node) ([]FlowElement, error) {
	if n.Type() != ast.TypeInline {
		return nil, fmt.Errorf("getFlowElements has been called with a non-inline node: %s", n.Kind())
	}

	_, tf := r.style(n)
	elements := []FlowElement{}

	switch n := n.(type) {
	case *ast.AutoLink:
		ts := &TextSpan{Format: tf, Text: string(n.URL(r.source))}
		elements = append(elements, ts)
	case *ast.Text:
		ts := &TextSpan{Format: tf, Text: string(n.Text(r.source))}
		elements = append(elements, ts)
		if n.HardLineBreak() {
			elements = append(elements, &HardBreak{})
		}
	case *ast.Image:
		info := r.imageLoader.load(string(n.Destination))
		if info != nil {
			// If the image can be retrieved, ignore descendants (alt text).
			elements = append(elements, &Image{Info: info})
			return elements, nil
		}

	case *ast.Emphasis, *ast.Link, *ast.CodeSpan, *xast.Strikethrough:
	default:
		return nil, fmt.Errorf("getFlowElements: unsupported kind: %s", n.Kind())
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		e, err := r.getFlowElements(c)
		if err != nil {
			return nil, err
		}
		elements = append(elements, e...)
	}
	return elements, nil
}

// renderFlowElements はテキストフローを描画し、その高さを返します
func (r *Renderer) renderFlowElements(elements []FlowElement, borderBox RenderContext, align xast.Alignment) (float64, error) {
	height := 0.0
	for len(elements) != 0 {
		line, rest := borderBox.Target.SplitFirstLine(elements, borderBox.W)
		if len(line) == 0 {
			break
		}

		var lineWidth, lineHeight float64
		for _, e := range line {
			w, h := e.size(borderBox.Target)
			lineWidth += w
			lineHeight = math.Max(lineHeight, h)
		}

		elements = rest

		if !borderBox.Preflight {
			x := borderBox.X
			y := borderBox.Y + height

			switch align {
			case xast.AlignRight:
				x += borderBox.W - lineWidth
			case xast.AlignCenter:
				x += (borderBox.W - lineWidth) / 2
			}

			for _, e := range line {
				w, h := e.size(borderBox.Target)
				if err := e.drawTo(x, y+lineHeight-h, borderBox.Target); err != nil {
					return 0, err
				}
				x += w
			}
		}

		height += lineHeight
	}
	return height, nil
}
