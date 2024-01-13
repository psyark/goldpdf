package goldpdftest

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
)

func CompareImages(wantBytes, gotBytes []byte) (image.Image, error) {
	if bytes.Equal(wantBytes, gotBytes) {
		return nil, nil
	}

	want, err := png.Decode(bytes.NewReader(wantBytes))
	if err != nil {
		return nil, err
	}

	got, err := png.Decode(bytes.NewReader(gotBytes))
	if err != nil {
		return nil, err
	}

	if want.Bounds() != got.Bounds() {
		return nil, fmt.Errorf("bounds mismatch")
	}

	equals := true
	di := image.NewGray(want.Bounds())
	for y := di.Rect.Min.Y; y < di.Rect.Max.Y; y++ {
		for x := di.Rect.Min.X; x < di.Rect.Max.X; x++ {
			if !colorEquals(want.At(x, y), got.At(x, y)) {
				di.Set(x, y, color.White)
				equals = false
			}
		}
	}

	if equals {
		return nil, nil
	} else {
		return di, nil
	}
}

func colorEquals(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
