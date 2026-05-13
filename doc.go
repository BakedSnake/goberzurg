// Package goberzurg displays images in terminal emulators.
//
// It provides multiple backends using terminal graphics protocols:
//   - Kitty protocol (kitty, wezterm, ghostty)
//   - Sixel (xterm, mlterm, foot)
//   - iTerm2 inline images (iTerm2)
//
// The backends write directly to the terminal via os.Stdout, making
// them compatible with bubbletea programs.
//
// Basic usage:
//
//	r := goberzurg.New()
//	defer r.Close()
//	r.Display("image.png", goberzurg.WithPos(5, 2))
//
// To force a specific backend:
//
//	r := goberzurg.New(goberzurg.WithBackend(goberzurg.NewKittyBackend()))
package goberzurg
