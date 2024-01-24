package goldpdftest

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"gopkg.in/gographics/imagick.v3/imagick"
)

func CompareAndOutputResults(pdfBytes []byte, pdfName, wantName, gotName, diffName string) error {
	if err := os.WriteFile(pdfName, pdfBytes, 0666); err != nil {
		return err
	}

	got, err := capturePDF(pdfBytes, color.White)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	if err := png.Encode(buf, got); err != nil {
		return err
	}

	gotBytes := buf.Bytes()

	if wantBytes, err := os.ReadFile(wantName); !os.IsNotExist(err) {
		diff, err := CompareImages(wantBytes, gotBytes)
		if err != nil {
			return err
		}
		if diff != nil {
			f, err := os.OpenFile(diffName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			if err != nil {
				return err
			}

			defer f.Close()
			if err := png.Encode(f, diff); err != nil {
				return err
			}

			if err := os.WriteFile(gotName, gotBytes, 0666); err != nil {
				return err
			}
			return fmt.Errorf("mismatch")
		} else {
			os.Remove(diffName)
			os.Remove(gotName)
		}
	} else {
		os.Remove(diffName)
		os.Remove(gotName)
		if err := os.WriteFile(wantName, gotBytes, 0666); err != nil {
			return err
		}
	}
	return nil
}

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

func capturePDF(pdfBytes []byte, bgColor color.Color) (image.Image, error) {
	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if err := mw.ReadImageBlob(pdfBytes); err != nil {
		return nil, err
	}

	var bg *image.RGBA

	for i := 0; i < int(mw.GetNumberImages()); i++ {
		mw.SetIteratorIndex(i)
		if err := mw.SetImageFormat("png"); err != nil {
			return nil, err
		}
		img, err := png.Decode(bytes.NewReader(mw.GetImageBlob()))
		if err != nil {
			return nil, err
		}

		if i == 0 {
			rect := img.Bounds()
			rect.Max.Y *= int(mw.GetNumberImages())

			bg = image.NewRGBA(rect)
			draw.Draw(bg, bg.Rect, image.NewUniform(bgColor), image.Point{}, draw.Over)
		}

		draw.Draw(bg, image.Rect(0, img.Bounds().Dy()*i, img.Bounds().Dx(), img.Bounds().Dy()*(i+1)), img, image.Point{}, draw.Over)
	}

	fmt.Println(bg.Rect)
	return bg, nil
}
