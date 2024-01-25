package goldpdf

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

// renderBlockNode draws a block node (or document node) inside a borderBox
// and returns a border box with the actual drawn height.
func (r *Renderer) renderBlockNode(n ast.Node, mc MeasureContext, borderBox HalfBounds) (Rect, error) {
	if n.Type() == ast.TypeInline {
		return Rect{}, fmt.Errorf("renderBlockNode has been called with an inline node: %v > %v", n.Parent().Kind(), n.Kind())
	}

	switch n := n.(type) {
	case *ast.ListItem:
		return r.renderListItem(n, mc, borderBox)
	case *xast.Table:
		return r.renderTable(n, mc, borderBox)
	default:
		return r.renderGenericBlockNode(n, mc, borderBox)
	}
}

// renderGenericBlockNode provides basic rendering for all block nodes
// except specific block nodes.
func (r *Renderer) renderGenericBlockNode(n ast.Node, mc MeasureContext, borderBox HalfBounds, borderBoxBottom ...VerticalCoord) (Rect, error) {
	bs := r.blockStyle(n)

	err := mc.GetRenderContext(func(rc RenderContext) error {
		var b Rect
		if len(borderBoxBottom) == 0 {
			var err error
			b, err = r.renderGenericBlockNode(n, mc, borderBox)
			if err != nil {
				return err
			}
		} else {
			b = borderBox.ToRect(borderBoxBottom[0])
		}

		rc.DrawBox(b, bs.BackgroundColor, bs.Border)
		return nil
	})
	if err != nil {
		return Rect{}, err
	}

	contentBox := borderBox.Shrink(bs.Border, bs.Padding)

	if elements := r.getFlowElements(n); len(elements) != 0 {
		r, err := r.renderInlineElements(elements, mc, contentBox, bs.TextAlign)
		if err != nil {
			return Rect{}, err
		}

		return r.Expand(bs.Border, bs.Padding), nil
	} else {
		// Render descendant block nodes
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			if c.Type() == ast.TypeBlock {
				bs2 := r.blockStyle(c)
				r, err := r.renderBlockNode(c, mc, contentBox.Shrink(bs2.Margin))
				if err != nil {
					return Rect{}, err
				}

				contentBox.Top = r.Bottom
				contentBox.Top.Position += bottom(bs2.Margin) // TODO Collapse vertical margins
			}
		}

		boxBottom := contentBox.Top
		boxBottom.Position += bottom(bs.Padding) + bottom(bs.Border)

		return borderBox.ToRect(boxBottom), nil
	}
}
