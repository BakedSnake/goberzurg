package goberzurg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var IsTmux bool

func init() {
	IsTmux = os.Getenv("TMUX") != ""
}

// tmuxWrite wraps seq in the tmux DCS passthrough wrapper if IsTmux is true.
//
// The wrapping format is:
//
//	ESC P t m u x ; ESC <escaped_seq> ESC \
//
// Inside the DCS, every ESC (0x1b) is doubled to ESC ESC. For sequences that
// end with ESC \ (ST), an extra backslash is appended after doubling. This
// ensures the complete sequence (including its terminating ST) reaches the
// outer terminal, even though tmux consumes the ST that ends the DCS wrapper
// and also swallows any ESC \ that follows outside the DCS.
//
// Example (Kitty transmit):
//
//	Original:  ESC _ G ... ESC \
//	Escaped:   ESC ESC _ G ... ESC ESC \ \
//	Wrapped:   ESC P t m u x ; ESC ESC ESC _ G ... ESC ESC \ \ ESC \
//
//	DCS content:   ESC ESC _ G ... ESC   (up to first ESC \)
//	DCS terminator: ESC \                 (consumes the \ of the original ST)
//	Remaining:      \                    (extra backslash → outer terminal)
//	Outer terminal: ESC _ G ... ESC \    (complete original sequence)
func tmuxWrite(w io.Writer, seq string) {
	if !IsTmux {
		fmt.Fprint(w, seq)
		return
	}

	escaped := strings.ReplaceAll(seq, "\x1b", "\x1b\x1b")
	if strings.HasSuffix(seq, "\x1b\\") {
		escaped += "\\"
	}

	fmt.Fprint(w, "\x1bPtmux;\x1b")
	fmt.Fprint(w, escaped)
	fmt.Fprint(w, "\x1b\\")
}

// ClearOnResetWriter wraps an io.Writer and calls clearFn when it detects
// a clear-screen escape sequence (\x1b[2J or \x1b[3J) being written.
// This ensures placed images are cleared when the terminal is cleared.
type ClearOnResetWriter struct {
	w       io.Writer
	clearFn func() error
	mu      sync.Mutex
	clearing bool
}

// NewClearOnResetWriter creates a writer that auto-clears images on terminal clear.
func NewClearOnResetWriter(w io.Writer, clearFn func() error) *ClearOnResetWriter {
	return &ClearOnResetWriter{w: w, clearFn: clearFn}
}

func (c *ClearOnResetWriter) Write(p []byte) (int, error) {
	c.mu.Lock()
	if !c.clearing && (bytes.Contains(p, []byte("\x1b[2J")) || bytes.Contains(p, []byte("\x1b[3J"))) {
		c.clearing = true
		c.mu.Unlock()
		c.clearFn()
		c.mu.Lock()
		c.clearing = false
	}
	c.mu.Unlock()
	return c.w.Write(p)
}
