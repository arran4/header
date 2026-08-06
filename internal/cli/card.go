package cli

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"bytes"
	_ "image/jpeg"
	"image/png"
	_ "golang.org/x/image/webp"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/arran4/header"
	"github.com/skip2/go-qrcode"
	"golang.design/x/clipboard"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// CardCmd handles the card subcommand.
func CardCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: header card <subcommand> [flags] <url>\n")
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		fmt.Fprintf(os.Stderr, "  file        Output destination: file path\n")
		fmt.Fprintf(os.Stderr, "  clipboard   Output destination: copy to system clipboard\n")
		fmt.Fprintf(os.Stderr, "  stdout      Output destination: standard output (pipe)\n")
		return 2
	}

	subcommand := args[0]
	fs := flag.NewFlagSet("header card "+subcommand, flag.ContinueOnError)
	userAgent := fs.String("user-agent", "Mozilla/5.0", "The User-Agent header to use when fetching URLs")

	var outputFile *string
	if subcommand == "file" {
		outputFile = fs.String("out", "", "Output file path (required)")
	}

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	pargs := fs.Args()
	if len(pargs) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: header card %s [flags] <url>\n", subcommand)
		fs.PrintDefaults()
		return 2
	}

	if subcommand == "file" && *outputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -out flag is required for file subcommand\n")
		return 2
	}

	urlStr := pargs[0]

	opts := header.FetchOptions{
		UserAgent: *userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	cardImg := createCardImage(info)

	switch subcommand {
	case "clipboard":
		err = clipboard.Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing clipboard: %v\n", err)
			return 1
		}
		var buf bytes.Buffer
		err = png.Encode(&buf, cardImg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding png: %v\n", err)
			return 1
		}
		clipboard.Write(clipboard.FmtImage, buf.Bytes())
		fmt.Println("Card copied to clipboard.")
	case "stdout":
		err = png.Encode(os.Stdout, cardImg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			return 1
		}
	case "file":
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			return 1
		}
		defer f.Close()
		err = png.Encode(f, cardImg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding png: %v\n", err)
			return 1
		}
		fmt.Printf("Card written to %s\n", *outputFile)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		return 2
	}

	return 0
}

func createCardImage(info *header.PageInfo) image.Image {
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
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				decoded, _, err := image.Decode(resp.Body)
				if err == nil {
					faviconImg = decoded
				}
			}
		}
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

	// Draw cover image (scaled to fit left half)
	if coverImg != nil {
		imgRect := image.Rect(20, 20, 380, 380)
		scaleImage(img, imgRect, coverImg)
	}

	// Draw texts
	textX := 400
	if coverImg == nil {
		textX = 20
	}
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

	drawText(img, textX, textY, "Title: "+title, color.RGBA{0, 0, 0, 255}, width-20)
	drawText(img, textX, textY+30, "Description: "+desc, color.RGBA{100, 100, 100, 255}, width-20)

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

	drawText(img, footerX, footerY-4, info.URL, color.RGBA{100, 100, 100, 255}, width-120)

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

func drawText(img draw.Image, x, y int, label string, col color.Color, maxX int) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
	}

	// basic word wrap
	words := strings.Split(label, " ")
	line := ""
	currentY := y

	for _, word := range words {
		testLine := line
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		d.Dot = fixed.Point26_6{X: fixed.Int26_6(x * 64), Y: fixed.Int26_6(currentY * 64)}
		width := d.MeasureString(testLine).Ceil()

		if x + width > maxX && line != "" {
			// Draw previous line
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
}
