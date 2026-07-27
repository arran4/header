//go:build !shiny
// +build !shiny

package ui

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func DisplayImage(urlStr string, viewer string) {
	if viewer == "print-url" {
		fmt.Println(urlStr)
		return
	}

	if viewer == "mock" {
		return
	}

	if viewer != "" && viewer != "default" && viewer != "shiny" && viewer != "open" && viewer != "xdg-open" && viewer != "gio" && viewer != "gnome-open" && viewer != "kde-open" && viewer != "export" {
		fmt.Printf("Error: unsupported viewer %q requested\n", viewer)
		return
	}

	if viewer == "" || viewer == "default" {
		fmt.Printf("Displaying image (exec fallback): %s\n", urlStr)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		fmt.Printf("Failed to download image: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to download image: HTTP %s\n", resp.Status)
		return
	}

	tmpFile, err := os.CreateTemp("", "header-cover-*.jpg")
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}
	tmpName := tmpFile.Name()
	defer func() { _ = tmpFile.Close() }()

	_, err = io.Copy(tmpFile, resp.Body)
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

	if viewer == "shiny" {
		fmt.Println("Warning: shiny viewer requested but not compiled with shiny tags, falling back to default")
		viewer = "default"
	}

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
