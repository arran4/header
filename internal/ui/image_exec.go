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

func DisplayImage(urlStr string) {
	fmt.Printf("Displaying image (exec fallback): %s\n", urlStr)

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
	// Depending on OS we may need to leave the file around for a moment for the viewer to open it

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		fmt.Printf("Failed to save image: %v\n", err)
		return
	}

	// Make sure we have the right extension for Windows mostly
	finalName := tmpName
	if filepath.Ext(tmpName) == "" {
		finalName = tmpName + ".jpg"
		if err := os.Rename(tmpName, finalName); err != nil {
			fmt.Printf("Failed to rename temp file: %v\n", err)
			return
		}
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", finalName)
	case "darwin":
		cmd = exec.Command("open", finalName)
	default:
		cmd = exec.Command("xdg-open", finalName)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to open image viewer: %v\n", err)
	}
}
