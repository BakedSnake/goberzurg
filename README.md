# goberzurg

A Go library for displaying images in terminal emulators, inspired by [ueberzug](https://github.com/ueberzug/ueberzug). Uses terminal graphics protocols — no X11/Wayland overlays, no CGo, no external dependencies.

## Features

- **Kitty protocol** (kitty, wezterm, ghostty) — transmit-once, place-many
- **Sixel** (xterm, mlterm, foot)
- **iTerm2 inline images** (iTerm2)
- **Auto-detection** — picks the right backend from `$TERM`, `$TERM_PROGRAM`, `$KITTY_WINDOW_ID`
- **bubbletea compatible** — backends write directly to `os.Stdout`
- **Zero external dependencies** — pure Go standard library

## Build

### Library

```sh
go build ./...
```

### CLI

```sh
go build -o goberzurg ./cmd/goberzurg/
```

### Tests

```sh
go test ./...
```

With race detector:

```sh
go test -race ./...
```

Verbose output:

```sh
go test -v ./...
```

## Usage (library)

```go
package main

import (
    "goberzurg"
)

func main() {
    r := goberzurg.New()
    defer r.Close()

    r.Display("photo.png",
        goberzurg.WithPos(5, 2),
        goberzurg.WithSize(40, 30),
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
    "goberzurg"
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
goberzurg photo.png 5 2
```

With width and height:

```sh
goberzurg photo.png 5 2 40 30
```

List available backends:

```sh
goberzurg --list-backends
```

Force a backend:

```sh
goberzurg --backend kitty photo.png 5 2
```

### stdin protocol (JSON)

```sh
echo '{"action":"add","path":"photo.png","x":5,"y":2}' | goberzurg
```

Simple text format also works:

```sh
echo 'add photo.png 5 2' | goberzurg
```

Commands: `add` / `display`, `clear` / `remove`, `quit` / `exit`.

## Backends

| Backend | Terminals | Protocol |
|---------|-----------|----------|
| Kitty | kitty, wezterm, ghostty | `\e_G…` escape sequences |
| Sixel | xterm, mlterm, foot, etc. | `\ePq…` sixel graphics |
| iTerm2 | iTerm2 | `\e]1337;File=…` OSC sequence |

Detection priority: `$KITTY_WINDOW_ID` → `$TERM_PROGRAM` → `$TERM` (sixel/xterm keywords) → Kitty (fallback).

## Summary

| File | Purpose |
|------|---------|
| `goberzurg.go` | Core types (`Pos`, `Size`, `Options`, `Option` funcs) |
| `image.go` | Image loading, format detection, PNG/JPEG encoding |
| `renderer.go` | `Backend` interface, `Renderer` high-level API |
| `kitty.go` | Kitty terminal protocol backend |
| `sixel.go` | Sixel graphics backend |
| `iterm2.go` | iTerm2 inline images backend |
| `detect.go` | Automatic terminal detection |
| `cmd/goberzurg/main.go` | CLI tool |
| `goberzurg_test.go` | Tests |

## License

MIT
