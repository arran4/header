// Package ui provides image display utilities.
package ui

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/arran4/header"
)

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// DisplayImage displays an image from the given URL using the specified viewer.
func DisplayImage(urlStr string, viewer string) {
	DisplayImageWithOptions(urlStr, viewer, header.FetchOptions{})
}

// DisplayImageWithOptions displays an image from the given URL using the specified viewer and fetch options.
func DisplayImageWithOptions(urlStr string, viewer string, opts header.FetchOptions) {
	if viewer == "print-url" {
		fmt.Println(urlStr)
		return
	}

	if viewer != "" && viewer != "default" && viewer != "shiny" && viewer != "open" && viewer != "xdg-open" && viewer != "gio" && viewer != "gnome-open" && viewer != "kde-open" && viewer != "export" && viewer != "mock" {
		fmt.Printf("Error: unsupported viewer %q requested\n", viewer)
		return
	}

	if (viewer == "" || viewer == "default") && !HasShiny() {
		fmt.Printf("Displaying image (exec fallback): %s\n", urlStr)
	}

	timeout := opts.ImageTimeout
	if timeout <= 0 {
		timeout = opts.Timeout
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	}

	policy := header.RetryPolicy{
		Attempts: opts.RetryAttempts,
		Delay:    opts.RetryDelay,
		MaxDelay: opts.RetryMaxDelay,
	}

	resp, _, err := header.ExecuteRequestWithRetry(ctx, client, req, policy)
	if err != nil {
		fmt.Printf("Failed to download image: %v\n", header.SanitizeError(err))
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to download image: HTTP %s\n", resp.Status)
		return
	}

	if viewer == "mock" {
		return
	}

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 20*1024*1024))
	if err != nil {
		fmt.Printf("Failed to read image: %v\n", err)
		return
	}

	// Try Shiny first if requested or default (and shiny is available)
	if (viewer == "shiny" || viewer == "" || viewer == "default") && HasShiny() {
		if viewer == "" || viewer == "default" {
			fmt.Printf("Displaying image (shiny UI): %s\n", urlStr)
		}
		if ShowShiny(bytes.NewReader(bodyBytes), urlStr) {
			return
		}
		// If ShowShiny returns false (e.g., error decoding), we can fall back to the rest
	} else if viewer == "shiny" && !HasShiny() {
		fmt.Println("Warning: shiny viewer requested but not compiled with shiny tags, falling back to default")
		viewer = "default"
	}

	tmpFile, err := os.CreateTemp("", "header-cover-*.jpg")
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}
	tmpName := tmpFile.Name()
	defer func() { _ = tmpFile.Close() }()

	_, err = tmpFile.Write(bodyBytes)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
		return
	}

	finalName := tmpName
	if filepath.Ext(tmpName) == "" {
		finalName = tmpName + ".jpg"
		if err := os.Rename(tmpName, finalName); err != nil {
			fmt.Printf("Failed to rename temp file: %v\n", err)
			return
		}
	}

	if viewer == "export" {
		fmt.Printf("Image exported to: %s\n", finalName)
		return
	}

	var cmd *exec.Cmd

	if viewer == "" || viewer == "default" {
		if envViewer := os.Getenv("IMAGE_VIEWER"); envViewer != "" {
			cmd = exec.Command(envViewer, finalName)
		} else {
			switch runtime.GOOS {
			case "windows":
				cmd = exec.Command("cmd", "/c", "start", "", finalName)
			case "darwin":
				cmd = exec.Command("open", finalName)
			default:
				if isCommandAvailable("xdg-open") {
					cmd = exec.Command("xdg-open", finalName)
				} else if isCommandAvailable("gio") {
					cmd = exec.Command("gio", "open", finalName)
				} else if isCommandAvailable("gnome-open") {
					cmd = exec.Command("gnome-open", finalName)
				} else if isCommandAvailable("kde-open") {
					cmd = exec.Command("kde-open", finalName)
				} else {
					fmt.Println("Warning: no default viewer found (e.g. xdg-open)")
					return
				}
			}
		}
	} else {
		switch viewer {
		case "open", "xdg-open", "gnome-open", "kde-open":
			cmd = exec.Command(viewer, finalName)
		case "gio":
			cmd = exec.Command("gio", "open", finalName)
		}
	}

	if cmd != nil {
		err = cmd.Start()
		if err != nil {
			fmt.Printf("Failed to open image viewer: %v\n", err)
		}
	}
}
