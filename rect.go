package goldpdf

// Rect is a rectangle that can span multiple pages of a PDF document.
// If Bottom is nil, it means that this Rect has infinite size downwards.
type Rect struct {
	Left, Right float64
	Top, Bottom VerticalCoord
	HasBottom   bool // TODO これでいいか検討
}

func (r Rect) Width() float64 {
	return r.Right - r.Left
}

func (rc Rect) Shrink(spacers ...Spacer) Rect {
	for _, s := range spacers {
		if s != nil {
			l, t, r, b := s.Space()
			rc.Left += l
			rc.Top.Position += t
			rc.Right -= r
			rc.Bottom.Position -= b // TODO: Positionが負数にならないか確認
		}
	}
	return rc
}

func (rc Rect) Expand(spacers ...Spacer) Rect {
	for _, s := range spacers {
		if s != nil {
			l, t, r, b := s.Space()
			rc.Left -= l
			rc.Top.Position -= t // TODO: Positionが負数にならないか確認
			rc.Right += r
			rc.Bottom.Position += b
		}
	}
	return rc
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
