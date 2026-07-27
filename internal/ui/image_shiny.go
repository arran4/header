//go:build shiny
// +build shiny

package ui

import (
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"io"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
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

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Get(urlStr)
	if err != nil {
		fmt.Printf("Failed to download image: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to download image: HTTP %s\n", resp.Status)
		return
	}

	if viewer == "export" || viewer == "open" || viewer == "xdg-open" || viewer == "gio" || viewer == "gnome-open" || viewer == "kde-open" {
		// Needs temp file for these
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
		if viewer == "open" {
			cmd = exec.Command("open", finalName)
		} else if viewer == "xdg-open" {
			cmd = exec.Command("xdg-open", finalName)
		} else if viewer == "gio" {
			cmd = exec.Command("gio", "open", finalName)
		} else if viewer == "gnome-open" {
			cmd = exec.Command("gnome-open", finalName)
		} else if viewer == "kde-open" {
			cmd = exec.Command("kde-open", finalName)
		}

		if cmd != nil {
			err = cmd.Start()
			if err != nil {
				fmt.Printf("Failed to open image viewer: %v\n", err)
			}
		}
		return
	}

	envViewer := os.Getenv("IMAGE_VIEWER")
	if (viewer == "" || viewer == "default") && envViewer != "" {
		// Needs temp file for these
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

		cmd := exec.Command(envViewer, finalName)
		if err := cmd.Start(); err != nil {
			fmt.Printf("Failed to open image viewer: %v\n", err)
		}
		return
	}

	if viewer == "" || viewer == "default" {
		fmt.Printf("Displaying image (shiny UI): %s\n", urlStr)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		return
	}

	driver.Main(func(s screen.Screen) {
		bounds := img.Bounds()
		w, err := s.NewWindow(&screen.NewWindowOptions{
			Title:  "Header Cover Image",
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		})
		if err != nil {
			fmt.Printf("Failed to create window: %v\n", err)
			return
		}
		defer w.Release()

		b, err := s.NewBuffer(bounds.Max)
		if err != nil {
			fmt.Printf("Failed to create buffer: %v\n", err)
			return
		}
		defer b.Release()

		draw.Copy(b.RGBA(), image.Point{}, img, bounds, draw.Src, nil)

		for {
			switch e := w.NextEvent().(type) {
			case lifecycle.Event:
				if e.To == lifecycle.StageDead {
					return
				}
			case key.Event:
				if e.Code == key.CodeEscape {
					return
				}
			case paint.Event:
				w.Upload(image.Point{}, b, bounds)
				w.Publish()
			case error:
				fmt.Printf("Error: %v\n", e)
			}
		}
	})
}
