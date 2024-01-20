package goldpdf

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// getFlowElements は指定されたインラインノードとその全ての子孫ノードをFlowElementのフラットな配列として返します
func (r *Renderer) getFlowElements(n ast.Node) ([]FlowElement, error) {
	if n.Type() != ast.TypeInline {
		return nil, fmt.Errorf("getFlowElements has been called with a non-inline node: %s", n.Kind())
	}

	_, tf := r.styler.Style(n)
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
func (r *Renderer) renderFlowElements(elements []FlowElement, borderBox RenderContext) (float64, error) {
	height := 0.0
	for len(elements) != 0 {
		line, rest, lineHeight := borderBox.Target.SplitFirstLine(elements, borderBox.W)
		if len(line) == 0 {
			break
		}

		elements = rest

		if !borderBox.Preflight {
			x := borderBox.X
			y := borderBox.Y + height
			for _, e := range line {
				// TODO ベースラインで揃える
				switch e := e.(type) {
				case *TextSpan:
					borderBox.Target.DrawTextSpan(x, y, e)
					x += borderBox.Target.GetSpanWidth(e)
				case *Image:
					borderBox.Target.DrawImage(x, y, e.Info)
					x += float64(e.Info.Width)
				default:
					return 0, fmt.Errorf("unsupported element: %v", e)
				}
			}
		}

		height += lineHeight
	}
	return height, nil
}
