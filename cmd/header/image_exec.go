//go:build !shiny
// +build !shiny

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func displayImage(urlStr string) {
	fmt.Printf("Displaying image (exec fallback): %s\n", urlStr)

	resp, err := http.Get(urlStr)
	if err != nil {
		fmt.Printf("Failed to download image: %v\n", err)
		return
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "header-cover-*.jpg")
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}
	tmpName := tmpFile.Name()
	defer tmpFile.Close()
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
		os.Rename(tmpName, finalName)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", finalName)
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
