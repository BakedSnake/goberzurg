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

	scaled, err := ScaleToCells(img, opts.Width, opts.Height)
	if err != nil {
		return err
	}

	cacheKey := key
	if opts.Width > 0 || opts.Height > 0 {
		cacheKey = fmt.Sprintf("%s:%dx%d", key, opts.Width, opts.Height)
	}

	k.mu.Lock()
	if k.imageIDs == nil {
		k.imageIDs = make(map[string]uint32)
	}
	id, exists := k.imageIDs[cacheKey]
	if !exists {
		id = k.nextID
		k.nextID++
		k.imageIDs[cacheKey] = id
	}
	k.mu.Unlock()

	if !exists {
		if err := k.transmit(id, scaled); err != nil {
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

	payload := fmt.Sprintf("a=t,i=%d,f=100,t=d,s=%d,v=%d,q=1",
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

	parts = append(parts, "C=1")
	if opts.ZIndex != 0 {
		parts = append(parts, fmt.Sprintf("z=%d", opts.ZIndex))
	}

	cmd := fmt.Sprintf("a=p,i=%d,q=1", id)
	if len(parts) > 0 {
		cmd += "," + join(parts, ",")
	}

	// The Kitty protocol places images at the cursor position.
	// Move cursor to (Y, X) before placing (ANSI CUP is 1-indexed: row, col).
	fmt.Fprintf(k.w, "\x1b7\x1b[%d;%dH\x1b_G%s\x1b\\\x1b8",
		opts.Y+1, opts.X+1, cmd)

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
