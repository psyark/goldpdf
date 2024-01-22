package goldpdf

// TODO X, Y, W は分離
// TODO PDFと統合
type RenderContext struct {
	X, Y, W     float64
	Target      PDF
	inPreflight bool
}

func (rc RenderContext) MoveDown(dy float64) RenderContext {
	rc.Y += dy
	return rc
}

func (rc RenderContext) Shrink(spacers ...Spacer) RenderContext {
	for _, s := range spacers {
		if s != nil {
			l, t, r, _ := s.Space()
			rc.X += l
			rc.Y += t
			rc.W -= l + r
		}
	}
	return rc
}

// Preflight はプリフライトモードを開始します
// プリフライトモード中は、この RenderContext を使った新たなプリフライトモードの開始がスキップされます
// このルールにより、ノードの背景やボーダーを描画する際、ノード固有のレンダリング関数を再帰的に呼び出して子孫を加味した高さを計算させることができ、
// 単一の関数がノードのサイズ計算とノードの描画を担当することができるようになります。
//
// TODO より実態に即した名前をつける
func (rc *RenderContext) Preflight(fn func() error) error {
	if !rc.inPreflight {
		rc.inPreflight = true
		defer func() { rc.inPreflight = false }()
		return fn()
	}
	return nil
}
