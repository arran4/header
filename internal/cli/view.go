package cli

import (
	"fmt"
	"strings"

	"github.com/arran4/header"
	"github.com/arran4/header/internal/ui"
)

// View is a subcommand `header view` -- View images or icons from a URL (e.g. view image <url>).
//
// Flags:
//
//	userAgent:     --user-agent (default: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36") The User-Agent header to use when fetching URLs
//	viewer:        --viewer (default: "default")            The viewer to use (e.g., open, xdg-open, shiny, mock, export, print-url, default)
//	httprettyFlag: --httpretty                              Print HTTP requests and responses
//	viewType:      @1 (required)                            Type of image to view
//	urlStr:        @2 (required)                            URL to fetch
func View(userAgent string, viewer string, httprettyFlag bool, viewType string, urlStr string) error {
	opts := header.FetchOptions{
		UserAgent: userAgent,
		Httpretty: httprettyFlag,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	var targetURL string

	switch viewType {
	case "image", "cover", "cover-image":
		images := info.CoverImages()
		if len(images) > 0 {
			targetURL = images[0]
		}
	case "og-image", "opengraph-image", "open-graph-image", "meta-image", "facebook-image":
		if info.OGImage != "" {
			targetURL = info.OGImage
		} else if info.OGImageSecureURL != "" {
			targetURL = info.OGImageSecureURL
		}
	case "twitter-image", "x-image":
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
		return fmt.Errorf("no image found for type '%s'", viewType)
	}

	ui.DisplayImage(targetURL, viewer)
	return nil
}
