// Package header provides tools for extracting metadata from web pages.
package header

import "io"

// Extractor is an interface for extracting metadata from an HTML document.
type Extractor interface {
	Name() string
	Extract(r io.Reader, info *PageInfo) error
}

var extractorRegistry []Extractor

// RegisterExtractor adds an extractor to the list of extractors used when fetching a URL.
func RegisterExtractor(e Extractor) {
	extractorRegistry = append(extractorRegistry, e)
}
