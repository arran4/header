package cards

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"  // register GIF format for image.Decode
	_ "image/jpeg" // register JPEG format for image.Decode
	_ "image/png"  // register PNG format for image.Decode
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arran4/header"
	"github.com/skip2/go-qrcode"
	_ "golang.org/x/image/webp" // register WebP format for image.Decode
)

const (
	// ImageStatusLoaded means the image candidate was successfully downloaded and decoded.
	ImageStatusLoaded = "loaded"
	// ImageStatusNoImageMetadata means no candidate URLs were found in metadata.
	ImageStatusNoImageMetadata = "no-image-metadata"
	// ImageStatusHTTPStatus means candidate fetch failed with a non-200 HTTP status.
	ImageStatusHTTPStatus = "http-status"
	// ImageStatusTimeout means request timed out.
	ImageStatusTimeout = "timeout"
	// ImageStatusNetworkError means network-level failure.
	ImageStatusNetworkError = "network-error"
	// ImageStatusDecodeError means image decoding failed or image format is invalid.
	ImageStatusDecodeError = "decode-error"
	// ImageStatusImageTooLarge means byte size or dimensions exceeded limits.
	ImageStatusImageTooLarge = "image-too-large"
	// ImageStatusAllCandidatesFailed means all available candidate URLs were tried and failed.
	ImageStatusAllCandidatesFailed = "all-candidates-failed"
)

const (
	// MaxImageBytes is the maximum allowed byte size for downloaded images (20 MB).
	MaxImageBytes = 20 * 1024 * 1024
	// MaxImageDimension is the maximum width or height for decoded images (10,000 px).
	MaxImageDimension = 10000
	// MaxImagePixels is the maximum total pixel count to prevent decompression bombs (40 MP).
	MaxImagePixels = 40000000
)

// CandidateOutcome records the fetch and decode outcome for a single image URL candidate.
type CandidateOutcome struct {
	URL        string `json:"url"`
	Status     string `json:"status"`
	StatusCode int    `json:"http_status,omitempty"`
	Attempts   int    `json:"attempts"`
	Error      string `json:"error,omitempty"`
}

// ImageResult contains structured diagnostics for image fetching.
type ImageResult struct {
	Status        string             `json:"status"`
	SelectedURL   string             `json:"selected_url,omitempty"`
	HTTPStatus    int                `json:"http_status,omitempty"`
	Attempts      int                `json:"attempts"`
	Candidates    []CandidateOutcome `json:"candidates,omitempty"`
	FailureReason string             `json:"failure_reason,omitempty"`
}

// CardAssets holds all fetched and pre-rendered assets required to generate card images.
type CardAssets struct {
	CoverImage    image.Image
	CoverResult   ImageResult
	FaviconImage  image.Image
	FaviconResult ImageResult
	QRImage       image.Image
}

// CardQualityReport is the machine-readable output for card generation.
type CardQualityReport struct {
	InputURL           string             `json:"input_url"`
	FinalURL           string             `json:"final_url"`
	Generator          string             `json:"generator"`
	SelectedCoverURL   string             `json:"selected_cover_url,omitempty"`
	Candidates         []CandidateOutcome `json:"candidates,omitempty"`
	HTTPStatus         int                `json:"http_status,omitempty"`
	Attempts           int                `json:"attempts"`
	CoverStatus        string             `json:"cover_status"`
	FaviconStatus      string             `json:"favicon_status"`
	DescriptionQuality string             `json:"description_quality"`
	UsedPlaceholder    bool               `json:"used_placeholder"`
}

// AssetFetchOptions configures asset fetching.
type AssetFetchOptions struct {
	UserAgent     string
	Timeout       time.Duration // Total deadline per asset
	RetryAttempts int
	RetryDelay    time.Duration
	RetryMaxDelay time.Duration
	HTTPClient    header.HTTPDoer
}

// DefaultAssetFetchOptions returns default options for asset fetching.
func DefaultAssetFetchOptions() AssetFetchOptions {
	return AssetFetchOptions{
		Timeout:       10 * time.Second,
		RetryAttempts: 3,
		RetryDelay:    250 * time.Millisecond,
		RetryMaxDelay: 5 * time.Second,
	}
}

// AssetGenerator is an interface implemented by generators that can render from pre-fetched assets.
type AssetGenerator interface {
	Generator
	GenerateWithAssets(info *header.PageInfo, assets *CardAssets) (image.Image, error)
}

// GenerateWithAssets generates a card image using the provided generator and pre-fetched assets.
func GenerateWithAssets(g Generator, info *header.PageInfo, assets *CardAssets) (image.Image, error) {
	if ag, ok := g.(AssetGenerator); ok {
		return ag.GenerateWithAssets(info, assets)
	}
	return g.Generate(info)
}

