package goldpdf

import (
	"fmt"

	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

func (r *Renderer) style(n ast.Node) (BlockStyle, TextFormat) {
	ancestors := []ast.Node{}
	for p := n; p != nil; p = p.Parent() {
		ancestors = append(ancestors, p)
	}

	var bs BlockStyle
	var tf TextFormat
	for i := range ancestors {
		bs, tf = r.styler.Style(ancestors[len(ancestors)-i-1], tf)
	}
	return bs, tf
}

// renderBlockNode draws a block node (or document node) inside a borderBox
// and returns the height of its border box.
func (r *Renderer) renderBlockNode(n ast.Node, borderBox RenderContext) (float64, error) {
	if n.Type() == ast.TypeInline {
		return 0, fmt.Errorf("renderBlockNode has been called with an inline node: %v > %v", n.Parent().Kind(), n.Kind())
	}

	switch n := n.(type) {
	case *ast.FencedCodeBlock:
		return r.renderFencedCodeBlock(n, borderBox)
	case *ast.ListItem:
		return r.renderListItem(n, borderBox)
	case *xast.Table:
		return r.renderTable(n, borderBox)
	default:
		return r.renderGenericBlockNode(n, borderBox, nil)
	}
}

type rgbnOption struct {
	elements    []FlowElement
	forceHeight float64
}

// renderGenericBlockNode provides basic rendering for all block nodes
// except specific block nodes.
func (r *Renderer) renderGenericBlockNode(n ast.Node, borderBox RenderContext, option *rgbnOption) (float64, error) {
	bs, _ := r.style(n)

	if !borderBox.Preflight {
		var h float64
		var err error
		if option != nil && option.forceHeight != 0 {
			h = option.forceHeight
		} else {
			h, err = r.renderGenericBlockNode(n, borderBox.InPreflight(), option)
			if err != nil {
				return 0, err
			}
		}
		borderBox.Target.DrawBox(
			borderBox.X,
			borderBox.Y,
			borderBox.W,
			h,
			bs.BackgroundColor,
			bs.Border,
		)
	}

	elements := []FlowElement{}
	if option != nil {
		elements = option.elements
	}
	contentBox := borderBox.Shrink(bs.Border, bs.Padding)
	height := top(bs.Border) + top(bs.Padding)

	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Type() {
		case ast.TypeBlock:
			bs2, _ := r.style(c)

			if h, err := r.renderBlockNode(c, contentBox.MoveDown(height).Shrink(bs2.Margin)); err != nil {
				return 0, err
			} else {
				borderBox.Y += h
				height += h
			}
			// TODO マージンの相殺
			height += vertical(bs2.Margin)
		case ast.TypeInline:
			if e, err := r.getFlowElements(c); err != nil {
				return 0, err
			} else {
				elements = append(elements, e...)
			}
		}
	}

	if len(elements) != 0 {
		if h, err := r.renderFlowElements(elements, contentBox, bs.TextAlign); err != nil {
			return 0, err
		} else {
			height += h
		}
	}

	height += bottom(bs.Padding) + bottom(bs.Border)
	return height, nil
}
