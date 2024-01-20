package goldpdf

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

func (rc *RenderContext) Preflight(fn func() error) error {
	if !rc.inPreflight {
		rc.inPreflight = true
		defer func() { rc.inPreflight = false }()
		return fn()
	}
	return nil
}
