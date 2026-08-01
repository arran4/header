// Package header provides tools for extracting metadata from web pages.
package header

import "golang.org/x/net/html"

// Extractor creates a stateful Processor for extracting metadata from a stream.
type Extractor interface {
	Name() string
	Init(info *PageInfo) Processor
}

// Processor receives a stream of HTML tokens and extracts metadata into PageInfo.
type Processor interface {
	// ProcessToken processes a single HTML token. It returns true if the processor
	// has finished and no longer needs to receive tokens.
	ProcessToken(t html.Token) (stop bool)
}

var extractorRegistry []Extractor

// RegisterExtractor adds an extractor to the list of extractors used when fetching a URL.
func RegisterExtractor(e Extractor) {
	extractorRegistry = append(extractorRegistry, e)
}
