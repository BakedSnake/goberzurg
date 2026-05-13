package goberzurg

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"
	"testing"
)

func makeTestPNG(t *testing.T) ([]byte, int, int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 25), uint8(y * 25), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), 10, 10
}

func writeTempPNG(t *testing.T) string {
	t.Helper()
	data, _, _ := makeTestPNG(t)
	tmp := t.TempDir() + "/test.png"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		t.Fatal(err)
	}
	return tmp
}

func TestOptions(t *testing.T) {
	var o Options

	WithPos(5, 10)(&o)
	if o.X != 5 || o.Y != 10 {
		t.Fatalf("WithPos: got (%d,%d), want (5,10)", o.X, o.Y)
	}

	WithSize(100, 200)(&o)
	if o.Width != 100 || o.Height != 200 {
		t.Fatalf("WithSize: got (%d,%d), want (100,200)", o.Width, o.Height)
	}

	WithZIndex(3)(&o)
	if o.ZIndex != 3 {
		t.Fatalf("WithZIndex: got %d, want 3", o.ZIndex)
	}
}

func TestOpenPNG(t *testing.T) {
	path := writeTempPNG(t)
	img, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if img.Width != 10 || img.Height != 10 {
		t.Fatalf("dimensions: got %dx%d, want 10x10", img.Width, img.Height)
	}
	if img.Format != "png" {
		t.Fatalf("format: got %s, want png", img.Format)
	}
	if img.Decoded == nil {
		t.Fatal("Decoded is nil")
	}
}

func TestEncodePNG(t *testing.T) {
	path := writeTempPNG(t)
	img, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := img.EncodePNG()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		t.Fatal("encoded data does not have PNG magic bytes")
	}
}

func TestOpenNonImage(t *testing.T) {
	tmp := t.TempDir() + "/notanimage.bin"
	if err := os.WriteFile(tmp, []byte{0, 1, 2, 3, 4, 5, 6, 7}, 0644); err != nil {
		t.Fatal(err)
	}
	img, err := Open(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if img.Decoded != nil {
		t.Fatal("expected Decoded to be nil for non-image")
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		data []byte
		fmt  string
	}{
		{[]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "png"},
		{[]byte{0xFF, 0xD8, 0xFF, 0xE0}, "jpeg"},
		{[]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61}, "gif"},
		{[]byte{0x47, 0x49, 0x46, 0x38, 0x37, 0x61}, "gif"},
		{[]byte{0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}, "unknown"},
	}
	for _, tc := range tests {
		got := detectFormat(tc.data)
		if got != tc.fmt {
			t.Errorf("detectFormat(%x) = %s, want %s", tc.data, got, tc.fmt)
		}
	}
}

func TestKittyBackendTransmit(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf}
	img := &Image{Width: 10, Height: 10, Format: "png", Data: []byte{0x89, 0x50, 0x4E, 0x47}}

	err := b.Display("test", img, Options{})
	if err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.HasPrefix(out, "\x1b_G") {
		t.Fatalf("output does not start with ESC_G: %q", out[:min(len(out), 20)])
	}
	if !strings.Contains(out, "a=T") {
		t.Fatal("output missing a=T (transmit)")
	}
	if !strings.Contains(out, "a=p") {
		t.Fatal("output missing a=p (place)")
	}
}

func TestKittyBackendCaching(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf, imageIDs: make(map[string]uint32), nextID: 1}
	img := &Image{Width: 1, Height: 1, Format: "png", Data: []byte{0x89, 0x50, 0x4E, 0x47}}

	err := b.Display("key", img, Options{Pos: Pos{X: 1, Y: 1}})
	if err != nil {
		t.Fatal(err)
	}

	first := buf.Len()

	err = b.Display("key", img, Options{Pos: Pos{X: 2, Y: 2}})
	if err != nil {
		t.Fatal(err)
	}

	if buf.Len() >= first*2 {
		t.Fatal("expected cached display (no retransmit)")
	}
}

func TestKittyBackendClear(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf}
	err := b.Clear()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "d=a") {
		t.Fatal("Clear should delete all images")
	}
}

