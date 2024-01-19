package goldpdf

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// getFlowElements は指定されたインラインノードとその全ての子孫ノードをFlowElementのフラットな配列として返します
func (r *Renderer) getFlowElements(n ast.Node, tf TextFormat) ([]FlowElement, error) {
	if n.Type() != ast.TypeInline {
		return nil, fmt.Errorf("getFlowElements has been called with a non-inline node: %s", n.Kind())
	}

	_, tf = r.styler.Style(n, tf)
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
			// TODO リンク切れ
			elements = append(elements, &Image{Info: info})
		}

	case *ast.Emphasis, *ast.Link, *ast.CodeSpan, *xast.Strikethrough:
	default:
		return nil, fmt.Errorf("getFlowElements: unsupported kind: %s", n.Kind())
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		e, err := r.getFlowElements(c, tf)
		if err != nil {
			return nil, err
		}
		elements = append(elements, e...)
	}
	return elements, nil
}

// renderFlowElements はテキストフローを描画し、そのサイズを返します
func (r *Renderer) renderFlowElements(elements FlowElements, borderBox RenderContext) (float64, error) {
	height := 0.0
	for !elements.IsEmpty() {
		line, lineHeight := elements.GetLine(borderBox.Target, borderBox.W)
		if len(line) == 0 {
			break
		}

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
