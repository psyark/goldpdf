package goldpdf

type RenderContext struct {
	X, Y, W   float64
	Preflight bool
	Target    PDF
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

func (rc RenderContext) InPreflight() RenderContext {
	rc.Preflight = true
	return rc
}