func TestIterm2Backend(t *testing.T) {
	var buf bytes.Buffer
	b := &Iterm2Backend{w: &buf}
	img := &Image{Format: "png", Data: []byte{0x89, 0x50, 0x4E, 0x47}}

	err := b.Display("test", img, Options{})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.HasPrefix(out, "\x1b]1337;") {
		t.Fatal("iterm2 output should start with OSC 1337")
	}
	if !strings.Contains(out, "File=inline=1") {
		t.Fatal("missing File=inline=1")
	}
}

func TestSixelBackendEmptyImg(t *testing.T) {
	var buf bytes.Buffer
	b := &SixelBackend{w: &buf}
	img := &Image{Width: 10, Height: 10, Data: []byte{1, 2, 3}}

	err := b.Display("test", img, Options{})
	if err == nil {
		t.Fatal("expected error for undecoded image")
	}
}

func TestRendererNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
	defer r.Close()
	if r.backend == nil {
		t.Fatal("backend should be auto-detected")
	}
}

func TestRendererWithBackend(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf}
	r := New(WithBackend(b))
	defer r.Close()
	if r.Backend() != b {
		t.Fatal("WithBackend not applied")
	}
}

func TestRendererDisplay(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf}
	r := New(WithBackend(b))
	defer r.Close()

	path := writeTempPNG(t)
	err := r.Display(path, WithPos(3, 5))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "a=T") {
		t.Fatal("expected transmit command")
	}
}

func TestRendererClear(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf}
	r := New(WithBackend(b))
	defer r.Close()

	err := r.Clear()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "d=a") {
		t.Fatal("expected delete all command")
	}
}

func TestChunkString(t *testing.T) {
	tests := []struct {
		s    string
		size int
		n    int
	}{
		{"", 4, 0},
		{"abc", 4, 1},
		{"abcdef", 4, 2},
		{"abcdefgh", 4, 2},
		{"abcdefghi", 4, 3},
	}
	for _, tc := range tests {
		chunks := chunkString(tc.s, tc.size)
		if len(chunks) != tc.n {
			t.Errorf("chunkString(%q, %d) = %d chunks, want %d", tc.s, tc.size, len(chunks), tc.n)
		}
		var joined string
		for _, c := range chunks {
			if len(c) > tc.size {
				t.Errorf("chunk too large: %d > %d", len(c), tc.size)
			}
			joined += c
		}
		if joined != tc.s {
			t.Errorf("chunks don't reconstruct: got %q, want %q", joined, tc.s)
		}
	}
}

func TestSixelEncoderCreatesOutput(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 3))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img.Set(1, 0, color.RGBA{0, 255, 0, 255})
	img.Set(2, 0, color.RGBA{0, 0, 255, 255})
	img.Set(0, 1, color.RGBA{255, 255, 0, 255})
	img.Set(1, 1, color.RGBA{0, 255, 255, 255})
	img.Set(2, 1, color.RGBA{255, 0, 255, 255})

	result := encodeSixel(img)
	if result == "" {
		t.Fatal("encodeSixel returned empty string")
	}
	if !strings.Contains(result, "#") {
		t.Fatal("sixel data should contain color definitions")
	}
}

func TestColorMatch(t *testing.T) {
	if !colorMatch(color.RGBA{255, 0, 0, 255}, color.RGBA{255, 0, 0, 255}) {
		t.Fatal("same colors should match")
	}
	if colorMatch(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 255, 0, 255}) {
		t.Fatal("different colors should not match")
	}
}

func TestRendererName(t *testing.T) {
	r := New(WithBackend(NewKittyBackend()))
	defer r.Close()
	if r.Name() != "kitty" {
		t.Fatalf("Name() = %s, want kitty", r.Name())
	}
}

func TestBackendLifecycle(t *testing.T) {
	var buf bytes.Buffer
	b := &KittyBackend{w: &buf}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	if !b.closed {
		t.Fatal("backend should be closed")
	}
	if err := b.Display("x", &Image{}, Options{}); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatal("closed backend should not write")
	}
}
