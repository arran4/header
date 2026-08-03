// Package cli provides the command line interface subcommands.
package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/arran4/header"
	"golang.org/x/net/html"
)

// EditCmd handles the edit subcommands (e.g., editing head or meta tags).
func EditCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  header edit head [flags] <file>\n")
		fmt.Fprintf(os.Stderr, "  header edit meta [flags] <name> <file>\n")
		return 2
	}

	commandType := args[0]

	fs := flag.NewFlagSet("header edit", flag.ContinueOnError)

	setVal := fs.String("s", "", "set (remove other versions if any)")
	addVal := fs.String("S", "", "add to set (add a new one, only works where valid)")
	replaceVal := fs.String("r", "", "replace (only works if replacing 1 value)")
	replaceIndex := fs.Int("n", -1, "replace number x, only works with -r")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing edit flags: %v\n", err)
		return 2
	}

	var hasS, hasSUpper, hasR bool
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "s":
			hasS = true
		case "S":
			hasSUpper = true
		case "r":
			hasR = true
		}
	})

	opts := header.EditOptions{
		CommandType:  commandType,
		ReplaceIndex: *replaceIndex,
	}

	if hasS {
		opts.SetVal = setVal
	}
	if hasSUpper {
		opts.AddVal = addVal
	}
	if hasR {
		opts.ReplaceVal = replaceVal
	}

	parsedArgs := fs.Args()
	var targetFile string
	var metaName string

	switch commandType {
	case "head":
		if len(parsedArgs) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: header edit head [flags] <file>\n")
			return 2
		}
		targetFile = parsedArgs[0]
	case "meta":
		if len(parsedArgs) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: header edit meta [flags] <name> <file>\n")
			return 2
		}
		metaName = parsedArgs[0]
		targetFile = parsedArgs[1]
	default:
		fmt.Fprintf(os.Stderr, "Unknown edit type: %s\n", commandType)
		return 2
	}

	opts.MetaName = metaName

	var inputData []byte
	var err error
	if targetFile == "-" {
		inputData, err = io.ReadAll(os.Stdin)
	} else {
		inputData, err = os.ReadFile(targetFile)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		return 1
	}

	doc, err := header.ApplyEdit(bytes.NewReader(inputData), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error applying edit: %v\n", err)
		return 1
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering HTML: %v\n", err)
		return 1
	}

	if targetFile == "-" {
		if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			return 1
		}
	} else {
		if err := os.WriteFile(targetFile, buf.Bytes(), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			return 1
		}
	}

	return 0
}
