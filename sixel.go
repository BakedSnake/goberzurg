package goberzurg

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
)

type SixelBackend struct {
	w      io.Writer
	closed bool
}

func NewSixelBackend() *SixelBackend {
	return &SixelBackend{w: os.Stdout}
}

func (s *SixelBackend) Name() string { return "sixel" }

func (s *SixelBackend) Display(key string, img *Image, opts Options) error {
	if s.closed {
		return nil
	}
	if img.Decoded == nil {
		return fmt.Errorf("sixel backend requires a decodable image")
	}
	sixelData := encodeSixel(img.Decoded)
	_, err := fmt.Fprintf(s.w, "\x1bPq%s\x1b\\", sixelData)
	return err
}

func (s *SixelBackend) Clear() error {
	if s.closed {
		return nil
	}
	_, err := fmt.Fprintf(s.w, "\x1bPq\x1b\\")
	return err
}

func (s *SixelBackend) Close() error {
	if s.closed {
		return nil
	}
	s.closed = true
	return nil
}

func encodeSixel(img image.Image) string {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	palette := buildPalette(img, 256)

	var result string
	for i, c := range palette {
		r, g, b, _ := c.RGBA()
		result += fmt.Sprintf("#%d;2;%d;%d;%d", i, r>>8, g>>8, b>>8)
	}

	for x := 0; x < w; x++ {
		for ci, pc := range palette {
			var any bool
			for y := 0; y < h; y++ {
				if colorMatch(img.At(x+bounds.Min.X, y+bounds.Min.Y), pc) {
					any = true
					break
				}
			}
			if !any {
				continue
			}

			result += fmt.Sprintf("#%d", ci)

			for y := 0; y < h; y += 6 {
				var val byte
				for b := 0; b < 6 && y+b < h; b++ {
					if colorMatch(img.At(x+bounds.Min.X, y+bounds.Min.Y+b), pc) {
						val |= 1 << uint(b)
					}
				}
				result += string(rune(63 + val))
			}
			result += "$"
		}
		result += "-"
	}

	return result
}

func colorMatch(a, b color.Color) bool {
	if a == b {
		return true
	}
	r1, g1, b1, _ := a.RGBA()
	r2, g2, b2, _ := b.RGBA()
	dr := int64(r1) - int64(r2)
	dg := int64(g1) - int64(g2)
	db := int64(b1) - int64(b2)
	return dr*dr+dg*dg+db*db < 50000
}

func buildPalette(img image.Image, maxColors int) []color.Color {
	bounds := img.Bounds()
	if maxColors < 1 {
		maxColors = 1
	}
	if maxColors > 256 {
		maxColors = 256
	}

	type count struct {
		r, g, b uint32
		n       int
	}
	grid := make(map[[3]uint16]*count)

	step := uint32(16)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			key := [3]uint16{uint16(r / step), uint16(g / step), uint16(b / step)}
			if c, ok := grid[key]; ok {
				c.n++
			} else {
				grid[key] = &count{r: r, g: g, b: b, n: 1}
			}
		}
	}

	sorted := make([]*count, 0, len(grid))
	for _, c := range grid {
		sorted = append(sorted, c)
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].n > sorted[i].n {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	if len(sorted) > maxColors {
		sorted = sorted[:maxColors]
	}

	if len(sorted) == 0 {
		return []color.Color{color.Black, color.White}
	}

	pal := make([]color.Color, len(sorted))
	for i, c := range sorted {
		pal[i] = color.RGBA{
			R: uint8(c.r >> 8),
			G: uint8(c.g >> 8),
			B: uint8(c.b >> 8),
			A: 255,
		}
	}

	return pal
}
