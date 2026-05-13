package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"goberzurg"
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

func main() {
	listBackends := flag.Bool("list-backends", false, "List available backends")
	useBackend := flag.String("backend", "", "Force a specific backend (kitty, sixel, iterm2)")
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
			fmt.Fprintf(os.Stderr, "unknown backend: %s\n", *useBackend)
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
		x := 0
		y := 0
		w := 0
		h := 0

		if n := len(args); n > 0 {
			path = args[0]
		}
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
				_ = 0
			}
		case "clear", "remove":
			r.Clear()
		case "quit", "exit":
			return
		}
	}
}
