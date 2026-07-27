package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/internal/ui"
)

func ViewCmd(args []string) int {
	fs := flag.NewFlagSet("header view", flag.ContinueOnError)
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	pargs := fs.Args()
	if len(pargs) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: header view <type> <url>\n")
		return 2
	}

	viewType := pargs[0]
	urlStr := pargs[1]

	opts := header.FetchOptions{
		UserAgent: *userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	var targetURL string

	switch viewType {
	case "image":
		images := info.CoverImages()
		if len(images) > 0 {
			targetURL = images[0]
		}
	case "og-image":
		if info.OGImage != "" {
			targetURL = info.OGImage
		} else if info.OGImageSecureURL != "" {
			targetURL = info.OGImageSecureURL
		}
	case "twitter-image":
		targetURL = info.TwitterImage
	case "favicon", "icon":
		for _, icon := range info.Icons {
			targetURL = icon.URL
			break
		}
	default:
		// Attempt to match dynamically, e.g., icon-32, favicon-32
		if strings.HasPrefix(viewType, "icon-") || strings.HasPrefix(viewType, "favicon-") {
			parts := strings.Split(viewType, "-")
			if len(parts) == 2 {
				targetSize := parts[1] + "x" + parts[1]
				for _, icon := range info.Icons {
					if strings.Contains(icon.Sizes, targetSize) {
						targetURL = icon.URL
						break
					}
				}
			}
		}
	}

	if targetURL == "" {
		fmt.Fprintf(os.Stderr, "No image found for type '%s'.\n", viewType)
		return 1
	}

	ui.DisplayImage(targetURL)
	return 0
}
