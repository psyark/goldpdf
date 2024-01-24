package goldpdf

import (
	"bytes"
	"encoding/base64"
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

var (
	_ ImageLoader = &DefaultImageLoader{}
)

type ImageLoader interface {
	LoadImage(string) *ImageElement // TODO error
}

type DefaultImageLoader struct {
	cache map[string]*ImageElement
	// TODO 動作モード（エラートレラントでaltを表示するか、呼び出し元にエラーを返すか）
}

func (il *DefaultImageLoader) LoadImage(src string) *ImageElement {
	if il.cache == nil {
		il.cache = map[string]*ImageElement{}
	}

	if img, ok := il.cache[src]; ok {
		return img
	}

	il.cache[src] = nil

	switch {
	case strings.HasPrefix(src, "http:"), strings.HasPrefix(src, "https:"):
		resp, err := http.Get(src)
		if err != nil {
			return nil
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil
		}

		if err := il.registerBytes(src, resp.Header.Get("Content-Type"), data); err != nil {
			return nil
		}

		return il.cache[src]
	case strings.HasPrefix(src, "data:"):
		if ind := strings.Index(src, ";base64,"); ind != -1 {
			contentType := src[5:ind]
			ind += len(";base64,")
			data, err := base64.StdEncoding.DecodeString(src[ind:])
			if err != nil {
				return nil
			}

			if err := il.registerBytes(src, contentType, data); err != nil {
				return nil
			}

			return il.cache[src]
		}
		return nil
	default:
		return nil // unsupported
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
		img_ := image.NewRGBA(image.Rect(0, 0, w, h))
		img = img_

		raster := rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, img_, img.Bounds()))
		icon.Draw(raster, 1.0)
		icon.DrawTexts(img_, 1.0)

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
			// fmt.Println(src)
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
