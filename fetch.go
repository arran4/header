package header

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// FetchOptions contains configuration for fetching a URL.
type FetchOptions struct {
	UserAgent string
}

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

// FetchURL retrieves a URL and extracts metadata from it.
func FetchURL(urlStr string, opts FetchOptions) (*PageInfo, error) {
	urlStr = normaliseURL(urlStr)

	client := &http.Client{
		Timeout: 20 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	if opts.UserAgent != "" {
		req.Header.Set("User-Agent", opts.UserAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	}
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	info := &PageInfo{
		URL: urlStr,
	}

	// Set up multicast reading
	extractors := extractorRegistry
	if len(extractors) == 0 {
		return info, nil
	}

	writers := make([]*io.PipeWriter, len(extractors))
	readers := make([]*io.PipeReader, len(extractors))

	for i := range extractors {
		r, w := io.Pipe()
		readers[i] = r
		writers[i] = w
	}

	var wg sync.WaitGroup
	for i, ext := range extractors {
		wg.Add(1)
		go func(e Extractor, r *io.PipeReader) {
			defer wg.Done()
			// Extract handles its own reading from the pipe
			_ = e.Extract(r, info)
			_ = r.Close() // Ensure the reader is closed when the extractor finishes or errors out
		}(ext, readers[i])
	}

	// Read from resp.Body and broadcast to all writers
	// Use a limited reader so we don't download infinite files
	limitedBody := io.LimitReader(resp.Body, 2*1024*1024)
	buf := make([]byte, 32*1024)
	for {
		n, err := limitedBody.Read(buf)
		if n > 0 {
			for _, w := range writers {
				// We ignore write errors. A write error just means the specific extractor
				// closed its reader early (e.g. it found what it needed).
				_, _ = w.Write(buf[:n])
			}
		}
		if err != nil {
			break
		}
	}

	// Close all writers to signal EOF to the extractors
	for _, w := range writers {
		_ = w.Close()
	}

	// Wait for all extractors to finish
	wg.Wait()

	return info, nil
}

func normaliseURL(input string) string {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return input
	}
	// Default to https
	return "https://" + input
}

func isYouTubeURL(urlStr string) bool {
	return strings.Contains(urlStr, "youtube.com") || strings.Contains(urlStr, "youtu.be")
}
