package goldpdf

import "github.com/yuin/goldmark/ast"

type State struct {
	Node  ast.Node
	X, Y  float64
	Style Style
	Link  string
}
