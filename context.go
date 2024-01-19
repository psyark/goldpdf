package goldpdf

type RenderContext struct {
	X, Y, W   float64
	Preflight bool
	Target    PDF

	// FencedCodeBlockなど、子孫ノード以外のFlowElementを持つノードが内部でrenderGenericBlockNodeを利用できるようにするため
	// renderGenericBlockNodeは次の呼び出しには渡さない
	// TODO renderXxxのようにgetFlowElemensXxxのようなフックを作って対応する？
	FlowElements FlowElements

	// Table -> TableRow / TableHeader -> TableCell にセル幅とAlignを渡すため
	CellFormats []TableCellFormat
}

type TableCellFormat struct {
	Width float64
	Align string
}

func (rc RenderContext) Extend(dx, dy, dw float64) RenderContext {
	rc.X += dx
	rc.Y += dy
	rc.W += dw
	return rc
}

func (rc RenderContext) InPreflight() RenderContext {
	rc.Preflight = true
	return rc
}

func (rc RenderContext) WithFlowElements(fe FlowElements) RenderContext {
	rc.FlowElements = append(FlowElements{}, fe...)
	return rc
}
