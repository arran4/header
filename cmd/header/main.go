package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: header <command> [args...]\n")
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  extract   Extract metadata from URL (default if no matching command is found)\n")
		fmt.Fprintf(os.Stderr, "  edit      Edit HTML file head/meta fields\n")
		os.Exit(2)
	}

	command := os.Args[1]

	switch command {
	case "edit":
		os.Exit(editCmd(os.Args[2:]))
	case "extract":
		os.Exit(extractCmd(os.Args[2:]))
	default:
		// To remain somewhat backwards compatible with old usage where people did `header https://...`,
		// if the command isn't recognized, fallback to extract treating the command as an arg.
		// e.g. `header -test https://example.com`
		os.Exit(extractCmd(os.Args[1:]))
	}
}