// ResolveAssetURL resolves rawURL against baseURL, supporting protocol-relative and relative URLs.
func ResolveAssetURL(rawURL, baseURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", errors.New("empty URL")
	}
	if strings.HasPrefix(rawURL, "//") {
		base, err := url.Parse(baseURL)
		if err == nil && base.Scheme != "" {
			return base.Scheme + ":" + rawURL, nil
		}
		return "https:" + rawURL, nil
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if parsed.IsAbs() {
		return parsed.String(), nil
	}
	if baseURL == "" {
		return parsed.String(), nil
	}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(parsed).String(), nil
}

// SanitizeURL scrubs passwords and sensitive query tokens from URLs.
func SanitizeURL(rawURL string) string {
	return header.SanitizeURL(rawURL)
}

// SanitizeError scrubs sensitive tokens and credentials from error messages.
func SanitizeError(err error) string {
	return header.SanitizeError(err)
}

// SanitizeErrorMessage scrubs sensitive tokens and credentials from error strings.
func SanitizeErrorMessage(msg string) string {
	return header.SanitizeErrorMessage(msg)
}

// FetchCardAssets fetches cover image and favicon candidates using bounded retries and diagnostics.
func FetchCardAssets(ctx context.Context, info *header.PageInfo, opts AssetFetchOptions) *CardAssets {
	assets := &CardAssets{}
	if info == nil {
		assets.CoverResult = ImageResult{Status: ImageStatusNoImageMetadata, FailureReason: "no page info"}
		assets.FaviconResult = ImageResult{Status: ImageStatusNoImageMetadata, FailureReason: "no page info"}
		return assets
	}

	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.RetryAttempts <= 0 {
		opts.RetryAttempts = 3
	}
	if opts.RetryDelay <= 0 {
		opts.RetryDelay = 250 * time.Millisecond
	}
	if opts.RetryMaxDelay <= 0 {
		opts.RetryMaxDelay = 5 * time.Second
	}

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: opts.Timeout,
		}
	}

	policy := header.RetryPolicy{
		Attempts: opts.RetryAttempts,
		Delay:    opts.RetryDelay,
		MaxDelay: opts.RetryMaxDelay,
	}

	// 1. Fetch Cover Image Candidates in priority order: OGImage, OGImageSecureURL, TwitterImage
	coverCandidates := collectCoverCandidates(info)
	assets.CoverImage, assets.CoverResult = fetchImageCandidates(ctx, client, policy, coverCandidates, opts.UserAgent, opts.Timeout)

	// 2. Fetch Favicon Candidates
	faviconCandidates := collectFaviconCandidates(info)
	assets.FaviconImage, assets.FaviconResult = fetchImageCandidates(ctx, client, policy, faviconCandidates, opts.UserAgent, opts.Timeout)

	// 3. Generate QR Code
	if info.URL != "" {
		qrCode, err := qrcode.Encode(info.URL, qrcode.Medium, 256)
		if err == nil {
			qrDecoded, _, err := image.Decode(bytes.NewReader(qrCode))
			if err == nil {
				assets.QRImage = qrDecoded
			}
		}
	}

	return assets
}

func collectCoverCandidates(info *header.PageInfo) []string {
	var candidates []string
	seen := make(map[string]bool)

	raw := []string{info.OGImage, info.OGImageSecureURL, info.TwitterImage}
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		resolved, err := ResolveAssetURL(r, info.URL)
		if err == nil && resolved != "" && !seen[resolved] {
			seen[resolved] = true
			candidates = append(candidates, resolved)
		}
	}
	return candidates
}

func collectFaviconCandidates(info *header.PageInfo) []string {
	var candidates []string
	seen := make(map[string]bool)

	for _, icon := range info.Icons {
		r := strings.TrimSpace(icon.URL)
		if r == "" {
			continue
		}
		resolved, err := ResolveAssetURL(r, info.URL)
		if err == nil && resolved != "" && !seen[resolved] {
			seen[resolved] = true
			candidates = append(candidates, resolved)
		}
	}
	return candidates
}

