package goldpdf

type RenderContext struct {
	X, Y, W   float64
	Preflight bool
	Target    PDF

	// FencedCodeBlockなど、子孫ノード以外のFlowElementを持つノードが内部でrenderGenericBlockNodeを利用できるようにするため
	// renderGenericBlockNodeは次の呼び出しには渡さない
	// TODO renderXxxのようにgetFlowElemensXxxのようなフックを作って対応する？
	// TODO RenderContextに置くか、個別のRender関数の可変長引数にするか検討
	FlowElements FlowElements

	// Table -> TableRow / TableHeader -> TableCell にセル幅とAlignを渡すため
	// TODO RenderContextに置くか、個別のRender関数の可変長引数にするか検討
	CellFormats []TableCellFormat
}

type TableCellFormat struct {
	Width float64
	Align string
}

// TODO 必要か確認する
func (rc RenderContext) Extend(dx, dy, dw float64) RenderContext {
	rc.X += dx
	rc.Y += dy
	rc.W += dw
	return rc
}

func (rc RenderContext) Shrink(spacers ...Spacer) RenderContext {
	for _, s := range spacers {
		l, t, r, _ := s.Space()
		rc.X += l
		rc.Y += t
		rc.W -= l + r
	}
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
