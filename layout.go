package goldpdf

// HalfBounds represents a bounds on the left, right and top coordinates.
type HalfBounds struct {
	Left, Right float64
	Top         VerticalCoord
	// CurrentMargin float64
}

func (b HalfBounds) Width() float64 {
	return b.Right - b.Left
}

func (b HalfBounds) Shrink(spacers ...Spacer) HalfBounds {
	for _, s := range spacers {
		if s != nil {
			ls, ts, rs, _ := s.Space()
			b.Left += ls
			b.Top.Position += ts
			b.Right -= rs
		}
	}
	return b
}

func (b HalfBounds) Expand(spacers ...Spacer) HalfBounds {
	for _, s := range spacers {
		if s != nil {
			ls, ts, rs, _ := s.Space()
			b.Left -= ls
			b.Top.Position -= ts // TODO: Positionが負数にならないか確認
			b.Right += rs
		}
	}
	return b
}

func (b HalfBounds) ToRect(Bottom VerticalCoord) Rect {
	return Rect{
		Left:   b.Left,
		Right:  b.Right,
		Top:    b.Top,
		Bottom: Bottom,
	}
}

// Rect is a rectangle that can span multiple pages of a PDF document.
type Rect struct {
	Left, Right float64
	Top, Bottom VerticalCoord
}

func (r Rect) Width() float64 {
	return r.Right - r.Left
}

func (r Rect) Shrink(spacers ...Spacer) Rect {
	for _, s := range spacers {
		if s != nil {
			ls, ts, rs, bs := s.Space()
			r.Left += ls
			r.Top.Position += ts
			r.Right -= rs
			r.Bottom.Position -= bs // TODO: Positionが負数にならないか確認
		}
	}
	return r
}

func (r Rect) Expand(spacers ...Spacer) Rect {
	for _, s := range spacers {
		if s != nil {
			ls, ts, rs, bs := s.Space()
			r.Left -= ls
			r.Top.Position -= ts // TODO: Positionが負数にならないか確認
			r.Right += rs
			r.Bottom.Position += bs
		}
	}
	return r
}

func (r Rect) ToHalfBounds() HalfBounds {
	return HalfBounds{
		Left:  r.Left,
		Right: r.Right,
		Top:   r.Top,
	}
}

type VerticalCoord struct {
	Page     int
	Position float64
}

func (vc VerticalCoord) LessThan(vc2 VerticalCoord) bool {
	// If the page is the same page, compare with the position in the page
	if vc.Page == vc2.Page && vc.Position < vc2.Position {
		return true
	}
	// Otherwise, compare with the page number
	return vc.Page < vc2.Page
}
