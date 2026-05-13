package goberzurg

type Pos struct {
	X, Y int
}

type Size struct {
	Width, Height int
}

type Options struct {
	Pos
	Size
	ZIndex int
}

type Option func(*Options)

func WithPos(x, y int) Option {
	return func(o *Options) {
		o.X = x
		o.Y = y
	}
}

func WithSize(w, h int) Option {
	return func(o *Options) {
		o.Width = w
		o.Height = h
	}
}

func WithZIndex(z int) Option {
	return func(o *Options) {
		o.ZIndex = z
	}
}
