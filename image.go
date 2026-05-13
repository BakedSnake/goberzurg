package goberzurg

import (
	"bytes"
	"image"
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
