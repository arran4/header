// Package header provides tools for extracting metadata from web pages.
package header

import "golang.org/x/net/html"

// Extractor is an interface for extracting metadata from an HTML document.
type Extractor interface {
	Name() string
	Extract(n *html.Node, info *PageInfo) error
}

var extractorRegistry []Extractor

// RegisterExtractor adds an extractor to the list of extractors used when fetching a URL.
func RegisterExtractor(e Extractor) {
	extractorRegistry = append(extractorRegistry, e)
}
