package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/bakedsnake/goberzurg"
)

type command struct {
	Action string `json:"action"`
	Path   string `json:"path"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Z      int    `json:"z"`
}

const usageText = `goberzurg — display images in the terminal

Usage:
  goberzurg [flags] <path> [x] [y] [width] [height]
  goberzurg [flags]              (reads JSON/text commands from stdin)

Positional arguments:
  path                   Path to image file
  x                      Column position (default: 0)
  y                      Row position (default: 0)
  width                  Width in character cells (default: 0 = auto)
  height                 Height in character cells (default: 0 = auto)

Flags:`

func main() {
	listBackends := flag.Bool("list-backends", false, "List available backends and exit")
	useBackend := flag.String("backend", "", "Force a specific backend (kitty, sixel, iterm2)")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usageText)
		fmt.Fprint(os.Stderr, "\n")
		flag.PrintDefaults()
		fmt.Fprint(os.Stderr, "\nStdin protocol (JSON or text):\n")
		fmt.Fprint(os.Stderr, "  {\"action\":\"add\",\"path\":\"img.png\",\"x\":5,\"y\":2,\"width\":40,\"height\":30}\n")
		fmt.Fprint(os.Stderr, "  add img.png 5 2 40 30\n")
		fmt.Fprint(os.Stderr, "  clear\n")
		fmt.Fprint(os.Stderr, "  quit\n")
	}
	flag.Parse()

	if *listBackends {
		fmt.Println("kitty")
		fmt.Println("sixel")
		fmt.Println("iterm2")
		return
	}

	var r *goberzurg.Renderer
	if *useBackend != "" {
		var b goberzurg.Backend
		switch strings.ToLower(*useBackend) {
		case "kitty":
			b = goberzurg.NewKittyBackend()
		case "sixel":
			b = goberzurg.NewSixelBackend()
		case "iterm2":
			b = goberzurg.NewIterm2Backend()
		default:
			fmt.Fprintf(os.Stderr, "error: unknown backend %q\n", *useBackend)
			fmt.Fprintf(os.Stderr, "available: kitty, sixel, iterm2\n")
			os.Exit(1)
		}
		r = goberzurg.New(goberzurg.WithBackend(b))
	} else {
		r = goberzurg.New()
	}
	defer r.Close()

	if flag.NArg() > 0 {
		args := flag.Args()
		path := args[0]
		x, y, w, h := 0, 0, 0, 0

		if len(args) > 1 {
			x, _ = strconv.Atoi(args[1])
		}
		if len(args) > 2 {
			y, _ = strconv.Atoi(args[2])
		}
		if len(args) > 3 {
			w, _ = strconv.Atoi(args[3])
		}
		if len(args) > 4 {
			h, _ = strconv.Atoi(args[4])
		}

		err := r.Display(path,
			goberzurg.WithPos(x, y),
			goberzurg.WithSize(w, h),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var cmd command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			cmd.Action = parts[0]
			cmd.Path = parts[1]
			if len(parts) > 2 {
				cmd.X, _ = strconv.Atoi(parts[2])
			}
			if len(parts) > 3 {
				cmd.Y, _ = strconv.Atoi(parts[3])
			}
			if len(parts) > 4 {
				cmd.Width, _ = strconv.Atoi(parts[4])
			}
			if len(parts) > 5 {
				cmd.Height, _ = strconv.Atoi(parts[5])
			}
		}

		switch cmd.Action {
		case "add", "display":
			err := r.Display(cmd.Path,
				goberzurg.WithPos(cmd.X, cmd.Y),
				goberzurg.WithSize(cmd.Width, cmd.Height),
				goberzurg.WithZIndex(cmd.Z),
			)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
			}
		case "clear", "remove":
			r.Clear()
		case "quit", "exit":
			return
		}
	}
}
