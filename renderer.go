package goberzurg

import (
	"fmt"
	"sync"
)

type Backend interface {
	Display(key string, img *Image, opts Options) error
	Clear() error
	Close() error
	Name() string
}

type Renderer struct {
	mu      sync.Mutex
	backend Backend
	images  map[string]*imageState
	opts    RendererOptions
}

type RendererOptions struct {
	Backend Backend
}

type imageState struct {
	img  *Image
	opts Options
}

func New(opts ...RendererOption) *Renderer {
	r := &Renderer{
		images: make(map[string]*imageState),
	}
	for _, o := range opts {
		o(&r.opts)
	}
	if r.opts.Backend == nil {
		r.opts.Backend = Detect()
	}
	r.backend = r.opts.Backend
	return r
}

type RendererOption func(*RendererOptions)

func WithBackend(b Backend) RendererOption {
	return func(o *RendererOptions) {
		o.Backend = b
	}
}

func (r *Renderer) Display(path string, opts ...Option) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	img, err := Open(path)
	if err != nil {
		return fmt.Errorf("open image %s: %w", path, err)
	}

	var o Options
	for _, opt := range opts {
		opt(&o)
	}

	if _, ok := r.images[path]; !ok {
		r.images[path] = &imageState{img: img}
	}
	r.images[path].opts = o

	return r.backend.Display(path, img, o)
}

func (r *Renderer) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.backend.Clear()
}

func (r *Renderer) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.backend.Close()
}

func (r *Renderer) Backend() Backend {
	return r.backend
}

func (r *Renderer) Name() string {
	return r.backend.Name()
}
