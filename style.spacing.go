package goldpdf

import (
	"image/color"
)

var (
	_ Spacer = Spacing{}
	_ Border = UniformBorder{}
	_ Border = IndividualBorder{}
)

// Spacer はボックス間のサイズを計算する際にマージン・ボーダー・パディングを共通で扱うためのインターフェースです
type Spacer interface {
	Space() (float64, float64, float64, float64)
}

func horizontal(spacer Spacer) float64 {
	if spacer == nil {
		return 0
	}
	l, _, r, _ := spacer.Space()
	return l + r
}
func vertical(spacer Spacer) float64 {
	if spacer == nil {
		return 0
	}
	_, t, _, b := spacer.Space()
	return t + b
}
func bottom(spacer Spacer) float64 {
	if spacer == nil {
		return 0
	}
	_, _, _, b := spacer.Space()
	return b
}

// Spacing は単純な余白です
type Spacing struct {
	Left, Top, Right, Bottom float64
}

func (s Spacing) Space() (float64, float64, float64, float64) {
	return s.Left, s.Top, s.Right, s.Bottom
}

type Border interface {
	Spacer
}

type UniformBorder struct {
	Width  float64
	Color  color.Color
	Radius float64
}

func (b UniformBorder) Space() (float64, float64, float64, float64) {
	return b.Width, b.Width, b.Width, b.Width
}

type IndividualBorder struct {
	Left, Top, Right, Bottom BorderEdge
}

func (b IndividualBorder) Space() (float64, float64, float64, float64) {
	return b.Left.Width, b.Top.Width, b.Right.Width, b.Bottom.Width
}

type BorderEdge struct {
	Width float64
	Color color.Color
}
