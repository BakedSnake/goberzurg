package goberzurg

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"
)

const kittyChunkSize = 4096

type KittyBackend struct {
	w        io.Writer
	closed   bool
	mu       sync.Mutex
	imageIDs map[string]uint32
	nextID   uint32
}

func NewKittyBackend() *KittyBackend {
	return &KittyBackend{
		w:        os.Stdout,
		imageIDs: make(map[string]uint32),
		nextID:   1,
	}
}

func (k *KittyBackend) Name() string { return "kitty" }

func (k *KittyBackend) Display(key string, img *Image, opts Options) error {
	if k.closed {
		return nil
	}

	k.mu.Lock()
	if k.imageIDs == nil {
		k.imageIDs = make(map[string]uint32)
	}
	id, exists := k.imageIDs[key]
	if !exists {
		id = k.nextID
		k.nextID++
		k.imageIDs[key] = id
	}
	k.mu.Unlock()

	if !exists {
		if err := k.transmit(id, img); err != nil {
			return err
		}
	}

	return k.place(id, opts)
}

func (k *KittyBackend) transmit(id uint32, img *Image) error {
	pngData, err := img.EncodePNG()
	if err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	if len(pngData) == 0 {
		return fmt.Errorf("empty image data")
	}

	b64 := base64.StdEncoding.EncodeToString(pngData)

	payload := fmt.Sprintf("a=T,i=%d,f=100,t=d,s=%d,v=%d",
		id, img.Width, img.Height)

	if len(b64) <= kittyChunkSize {
		fmt.Fprintf(k.w, "\x1b_G%s,m=0;%s\x1b\\", payload, b64)
	} else {
		chunks := chunkString(b64, kittyChunkSize)
		for i, ch := range chunks {
			isLast := i == len(chunks)-1
			m := 0
			if !isLast {
				m = 1
			}
			if i == 0 {
				fmt.Fprintf(k.w, "\x1b_G%s,m=%d;%s\x1b\\", payload, m, ch)
			} else {
				fmt.Fprintf(k.w, "\x1b_Gm=%d;%s\x1b\\", m, ch)
			}
		}
	}

	return nil
}

func (k *KittyBackend) place(id uint32, opts Options) error {
	var parts []string

	if opts.X != 0 {
		parts = append(parts, fmt.Sprintf("c=%d", opts.X))
	}
	if opts.Y != 0 {
		parts = append(parts, fmt.Sprintf("r=%d", opts.Y))
	}
	if opts.Width != 0 {
		parts = append(parts, fmt.Sprintf("w=%d", opts.Width))
	}
	if opts.Height != 0 {
		parts = append(parts, fmt.Sprintf("h=%d", opts.Height))
	}
	if opts.ZIndex != 0 {
		parts = append(parts, fmt.Sprintf("z=%d", opts.ZIndex))
	}

	cmd := fmt.Sprintf("a=p,i=%d", id)
	if len(parts) > 0 {
		cmd += "," + join(parts, ",")
	}
	fmt.Fprintf(k.w, "\x1b_G%s\x1b\\", cmd)

	return nil
}

func (k *KittyBackend) Clear() error {
	if k.closed {
		return nil
	}
	k.mu.Lock()
	k.imageIDs = make(map[string]uint32)
	k.mu.Unlock()
	fmt.Fprintf(k.w, "\x1b_Ga=d,d=a\x1b\\")
	return nil
}

func (k *KittyBackend) Close() error {
	if k.closed {
		return nil
	}
	k.closed = true
	return nil
}

func chunkString(s string, size int) []string {
	if len(s) == 0 {
		return nil
	}
	var chunks []string
	for i := 0; i < len(s); i += size {
		end := i + size
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	n := len(sep) * (len(parts) - 1)
	for _, p := range parts {
		n += len(p)
	}
	b := make([]byte, n)
	i := 0
	for idx, p := range parts {
		if idx > 0 {
			copy(b[i:], sep)
			i += len(sep)
		}
		copy(b[i:], p)
		i += len(p)
	}
	return string(b)
}
