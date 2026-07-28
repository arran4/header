package header

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/henvic/httpretty"
)

// FetchOptions contains configuration for fetching a URL.
type FetchOptions struct {
	UserAgent         string
	TraceRedirects    bool
	FollowMetaRefresh bool
	UpdateReferer     bool
	StoreCookies      bool
	RedirectLimit     int
	RedirectDelay     time.Duration
	Timeout           time.Duration
	MaxBodySize       int64
	Httpretty         bool
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
	if opts.RedirectLimit <= 0 {
		opts.RedirectLimit = 20
	}

	urlStr = normaliseURL(urlStr)

	var jar http.CookieJar
	if opts.StoreCookies {
		jar, _ = cookiejar.New(nil)
	}

	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 20 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	if opts.Httpretty {
		logger := &httpretty.Logger{
			Time:           true,
			TLS:            true,
			RequestHeader:  true,
			RequestBody:    true,
			ResponseHeader: true,
			ResponseBody:   true,
			Colors:         true,
			Formatters:     []httpretty.Formatter{&httpretty.JSONFormatter{}},
		}
		client.Transport = logger.RoundTripper(http.DefaultTransport)
	}

	info := &PageInfo{
		URL: urlStr,
	}

	nextURL := urlStr
	referer := ""

	for i := 0; i < opts.RedirectLimit; i++ {
		if nextURL == "" {
			break
		}

		if opts.TraceRedirects {
			info.Mutate(func(i *PageInfo) {
				i.Trace = append(i.Trace, nextURL)
			})
		}

		if i > 0 && opts.RedirectDelay > 0 {
			time.Sleep(opts.RedirectDelay)
		}

		req, err := http.NewRequest(http.MethodGet, nextURL, nil)
		if err != nil {
			return info, err
		}

		if opts.UserAgent != "" {
			req.Header.Set("User-Agent", opts.UserAgent)
		} else {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		}
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		if opts.UpdateReferer && referer != "" {
			req.Header.Set("Referer", referer)
		}

		resp, err := client.Do(req)
		if err != nil {
			return info, err
		}

		isRedirect := false
		if resp.StatusCode >= 300 && resp.StatusCode < 400 && resp.StatusCode != http.StatusNotModified {
			for _, header := range []string{"Location", "Redirect"} {
				if loc := resp.Header.Get(header); loc != "" {
					locURL, err := url.Parse(loc)
					if err == nil {
						nextURL = resp.Request.URL.ResolveReference(locURL).String()
						isRedirect = true
						break
					}
				}
			}
		}

		if !isRedirect {
			// Extract HTML content
			err = extractFromResponse(resp, info, opts.MaxBodySize)
			resp.Body.Close()
			if err != nil {
				return info, err
			}

			if opts.FollowMetaRefresh && info.MetaRefresh != "" {
				// Parse meta refresh content, format is usually "5; url=https://example.com"
				if idx := strings.Index(strings.ToLower(info.MetaRefresh), "url="); idx != -1 {
					metaURLStr := strings.TrimSpace(info.MetaRefresh[idx+4:])
					// Clean quotes if any
					metaURLStr = strings.Trim(metaURLStr, "\"'")
					metaURL, err := url.Parse(metaURLStr)
					if err == nil {
						nextURL = resp.Request.URL.ResolveReference(metaURL).String()
						isRedirect = true
						// Clear MetaRefresh so we don't loop on it if it's not a real redirect
						info.MetaRefresh = ""
					}
				}
			}
		} else {
			resp.Body.Close()
		}

		if !isRedirect {
			break
		}

		if opts.UpdateReferer {
			referer = resp.Request.URL.String()
		}

		if i == opts.RedirectLimit-1 {
			return info, fmt.Errorf("too many redirects")
		}
	}

	return info, nil
}

func extractFromResponse(resp *http.Response, info *PageInfo, maxBodySize int64) error {
	// Revert return nil inside extractFromResponse

	// Set up multicast reading
	extractors := extractorRegistry
	if len(extractors) == 0 {
		return nil
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
	if maxBodySize <= 0 {
		maxBodySize = 2 * 1024 * 1024 // 2MB default
	}
	limitedBody := io.LimitReader(resp.Body, maxBodySize)
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

	return nil
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
