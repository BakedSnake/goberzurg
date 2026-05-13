package goberzurg

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"
)

type Image struct {
	Format string
	Width  int
	Height int
	Data   []byte
	Decoded image.Image
}

func Open(path string) (*Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	img, fmt, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return &Image{
			Format: detectFormat(data),
			Data:   data,
		}, nil
	}

	imgFmt := fmt
	if imgFmt == "png" || imgFmt == "jpeg" {
		imgFmt = fmt
	}

	return &Image{
		Format:  imgFmt,
		Width:   img.Bounds().Dx(),
		Height:  img.Bounds().Dy(),
		Data:    data,
		Decoded: img,
	}, nil
}

func (img *Image) EncodePNG() ([]byte, error) {
	if img.Decoded == nil {
		return img.Data, nil
	}
	if strings.EqualFold(img.Format, "png") {
		return img.Data, nil
	}
	var buf bytes.Buffer
	err := png.Encode(&buf, img.Decoded)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (img *Image) EncodeJPEG() ([]byte, error) {
	if img.Decoded == nil {
		return img.Data, nil
	}
	if strings.EqualFold(img.Format, "jpeg") || strings.EqualFold(img.Format, "jpg") {
		return img.Data, nil
	}
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, img.Decoded, &jpeg.Options{Quality: 95})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const pxPerCell = 10

func ScaleToCells(img *Image, cellW, cellH int) (*Image, error) {
	if (cellW <= 0 && cellH <= 0) || img.Decoded == nil {
		return img, nil
	}

	srcW := img.Width
	srcH := img.Height

	targetW := cellW * pxPerCell
	targetH := cellH * pxPerCell

	if targetW <= 0 {
		r := float64(targetH) / float64(srcH)
		targetW = int(float64(srcW) * r)
	}
	if targetH <= 0 {
		r := float64(targetW) / float64(srcW)
		targetH = int(float64(srcH) * r)
	}
	if targetW <= 0 {
		targetW = srcW
	}
	if targetH <= 0 {
		targetH = srcH
	}

	dst := nearestNeighbor(img.Decoded, targetW, targetH)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}

	return &Image{
		Format:  "png",
		Width:   targetW,
		Height:  targetH,
		Data:    buf.Bytes(),
		Decoded: dst,
	}, nil
}

func nearestNeighbor(src image.Image, dstW, dstH int) *image.RGBA {
	b := src.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for dy := 0; dy < dstH; dy++ {
		sy := b.Min.Y + dy*srcH/dstH
		for dx := 0; dx < dstW; dx++ {
			sx := b.Min.X + dx*srcW/dstW
			r, g, bv, a := src.At(sx, sy).RGBA()
			off := dst.PixOffset(dx, dy)
			dst.Pix[off+0] = uint8(r >> 8)
			dst.Pix[off+1] = uint8(g >> 8)
			dst.Pix[off+2] = uint8(bv >> 8)
			dst.Pix[off+3] = uint8(a >> 8)
		}
	}

	return dst
}

func bilinear(src image.Image, dstW, dstH int) *image.RGBA {
	b := src.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))

	for dy := 0; dy < dstH; dy++ {
		for dx := 0; dx < dstW; dx++ {
			xf := float64(dx) * float64(srcW) / float64(dstW)
			yf := float64(dy) * float64(srcH) / float64(dstH)

			ix := int(xf)
			iy := int(yf)
			fx := xf - float64(ix)
			fy := yf - float64(iy)

			if ix >= srcW-1 {
				ix = srcW - 2
			}
			if iy >= srcH-1 {
				iy = srcH - 2
			}
			if ix < 0 {
				ix = 0
			}
			if iy < 0 {
				iy = 0
			}

			ax := b.Min.X + ix
			ay := b.Min.Y + iy

			c00 := src.At(ax, ay)
			c10 := src.At(ax+1, ay)
			c01 := src.At(ax, ay+1)
			c11 := src.At(ax+1, ay+1)

			r := bilerp(component(c00, 0), component(c10, 0), component(c01, 0), component(c11, 0), fx, fy)
			g := bilerp(component(c00, 1), component(c10, 1), component(c01, 1), component(c11, 1), fx, fy)
			bv := bilerp(component(c00, 2), component(c10, 2), component(c01, 2), component(c11, 2), fx, fy)
			a := bilerp(component(c00, 3), component(c10, 3), component(c01, 3), component(c11, 3), fx, fy)

			off := dst.PixOffset(dx, dy)
			dst.Pix[off+0] = uint8(r >> 8)
			dst.Pix[off+1] = uint8(g >> 8)
			dst.Pix[off+2] = uint8(bv >> 8)
			dst.Pix[off+3] = uint8(a >> 8)
		}
	}

	return dst
}

func bilerp(v00, v10, v01, v11 uint32, fx, fy float64) uint32 {
	v0 := float64(v00)*(1-fx) + float64(v10)*fx
	v1 := float64(v01)*(1-fx) + float64(v11)*fx
	return uint32(v0*(1-fy) + v1*fy)
}

func component(c color.Color, idx int) uint32 {
	r, g, b, a := c.RGBA()
	switch idx {
	case 0:
		return r
	case 1:
		return g
	case 2:
		return b
	default:
		return a
	}
}

func detectFormat(data []byte) string {
	if len(data) < 4 {
		if len(data) >= 2 && bytes.HasPrefix(data, []byte{0xFF, 0xD8}) {
			return "jpeg"
		}
		return "unknown"
	}
	switch {
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}):
		return "png"
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8}):
		return "jpeg"
	case bytes.HasPrefix(data, []byte{0x47, 0x49, 0x46}):
		return "gif"
	case bytes.HasPrefix(data, []byte{0x52, 0x49, 0x46, 0x46}):
		return "webp"
	default:
		return "unknown"
	}
}
