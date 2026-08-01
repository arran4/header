package images

import (
	"fmt"
	"image"
	// Blank imports to support decoding various image formats natively.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/arran4/header"
	"github.com/arran4/header/lint"
)

const RuleID = "images"

// imagesRule is a lint rule that checks the validity and metadata of referenced images.
// What it checks:
// - Verifies image reachability by making a GET request to the image URLs (like `og:image`, `twitter:image`).
// - Checks that the `Content-Type` header is an image type and matches the expected type (e.g., `og:image:type`).
// - Downloads the beginning of the image to decode its configuration (dimensions and format).
// - Cross-references the actual decoded dimensions (width/height) with the expected metadata dimensions (like `og:image:width` and `og:image:height`).
//
// Why it checks this:
// - Providing incorrect image dimensions or formats in metadata can lead to broken or incorrectly cropped previews.
// - Ensuring the image is reachable and valid prevents social platforms from falling back to empty or default thumbnails.
//
// References:
// - Open Graph Image Guidelines: https://ogp.me/#structured
// - Twitter Card Image Guidelines: https://developer.x.com/en/docs/twitter-for-websites/cards/overview/summary-card-with-large-image
type imagesRule struct{}

func (r imagesRule) ID() string   { return RuleID }
func (r imagesRule) Name() string { return "Image Validity & Metadata Checks" }

// HTTPClient represents the interface we need from the HTTP client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (r imagesRule) Check(info *header.PageInfo, ops ...any) []lint.Result {
	var results []lint.Result

	var client HTTPClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	for _, opt := range ops {
		switch o := opt.(type) {
		case HTTPClient:
			client = o
		}
	}

	type checkTarget struct {
		url            string
		expectedWidth  string
		expectedHeight string
		expectedType   string
	}

	targets := make(map[string]checkTarget)

	add := func(u string, width string, height string, expectedType string) {
		if u == "" {
			return
		}
		if info.URL != "" {
			base, err := url.Parse(info.URL)
			if err == nil {
				parsedURL, err := url.Parse(u)
				if err == nil {
					u = base.ResolveReference(parsedURL).String()
				}
			}
		}
		// If same URL is already present, we just keep the first one or merge.
		// For simplicity, we just take the first definition.
		if _, exists := targets[u]; !exists {
			targets[u] = checkTarget{
				url:            u,
				expectedWidth:  width,
				expectedHeight: height,
				expectedType:   expectedType,
			}
		}
	}

	// Add images we care about
	add(info.OGImage, info.OGImageWidth, info.OGImageHeight, info.OGImageType)
	add(info.OGImageSecureURL, info.OGImageWidth, info.OGImageHeight, info.OGImageType)
	add(info.TwitterImage, "", "", "")

	if len(targets) == 0 {
		return results
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, t := range targets {
		wg.Add(1)
		go func(t checkTarget) {
			defer wg.Done()

			req, err := http.NewRequest("GET", t.url, nil)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Invalid image URL %s: %v", t.url, err)})
				mu.Unlock()
				return
			}
			req.Header.Set("User-Agent", "Header-Lint-Bot/1.0")

			// We only need the start of the image to decode config, but standard http client downloads full.
			// Ideally we would do a Range request, but many servers don't support it or require it.
			// Getting the whole image is okay for a linting tool within a reasonable timeout.
			resp, err := client.Do(req)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Unreachable image URL %s: %v", t.url, err)})
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Image URL %s returned status %d", t.url, resp.StatusCode)})
				mu.Unlock()
				return
			}

			// Check Content-Type
			contentType := resp.Header.Get("Content-Type")
			if contentType == "" {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("Image URL %s missing Content-Type header", t.url)})
				mu.Unlock()
			} else if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Image URL %s returned non-image Content-Type: %s", t.url, contentType)})
				mu.Unlock()
			} else if t.expectedType != "" && contentType != t.expectedType {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("Image URL %s expected type %s but got %s", t.url, t.expectedType, contentType)})
				mu.Unlock()
			}

			// Parse image configuration
			config, _, err := image.DecodeConfig(resp.Body)
			if err != nil {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Failed to decode image at %s: %v", t.url, err)})
				mu.Unlock()
				return
			}

			if config.Width <= 0 || config.Height <= 0 {
				mu.Lock()
				results = append(results, lint.Result{Severity: lint.SeverityError, Message: fmt.Sprintf("Image at %s has invalid dimensions: %dx%d", t.url, config.Width, config.Height)})
				mu.Unlock()
			}

			if t.expectedType != "" {
				expectedFormat := strings.TrimPrefix(t.expectedType, "image/")
				if expectedFormat == "jpeg" {
					expectedFormat = "jpeg"
				}
				// The image decoder determines the format. A mismatch is noted but we do not log it
				// directly here to avoid noise, as content-type mismatches are already flagged.
				_ = expectedFormat
			}

			if t.expectedWidth != "" {
				if ew, err := strconv.Atoi(t.expectedWidth); err == nil && ew != config.Width {
					mu.Lock()
					results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("Image at %s width mismatch: metadata says %d, actual is %d", t.url, ew, config.Width)})
					mu.Unlock()
				}
			}

			if t.expectedHeight != "" {
				if eh, err := strconv.Atoi(t.expectedHeight); err == nil && eh != config.Height {
					mu.Lock()
					results = append(results, lint.Result{Severity: lint.SeverityWarning, Message: fmt.Sprintf("Image at %s height mismatch: metadata says %d, actual is %d", t.url, eh, config.Height)})
					mu.Unlock()
				}
			}

		}(t)
	}

	wg.Wait()

	return results
}

func init() {
	lint.RegisterRule(imagesRule{})
}
