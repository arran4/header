package header

import (
	"sync"
	"time"
)

// PageInfo represents the extracted information from a URL.
type PageInfo struct {
	mu sync.Mutex

	// General Information
	URL   string
	Title string

	// OpenGraph
	OGType        string
	OGImage       string
	OGDescription string

	// Twitter
	TwitterCard        string
	TwitterImage       string
	TwitterTitle       string
	TwitterDescription string

	// YouTube Specific
	YouTubeVideoID       string
	YouTubeLengthSeconds int
	YouTubeDuration      time.Duration

	// JSON-LD (First found entity or merged data)
	JSONLDType string
	JSONLDName string
}

// CoverImages returns a list of potential image URLs.
func (p *PageInfo) CoverImages() []string {
	var images []string
	if p.OGImage != "" {
		images = append(images, p.OGImage)
	}
	if p.TwitterImage != "" && p.TwitterImage != p.OGImage {
		images = append(images, p.TwitterImage)
	}
	return images
}
