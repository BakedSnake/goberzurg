# goberzurg

A Go library for displaying images in terminal emulators, inspired by [ueberzug](https://github.com/ueberzug/ueberzug). Uses terminal graphics protocols :: no X11/Wayland overlays, no CGo, no external dependencies.

## Features

- **Kitty protocol** (kitty, wezterm, ghostty)
- **Sixel** (xterm, mlterm, foot)
- **iTerm2 inline images** (iTerm2)
- **Auto-detection** :: picks the right backend from `$TERM`, `$TERM_PROGRAM`, `$KITTY_WINDOW_ID`
- **Programmatic image scaling** :: resizing works identically across all backends (nearest-neighbor)
- **tmux support** :: auto-detects tmux and wraps escape sequences in DCS passthrough
- **bubbletea compatible** :: backends write directly to `os.Stdout`
- **Zero external dependencies** :: pure Go standard library

Requires Go 1.26.2 or later.

## Build

### CLI

```sh
go build -o goberzurg ./cmd/goberzurg/
```

### Tests

```sh
go test -v ./...
go test -race ./...
```

### Makefile

```sh
make lib         # build the library
make cli         # build the CLI binary
make test        # run tests with race detector
make install     # install CLI + man pages (PREFIX=/usr/local)
make clean       # remove built binary
```

Install to a custom prefix:

```sh
make install PREFIX=$HOME/.local
```

## Usage (library)

```go
package main

import (
    "github.com/bakedsnake/goberzurg"
)

func main() {
    r := goberzurg.New()
    defer r.Close()

    r.Display("image.png",
        goberzurg.WithPos(5, 2),       // column, row
        goberzurg.WithSize(40, 30),    // width, height in character cells
    )
}
```

Force a specific backend:

```go
r := goberzurg.New(goberzurg.WithBackend(goberzurg.NewKittyBackend()))
```

### bubbletea integration

```go
import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/bakedsnake/goberzurg"
)

type model struct {
    renderer *goberzurg.Renderer
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "q" {
            return m, tea.Quit
        }
    }
    return m, nil
}

func (m model) View() string {
    m.renderer.Display("photo.png", goberzurg.WithPos(0, 0))
    return "Hello\n"
}

func main() {
    r := goberzurg.New()
    defer r.Close()
    tea.NewProgram(model{renderer: r}).Run()
}
```

## Usage (CLI)

Display an image at position (column, row):

```sh
goberzurg image.png 5 2
```

With width and height (in character cells):

```sh
goberzurg image.png 5 2 40 30
```

List available backends:

```sh
goberzurg --list-backends
```

Force a backend:

```sh
goberzurg --backend kitty image.png 5 2
```

### stdin protocol (JSON)

```sh
echo '{"action":"add","path":"image.png","x":5,"y":2,"width":40,"height":30,"z":1}' | goberzurg
```

Simple text format also works:

```sh
echo 'add image.png 5 2 40 30' | goberzurg
```

Commands: `add` / `display`, `clear` / `remove`, `quit` / `exit`. JSON fields: `action`, `path`, `x`, `y`, `width`, `height`, `z`.

## Backends

| Backend | Terminals | Protocol |
|---------|-----------|----------|
| Kitty | kitty, wezterm, ghostty | `\e_G…` escape sequences |
| Sixel | xterm, mlterm, foot, etc. | `\ePq…` sixel graphics |
| iTerm2 | iTerm2 | `\e]1337;File=…` OSC sequence |

Detection priority: `$KITTY_WINDOW_ID` → `$TERM_PROGRAM` → `$TERM` (sixel/xterm keywords) → Kitty (fallback).

Sizing (width, height) is handled programmatically in Go :: images are scaled before being passed to any backend. This means `WithSize(40, 30)` produces the same result on every backend regardless of terminal protocol capabilities.

## Summary

### Source files

| File | Purpose |
|------|---------|
| `goberzurg.go` | Core types (`Pos`, `Size`, `Options`, `Option` funcs) |
| `image.go` | Image loading, format detection, PNG/JPEG encoding, `ScaleToCells()` |
| `renderer.go` | `Backend` interface, `Renderer` high-level API |
| `kitty.go` | Kitty terminal protocol backend |
| `sixel.go` | Sixel graphics backend |
| `iterm2.go` | iTerm2 inline images backend |
| `detect.go` | Automatic terminal detection |
| `tmux.go` | tmux DCS passthrough, `ClearOnResetWriter` |
| `doc.go` | Package documentation |
| `cmd/goberzurg/main.go` | CLI tool |
| `goberzurg_test.go` | Tests |

### Man pages

| Section | File | Content |
|---------|------|---------|
| 1 | `man/goberzurg.1` | CLI usage, options, stdin protocol, backends, examples |
| 3 | `man/goberzurg.3` | Library API documentation (`Renderer`, `Image`, `Open`, `ScaleToCells`, etc.) |

## License

MIT
