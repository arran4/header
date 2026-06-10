//go:build shiny
// +build shiny

package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/image/draw"
)

func displayImage(urlStr string) {
	fmt.Printf("Displaying image (shiny UI): %s\n", urlStr)

	resp, err := http.Get(urlStr)
	if err != nil {
		fmt.Printf("Failed to download image: %v\n", err)
		return
	}
	defer resp.Body.Close()

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

		var sz size.Event
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
			case size.Event:
				sz = e
			case paint.Event:
				w.Upload(image.Point{}, b, bounds)
				w.Publish()
			case error:
				fmt.Printf("Error: %v\n", e)
			}
		}
	})
}
