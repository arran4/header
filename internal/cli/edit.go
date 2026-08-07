// Package cli provides the command line interface subcommands.
package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/arran4/header"
	"golang.org/x/net/html"
)

// EditHead is a subcommand `header edit head` -- Edit HTML file head fields.
//
// Flags:
//   setVal:       -s         set (remove other versions if any)
//   addVal:       -S         add to set (add a new one, only works where valid)
//   replaceVal:   -r         replace (only works if replacing 1 value)
//   replaceIndex: -n (default: -1) replace number x
//   targetFile:   @1 (required)    Target file to edit or - for stdin
func EditHead(setVal *string, addVal *string, replaceVal *string, replaceIndex int, targetFile string) error {
	opts := header.EditOptions{
		CommandType:  "head",
		ReplaceIndex: replaceIndex,
		SetVal:       setVal,
		AddVal:       addVal,
		ReplaceVal:   replaceVal,
	}

	return applyEdit(targetFile, opts)
}

// EditMeta is a subcommand `header edit meta` -- Edit HTML file meta fields.
//
// Flags:
//   setVal:       -s         set (remove other versions if any)
//   addVal:       -S         add to set (add a new one, only works where valid)
//   replaceVal:   -r         replace (only works if replacing 1 value)
//   replaceIndex: -n (default: -1) replace number x
//   metaName:     @1 (required)    Meta name to edit
//   targetFile:   @2 (required)    Target file to edit or - for stdin
func EditMeta(setVal *string, addVal *string, replaceVal *string, replaceIndex int, metaName string, targetFile string) error {
	opts := header.EditOptions{
		CommandType:  "meta",
		ReplaceIndex: replaceIndex,
		SetVal:       setVal,
		AddVal:       addVal,
		ReplaceVal:   replaceVal,
		MetaName:     metaName,
	}

	return applyEdit(targetFile, opts)
}

func applyEdit(targetFile string, opts header.EditOptions) error {
	var inputData []byte
	var err error
	if targetFile == "-" {
		inputData, err = io.ReadAll(os.Stdin)
	} else {
		inputData, err = os.ReadFile(targetFile)
	}
	if err != nil {
		return fmt.Errorf("Error reading input: %w", err)
	}

	doc, err := header.ApplyEdit(bytes.NewReader(inputData), opts)
	if err != nil {
		return fmt.Errorf("Error applying edit: %w", err)
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return fmt.Errorf("Error rendering HTML: %w", err)
	}

	if targetFile == "-" {
		if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
			return fmt.Errorf("Error writing to stdout: %w", err)
		}
	} else {
		if err := os.WriteFile(targetFile, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("Error writing output: %w", err)
		}
	}

	return nil
}
