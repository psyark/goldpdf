package goldpdf

// Rect is a rectangle that can span multiple pages of a PDF document.
// If Bottom is nil, it means that this Rect has infinite size downwards.
type Rect struct {
	Left, Right float64
	Top, Bottom *VerticalCoord
}

type VerticalCoord struct {
	Page     int
	Position float64
}
