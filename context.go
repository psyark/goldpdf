package goldpdf

type RenderContext struct {
	X, Y, W   float64
	Preflight bool
	Target    PDF
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
		if s != nil {
			l, t, r, _ := s.Space()
			rc.X += l
			rc.Y += t
			rc.W -= l + r
		}
	}
	return rc
}

func (rc RenderContext) InPreflight() RenderContext {
	rc.Preflight = true
	return rc
}
