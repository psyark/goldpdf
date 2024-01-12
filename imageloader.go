package goldpdf

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type imageInfo struct {
	Name          string
	Type          string
	Width, Height int
	Data          []byte
}

type imageLoader struct {
	cache map[string]*imageInfo
}

func (il *imageLoader) load(src string) *imageInfo {
	if il.cache == nil {
		il.cache = map[string]*imageInfo{}
	}

	if info, ok := il.cache[src]; ok {
		return info
	}

	il.cache[src] = nil

	switch {
	case strings.HasPrefix(src, "http"), strings.HasPrefix(src, "https"):
		resp, err := http.Get(src)
		if err != nil {
			return nil
		}

		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil
		}

		if err := il.registerBytes(src, data); err != nil {
			return nil
		}

		return il.cache[src]
	case strings.HasPrefix(src, "data:"):
		if ind := strings.Index(src, ";base64,"); ind != -1 {
			ind += len(";base64,")
			data, err := base64.StdEncoding.DecodeString(src[ind:])
			if err != nil {
				return nil
			}

			// TODO svg support
			if err := il.registerBytes(src, data); err != nil {
				return nil
			}

			return il.cache[src]
		}
		return nil
	default:
		return nil // unsupported
	}
}

func (il *imageLoader) registerBytes(src string, data []byte) error {
	img, imgType, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}

	name := strconv.Itoa(len(il.cache))
	il.cache[src] = &imageInfo{
		Name:   name,
		Type:   imgType,
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
		Data:   data,
	}
	return nil
}
