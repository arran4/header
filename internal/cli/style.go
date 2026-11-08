package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/arran4/header/templates"
)

// PrintStyles writes the available extraction styles and their descriptions to w.
func PrintStyles(w io.Writer) error {
	styles, err := templates.ListStyles()
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(w, "Available Styles:"); err != nil {
		return err
	}
	for _, s := range styles {
		if s.Description != "" {
			if _, err := fmt.Fprintf(w, "  %-16s - %s\n", s.Name, s.Description); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "  %s\n", s.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// StyleList is a subcommand `header style list` -- List available extraction styles.
func StyleList() error {
	return PrintStyles(os.Stdout)
}

// StyleExport is a subcommand `header style export` -- Export an extraction style template.
//
// Flags:
//
//	out:   --out                                         Output file path (default is stdout)
//	force: --force                                       Overwrite existing file
//	name:  @1 (required)                                 Name of the style to export
func StyleExport(out string, force bool, name string) error {
	content, err := templates.GetTemplate(name)
	if err != nil {
		return err
	}

	if out == "" {
		if _, err := os.Stdout.Write(content); err != nil {
			return fmt.Errorf("error writing template to stdout: %w", err)
		}
		return nil
	}

	if !force {
		if _, err := os.Stat(out); err == nil {
			return fmt.Errorf("file %q already exists (use --force to overwrite)", out)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error checking output file %q: %w", out, err)
		}
	}

	if err := os.WriteFile(out, content, 0644); err != nil {
		return fmt.Errorf("error writing template to %q: %w", out, err)
	}

	return nil
}

// Styles is a subcommand `header styles` -- List available extraction styles.
func Styles() error {
	return StyleList()
}