func fetchImageCandidates(
	ctx context.Context,
	client header.HTTPDoer,
	policy header.RetryPolicy,
	candidates []string,
	userAgent string,
	timeout time.Duration,
) (image.Image, ImageResult) {
	result := ImageResult{}
	if len(candidates) == 0 {
		result.Status = ImageStatusNoImageMetadata
		result.FailureReason = "no candidate image URLs in metadata"
		return nil, result
	}

	for _, rawURL := range candidates {
		sanitizedURL := SanitizeURL(rawURL)
		reqCtx, cancel := context.WithTimeout(ctx, timeout)

		req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, rawURL, nil)
		if err != nil {
			cancel()
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:      sanitizedURL,
				Status:   ImageStatusNetworkError,
				Attempts: 1,
				Error:    SanitizeError(err),
			})
			result.Attempts++
			continue
		}

		if userAgent != "" {
			req.Header.Set("User-Agent", userAgent)
		} else {
			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
		}
		req.Header.Set("Accept", "image/webp,image/png,image/jpeg,image/*;q=0.8,*/*;q=0.5")

		resp, attempts, err := header.ExecuteRequestWithRetry(reqCtx, client, req, policy)
		result.Attempts += attempts

		if err != nil {
			cancel()
			status := ImageStatusNetworkError
			if errors.Is(err, context.DeadlineExceeded) || reqCtx.Err() == context.DeadlineExceeded {
				status = ImageStatusTimeout
			}
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:      sanitizedURL,
				Status:   status,
				Attempts: attempts,
				Error:    SanitizeError(err),
			})
			continue
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			cancel()
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusHTTPStatus,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      fmt.Sprintf("HTTP %d", resp.StatusCode),
			})
			continue
		}

		// Read body bounded by MaxImageBytes + 1
		limitReader := io.LimitReader(resp.Body, MaxImageBytes+1)
		bodyBytes, err := io.ReadAll(limitReader)
		_ = resp.Body.Close()
		cancel()

		if err != nil {
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusNetworkError,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      SanitizeError(err),
			})
			continue
		}

		if int64(len(bodyBytes)) > MaxImageBytes {
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusImageTooLarge,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      fmt.Sprintf("image byte size %d exceeds limit %d", len(bodyBytes), MaxImageBytes),
			})
			continue
		}

		// DecodeConfig to guard against decompression bombs and pathological dimensions
		cfg, _, err := image.DecodeConfig(bytes.NewReader(bodyBytes))
		if err != nil {
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusDecodeError,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      SanitizeError(err),
			})
			continue
		}

		if cfg.Width <= 0 || cfg.Height <= 0 {
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusDecodeError,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      "invalid image dimensions",
			})
			continue
		}

		if cfg.Width > MaxImageDimension || cfg.Height > MaxImageDimension || int64(cfg.Width)*int64(cfg.Height) > MaxImagePixels {
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusImageTooLarge,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      fmt.Sprintf("image dimensions %dx%d exceed maximum bounds", cfg.Width, cfg.Height),
			})
			continue
		}

		// Decode image pixels
		decoded, _, err := image.Decode(bytes.NewReader(bodyBytes))
		if err != nil {
			result.Candidates = append(result.Candidates, CandidateOutcome{
				URL:        sanitizedURL,
				Status:     ImageStatusDecodeError,
				StatusCode: resp.StatusCode,
				Attempts:   attempts,
				Error:      SanitizeError(err),
			})
			continue
		}

		// Loaded successfully!
		outcome := CandidateOutcome{
			URL:        sanitizedURL,
			Status:     ImageStatusLoaded,
			StatusCode: resp.StatusCode,
			Attempts:   attempts,
		}
		result.Candidates = append(result.Candidates, outcome)
		result.Status = ImageStatusLoaded
		result.SelectedURL = sanitizedURL
		result.HTTPStatus = resp.StatusCode
		return decoded, result
	}

	// If no candidate succeeded
	if len(result.Candidates) == 1 {
		result.Status = result.Candidates[0].Status
		result.HTTPStatus = result.Candidates[0].StatusCode
		result.FailureReason = result.Candidates[0].Error
	} else {
		result.Status = ImageStatusAllCandidatesFailed
		result.FailureReason = "all candidates failed"
	}

	return nil, result
}

// WriteQualityReport writes the card quality report JSON atomically to path.
func WriteQualityReport(path string, report CardQualityReport) error {
	if path == "" {
		return nil
	}
	report.InputURL = SanitizeURL(report.InputURL)
	report.FinalURL = SanitizeURL(report.FinalURL)
	report.SelectedCoverURL = SanitizeURL(report.SelectedCoverURL)
	for i := range report.Candidates {
		report.Candidates[i].URL = SanitizeURL(report.Candidates[i].URL)
		report.Candidates[i].Error = SanitizeErrorMessage(report.Candidates[i].Error)
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling quality report: %w", err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	tmpFile, err := os.CreateTemp(dir, ".header-report-*.tmp")
	if err != nil {
		return fmt.Errorf("error creating temporary quality report: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("error writing temporary quality report: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary quality report: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("error replacing quality report %s: %w", path, err)
	}

	return nil
}
