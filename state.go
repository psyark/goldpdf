package goldpdf

import "github.com/yuin/goldmark/ast"

type State struct {
	Node       ast.Node
	Style      Style
	Link       string
	XMin, XMax float64
}
