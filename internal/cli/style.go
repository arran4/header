package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
// By default without --force, exports are created atomically and exclusively; existing files,
// directories, and symlinks (including dangling symlinks) are never overwritten.
// When --force is specified, the destination is safely replaced via an atomic temporary file rename.
// If the destination is an existing symlink, --force replaces the symlink itself and does not follow
// or overwrite the symlink target. Exporting to an existing directory is always rejected.
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
		if fi, err := os.Lstat(out); err == nil {
			if fi.IsDir() {
				return fmt.Errorf("cannot export to %q: destination is a directory", out)
			}
			return fmt.Errorf("file %q already exists (use --force to overwrite)", out)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("error checking output file %q: %w", out, err)
		}

		f, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				return fmt.Errorf("file %q already exists (use --force to overwrite)", out)
			}
			return fmt.Errorf("error creating output file %q: %w", out, err)
		}

		if _, err := f.Write(content); err != nil {
			_ = f.Close()
			_ = os.Remove(out)
			return fmt.Errorf("error writing template to %q: %w", out, err)
		}
		if err := f.Close(); err != nil {
			_ = os.Remove(out)
			return fmt.Errorf("error closing output file %q: %w", out, err)
		}

		return nil
	}

	if fi, err := os.Lstat(out); err == nil && fi.IsDir() {
		return fmt.Errorf("cannot export to %q: destination is a directory", out)
	}

	dir := filepath.Dir(out)
	tmpFile, err := os.CreateTemp(dir, ".export-*.tmp")
	if err != nil {
		return fmt.Errorf("error creating temporary file in %q: %w", dir, err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if err := os.Chmod(tmpName, 0644); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("error setting permissions on %q: %w", tmpName, err)
	}

	if _, err := tmpFile.Write(content); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("error writing template to temporary file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}

	if err := os.Rename(tmpName, out); err != nil {
		return fmt.Errorf("error writing template to %q: %w", out, err)
	}

	return nil
}

// Styles is a subcommand `header styles` -- List available extraction styles.
func Styles() error {
	return StyleList()
}
