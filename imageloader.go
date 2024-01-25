package goldpdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/raykov/oksvg"
	"github.com/srwiley/rasterx"
)

type DefaultImageLoaderErrorMode int

const (
	ReturnError DefaultImageLoaderErrorMode = iota
	IgnoreErrorAndShowAlt
)

var (
	_ ImageLoader = &DefaultImageLoader{}
)

type ImageLoader interface {
	LoadImage(string) (*ImageElement, error)
}

type DefaultImageLoader struct {
	ErrorMode DefaultImageLoaderErrorMode
	cache     map[string]*ImageElement
}

func (il *DefaultImageLoader) LoadImage(src string) (img *ImageElement, err error) {
	defer func() {
		if err != nil && il.ErrorMode == IgnoreErrorAndShowAlt {
			err = nil
		}
	}()

	if il.cache == nil {
		il.cache = map[string]*ImageElement{}
	}

	if img, ok := il.cache[src]; ok {
		return img, nil
	}

	il.cache[src] = nil

	switch {
	case strings.HasPrefix(src, "http:"), strings.HasPrefix(src, "https:"):
		resp, err := http.Get(src)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		if err := il.registerBytes(src, resp.Header.Get("Content-Type"), data); err != nil {
			return nil, err
		}

		return il.cache[src], nil
	case strings.HasPrefix(src, "data:"):
		if ind := strings.Index(src, ";base64,"); ind != -1 {
			contentType := src[5:ind]
			ind += len(";base64,")
			data, err := base64.StdEncoding.DecodeString(src[ind:])
			if err != nil {
				return nil, err
			}

			if err := il.registerBytes(src, contentType, data); err != nil {
				return nil, err
			}

			return il.cache[src], nil
		}
		return nil, fmt.Errorf("unsupported")
	default:
		return nil, fmt.Errorf("unsupported")
	}
}

func (il *DefaultImageLoader) registerBytes(src string, mimeType string, data []byte) error {
	var img image.Image
	var imgType string

	if mimeType == "image/svg+xml" {
		icon, err := oksvg.ReadIconStream(bytes.NewReader(data), oksvg.StrictErrorMode)
		if err != nil {
			return err
		}

		w, h := int(icon.ViewBox.W), int(icon.ViewBox.H)
		drawImg := image.NewRGBA(image.Rect(0, 0, w, h))
		img = drawImg

		raster := rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, drawImg, img.Bounds()))
		icon.Draw(raster, 1.0)
		icon.DrawTexts(drawImg, 1.0)

		imgType = "png"
		buf := bytes.NewBuffer(nil)
		if err := png.Encode(buf, img); err != nil {
			return err
		}

		data = buf.Bytes()
	} else {
		var err error
		img, imgType, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
	}

	name := strconv.Itoa(len(il.cache))
	il.cache[src] = &ImageElement{
		name:      name,
		imageType: imgType,
		img:       img,
		data:      data,
	}
	return nil
}
