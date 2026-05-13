package goberzurg

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

type Iterm2Backend struct {
	w      io.Writer
	closed bool
}

func NewIterm2Backend() *Iterm2Backend {
	return &Iterm2Backend{w: os.Stdout}
}

func (i *Iterm2Backend) Name() string { return "iterm2" }

func (i *Iterm2Backend) Display(key string, img *Image, opts Options) error {
	if i.closed {
		return nil
	}

	scaled, err := ScaleToCells(img, opts.Width, opts.Height)
	if err != nil {
		return err
	}

	var name string
	switch scaled.Format {
	case "png":
		name = "image.png"
	case "jpeg", "jpg":
		name = "image.jpg"
	case "gif":
		name = "image.gif"
	default:
		name = "image"
	}

	b64 := base64.StdEncoding.EncodeToString(scaled.Data)
	_, err = fmt.Fprintf(i.w, "\x1b]1337;File=inline=1;size=%d;name=%s:%s\x07",
		len(scaled.Data), name, b64)

	if opts.Y > 0 && opts.X == 0 {
		fmt.Fprintf(i.w, "\x1b[%dB", opts.Y)
	}

	return err
}

func (i *Iterm2Backend) Clear() error {
	if i.closed {
		return nil
	}
	return nil
}

func (i *Iterm2Backend) Close() error {
	if i.closed {
		return nil
	}
	i.closed = true
	return nil
}
