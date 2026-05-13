package goberzurg

import (
	"os"
	"strings"
)

const (
	BackendUnknown BackendType = iota
	BackendKitty
	BackendSixel
	BackendIterm2
)

type BackendType int

func Detect() Backend {
	switch detectTerm() {
	case BackendKitty:
		return NewKittyBackend()
	case BackendIterm2:
		return NewIterm2Backend()
	case BackendSixel:
		return NewSixelBackend()
	default:
		return NewKittyBackend()
	}
}

func DetectType() BackendType {
	return detectTerm()
}

func detectTerm() BackendType {
	term := os.Getenv("TERM")
	kittyPID := os.Getenv("KITTY_WINDOW_ID")
	termProgram := os.Getenv("TERM_PROGRAM")

	switch {
	case kittyPID != "":
		return BackendKitty
	case termProgram == "iTerm.app" || termProgram == "iTerm2":
		return BackendIterm2
	case termProgram == "WezTerm" || termProgram == "wezterm":
		return BackendKitty
	case strings.Contains(strings.ToLower(term), "sixel"):
		return BackendSixel
	case termProgram != "":
		return BackendKitty
	case strings.Contains(strings.ToLower(term), "xterm"):
		return BackendSixel
	default:
		return BackendKitty
	}
}
