// Package cards provides generators for social media summary cards.
package cards

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/arran4/header"
	"github.com/golang/freetype/truetype"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

func init() {
	Register(&BasicCard{})
}

// BasicCard represents the basic card generator
type BasicCard struct{}

// Name returns the name of the basic card generator
func (b *BasicCard) Name() string {
	return "basic"
}

// Description returns a description of the basic card generator
func (b *BasicCard) Description() string {
	return "A basic card containing the title, description, cover image, favicon and a QR code"
}

// Generate creates a basic card image from the given PageInfo
func (b *BasicCard) Generate(info *header.PageInfo) (image.Image, error) {
	width, height := 800, 400
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 245, G: 245, B: 245, A: 255}}, image.Point{}, draw.Src)

	// Fetch Favicon
	var faviconImg image.Image
	if len(info.Icons) > 0 {
		faviconURL := info.Icons[0].URL
		if !strings.HasPrefix(faviconURL, "http://") && !strings.HasPrefix(faviconURL, "https://") {
			base, err := url.Parse(info.URL)
			if err == nil {
				parsedURL, err := url.Parse(faviconURL)
				if err == nil {
					faviconURL = base.ResolveReference(parsedURL).String()
				}
			}
		}

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(faviconURL)
		if err == nil {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode == http.StatusOK {
				decoded, _, err := image.Decode(resp.Body)
				if err == nil {
					faviconImg = decoded
				}
			}
		}
	}
	if faviconImg == nil {
		faviconImg = createPlaceholderImage(16, 16, color.RGBA{150, 150, 150, 255})
	}

	// Fetch QR Code
	var qrImg image.Image
	qrCode, err := qrcode.Encode(info.URL, qrcode.Medium, 256)
	if err == nil {
		qrDecoded, _, err := image.Decode(bytes.NewReader(qrCode))
		if err == nil {
			qrImg = qrDecoded
		}
	}

	// Try to get an image
	var coverImg image.Image
	images := info.CoverImages()
	if len(images) > 0 {
		imgURL := images[0]

		// Handle relative URLs
		if !strings.HasPrefix(imgURL, "http://") && !strings.HasPrefix(imgURL, "https://") {
			base, err := url.Parse(info.URL)
			if err == nil {
				parsedURL, err := url.Parse(imgURL)
				if err == nil {
					imgURL = base.ResolveReference(parsedURL).String()
				}
			}
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(imgURL)
		if err == nil {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode == http.StatusOK {
				decoded, _, err := image.Decode(resp.Body)
				if err == nil {
					coverImg = decoded
				}
			}
		}
	}
	if coverImg == nil {
		coverImg = createPlaceholderImage(380, 380, color.RGBA{200, 200, 200, 255})
	}

	// Draw cover image (scaled to fit left half)
	imgRect := image.Rect(20, 20, 380, 380)
	scaleImage(img, imgRect, coverImg)

	// Draw texts
	textX := 400
	textY := 60

	title := info.Title
	if title == "" {
		title = info.OGTitle
	}
	if title == "" {
		title = info.TwitterTitle
	}
	if title == "" {
		title = "No Title"
	}

	desc := info.OGDescription
	if desc == "" {
		desc = info.TwitterDescription
	}

	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing font: %v\n", err)
	}
	var fce font.Face
	if f != nil {
		fce = truetype.NewFace(f, &truetype.Options{Size: 20})
	}

	titleEnd := drawTextWrap(img, textX, textY, "Title: "+title, color.RGBA{0, 0, 0, 255}, width-20, fce)

	yOff := 30
	if fce != nil {
		yOff = 45
	}
	drawTextWrap(img, textX, titleEnd+yOff, "Description: "+desc, color.RGBA{100, 100, 100, 255}, width-20, fce)

	// Draw QR Code in bottom right
	if qrImg != nil {
		qrRect := image.Rect(width-100, height-100, width-20, height-20)
		scaleImage(img, qrRect, qrImg)
	}

	// Draw Footer/Header (URL + Favicon)
	footerY := height - 20
	footerX := 20

	if faviconImg != nil {
		favRect := image.Rect(footerX, footerY-16, footerX+16, footerY)
		scaleImage(img, favRect, faviconImg)
		footerX += 25
	}

	fceSmall := fce
	if f != nil {
		fceSmall = truetype.NewFace(f, &truetype.Options{Size: 14})
	}

	drawTextWrap(img, footerX, footerY, info.URL, color.RGBA{100, 100, 100, 255}, width-120, fceSmall)

	return img, nil
}

func createPlaceholderImage(w, h int, color color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color}, image.Point{}, draw.Src)
	return img
}

func scaleImage(dst draw.Image, r image.Rectangle, src image.Image) {
	// Simple scaling
	bounds := src.Bounds()
	dx := bounds.Dx()
	dy := bounds.Dy()
	if dx == 0 || dy == 0 {
		return
	}

	// Aspect ratio preserving scale
	srcRatio := float64(dx) / float64(dy)
	dstRatio := float64(r.Dx()) / float64(r.Dy())

	var finalRect image.Rectangle
	if srcRatio > dstRatio {
		// wider
		h := int(float64(r.Dx()) / srcRatio)
		if h == 0 { h = 1 } // prevent division by zero
		y := r.Min.Y + (r.Dy()-h)/2
		finalRect = image.Rect(r.Min.X, y, r.Max.X, y+h)
	} else {
		// taller
		w := int(float64(r.Dy()) * srcRatio)
		if w == 0 { w = 1 } // prevent division by zero
		x := r.Min.X + (r.Dx()-w)/2
		finalRect = image.Rect(x, r.Min.Y, x+w, r.Max.Y)
	}

	if finalRect.Dx() == 0 || finalRect.Dy() == 0 {
		return
	}

	for y := finalRect.Min.Y; y < finalRect.Max.Y; y++ {
		sy := bounds.Min.Y + (y-finalRect.Min.Y)*dy/finalRect.Dy()
		for x := finalRect.Min.X; x < finalRect.Max.X; x++ {
			sx := bounds.Min.X + (x-finalRect.Min.X)*dx/finalRect.Dx()
			dst.Set(x, y, src.At(sx, sy))
		}
	}
}

func drawTextWrap(img draw.Image, x, y int, label string, col color.Color, maxX int, fce font.Face) int {
	if fce == nil {
		// fallback to basic font without real wordwrap package if font failed
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(col),
			Face: basicfont.Face7x13,
		}

		words := strings.Split(label, " ")
		line := ""
		currentY := y
		for _, word := range words {
			testLine := line
			if testLine != "" { testLine += " " }
			testLine += word
			d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
			width := d.MeasureString(testLine).Ceil()
			if x + width > maxX && line != "" {
				d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
				d.DrawString(line)
				line = word
				currentY += 15
			} else {
				line = testLine
			}
		}
		if line != "" {
			d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
			d.DrawString(line)
		}
		return currentY
	}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: fce,
	}

	// Use our own simple basic wordwrap but with the true type font
	words := strings.Split(label, " ")
	line := ""
	currentY := y
	height := int(fce.Metrics().Height.Ceil())

	for _, word := range words {
		testLine := line
		if testLine != "" { testLine += " " }
		testLine += word
		d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
		width := d.MeasureString(testLine).Ceil()
		if x + width > maxX && line != "" {
			d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
			d.DrawString(line)
			line = word
			currentY += height + 5
		} else {
			line = testLine
		}
	}
	if line != "" {
		d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
		d.DrawString(line)
	}
	return currentY
}
