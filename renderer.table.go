package goldpdf

import (
	"github.com/yuin/goldmark/ast"
	xast "github.com/yuin/goldmark/extension/ast"
)

func (r *Renderer) renderTable(n *xast.Table, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(20)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableHeader(n *xast.TableHeader, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(20)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableRow(n *xast.TableRow, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(20)
	}
	return ast.WalkContinue, nil
}

func (r *Renderer) renderTableCell(n *xast.TableCell, entering bool) (ast.WalkStatus, error) {
	if entering {
		r.pdf.Ln(20)
	}
	return ast.WalkContinue, nil
}
