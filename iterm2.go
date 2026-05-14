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
	i := &Iterm2Backend{}
	i.w = NewClearOnResetWriter(os.Stdout, i.Clear)
	return i
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
	// Move cursor to position (Y, X) before rendering
	tmuxWrite(i.w, fmt.Sprintf("\x1b7\x1b[%d;%dH\x1b]1337;File=inline=1;size=%d;name=%s:%s\x07\x1b8",
		opts.Y+1, opts.X+1, len(scaled.Data), name, b64))

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
