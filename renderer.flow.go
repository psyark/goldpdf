package goldpdf

import (
	"math"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// getFlowElements は指定されたノードに所属するFlowElementのスライスを取得します
// 指定されたノードがブロックノードである場合、その直下のインラインノード（およびその子孫）を探索します
// 指定されたノードがインラインノードである場合、その子孫が対象となります
func (r *Renderer) getFlowElements(n ast.Node) []FlowElement {
	elements := []FlowElement{}

	switch n := n.(type) {
	case *ast.CodeBlock, *ast.FencedCodeBlock:
		_, tf := r.style(n)
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			line := lines.At(i)
			ts := &TextSpan{Text: string(line.Value(r.source)), Format: tf}
			elements = append(elements, ts, &HardBreak{})
		}
	case *ast.AutoLink:
		_, tf := r.style(n)
		ts := &TextSpan{Format: tf, Text: string(n.URL(r.source))}
		elements = append(elements, ts)
	case *ast.Text:
		_, tf := r.style(n)
		ts := &TextSpan{Format: tf, Text: string(n.Text(r.source))}
		elements = append(elements, ts)
		if n.HardLineBreak() {
			elements = append(elements, &HardBreak{})
		}
	case *ast.Image:
		img := r.imageLoader.LoadImage(string(n.Destination))
		if img != nil {
			// If the image can be retrieved, ignore descendants (alt text).
			elements = append(elements, img)
			return elements
		} else {
			e := r.getFlowElements(n)
			elements = append(elements, e...)
		}
	}

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Type() == ast.TypeInline {
			e := r.getFlowElements(c)
			elements = append(elements, e...)
		}
	}

	return elements
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

		err := borderBox.Preflight(func() error {
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
					return err
				}
				x += w
			}
			return nil
		})
		if err != nil {
			return 0, err
		}

		height += lineHeight
	}
	return height, nil
}
