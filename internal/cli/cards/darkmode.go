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
	"golang.org/x/image/font/gofont/goregular"
)

func init() {
	Register(&DarkmodeCard{})
}

// DarkmodeCard represents the basic card generator
type DarkmodeCard struct{}

func (b *DarkmodeCard) Name() string {
	return "darkmode"
}

func (b *DarkmodeCard) Description() string {
	return "A darkmode card containing the title, description, cover image, favicon and a QR code"
}

func (b *DarkmodeCard) Generate(info *header.PageInfo) (image.Image, error) {
	width, height := 800, 400
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{R: 30, G: 30, B: 30, A: 255}}, image.Point{}, draw.Src)

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
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				decoded, _, err := image.Decode(resp.Body)
				if err == nil {
					faviconImg = decoded
				}
			}
		}
	}
	if faviconImg == nil {
		faviconImg = createPlaceholderImage(16, 16, color.RGBA{50, 50, 50, 255})
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
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				decoded, _, err := image.Decode(resp.Body)
				if err == nil {
					coverImg = decoded
				}
			}
		}
	}
	if coverImg == nil {
		coverImg = createPlaceholderImage(380, 380, color.RGBA{50, 50, 50, 255})
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

	drawTextWrap(img, textX, textY, "Title: "+title, color.RGBA{255, 255, 255, 255}, width-20, fce)

	yOff := 30
	if fce != nil {
		yOff = 45
	}
	drawTextWrap(img, textX, textY+yOff, "Description: "+desc, color.RGBA{200, 200, 200, 255}, width-20, fce)

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

	drawTextWrap(img, footerX, footerY, info.URL, color.RGBA{200, 200, 200, 255}, width-120, fceSmall)

	return img, nil
}
