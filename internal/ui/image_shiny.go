//go:build shiny
// +build shiny

package ui

import (
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
)

func HasShiny() bool {
	return true
}

func ShowShiny(r io.Reader, urlStr string) bool {
	img, _, err := image.Decode(r)
	if err != nil {
		fmt.Printf("Failed to decode image: %v\n", err)
		return false
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

	return true
}
