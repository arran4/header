package cli

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/arran4/header"
	"github.com/arran4/header/internal/cli/cards"
	"golang.design/x/clipboard"
)

// CardList is a subcommand `header card list` -- List available card generators.
func CardList() error {
	fmt.Println("Available Card Generators:")
	for _, g := range cards.List() {
		fmt.Printf("  %-10s - %s\n", g.Name(), g.Description())
	}
	return nil
}

// CardFile is a subcommand `header card file` -- Output destination: file path.
//
// Flags:
//
//	userAgent:     --user-agent (default: "Mozilla/5.0") The User-Agent header to use when fetching URLs
//	cardType:      --type (default: "basic")             The type of card generator to use
//	retryAttempts: --retry-attempts (default: 3)         Number of attempts for idempotent HTTP requests (1 means no retries)
//	retryDelay:    --retry-delay (default: 250ms)        Initial delay between retries
//	retryMaxDelay: --retry-max-delay (default: 5s)       Maximum delay between retries
//	timeout:       --timeout (default: 20s)              Maximum total request timeout
//	imageTimeout:  --image-timeout (default: 10s)        Timeout for fetching images
//	requireCover:  --require-cover                       Fail if no usable cover image was fetched
//	qualityReport: --quality-report                      Write machine readable card quality JSON to file
//	out:           --out (required)                      Output file path
//	urlStr:        @1 (required)                         URL to fetch
func CardFile(
	userAgent string,
	cardType string,
	retryAttempts int,
	retryDelay time.Duration,
	retryMaxDelay time.Duration,
	timeout time.Duration,
	imageTimeout time.Duration,
	requireCover bool,
	qualityReport string,
	out string,
	urlStr string,
) error {
	cardImg, _, err := generateCardData(
		userAgent,
		cardType,
		retryAttempts,
		retryDelay,
		retryMaxDelay,
		timeout,
		imageTimeout,
		requireCover,
		qualityReport,
		urlStr,
	)
	if err != nil {
		return err
	}

	dir := filepath.Dir(out)
	if dir == "" {
		dir = "."
	}
	tmpFile, err := os.CreateTemp(dir, ".header-card-*.tmp")
	if err != nil {
		return fmt.Errorf("error creating temporary file: %w", err)
	}
	tmpName := tmpFile.Name()
	defer func() {
		_ = os.Remove(tmpName)
	}()

	if err = png.Encode(tmpFile, cardImg); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("error encoding png: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("error closing temporary file: %w", err)
	}

	if err = os.Rename(tmpName, out); err != nil {
		return fmt.Errorf("error replacing output file %s: %w", out, err)
	}

	fmt.Printf("Card written to %s\n", out)
	return nil
}

// CardClipboard is a subcommand `header card clipboard` -- Output destination: copy to system clipboard.
//
// Flags:
//
//	userAgent:     --user-agent (default: "Mozilla/5.0") The User-Agent header to use when fetching URLs
//	cardType:      --type (default: "basic")             The type of card generator to use
//	retryAttempts: --retry-attempts (default: 3)         Number of attempts for idempotent HTTP requests (1 means no retries)
//	retryDelay:    --retry-delay (default: 250ms)        Initial delay between retries
//	retryMaxDelay: --retry-max-delay (default: 5s)       Maximum delay between retries
//	timeout:       --timeout (default: 20s)              Maximum total request timeout
//	imageTimeout:  --image-timeout (default: 10s)        Timeout for fetching images
//	requireCover:  --require-cover                       Fail if no usable cover image was fetched
//	qualityReport: --quality-report                      Write machine readable card quality JSON to file
//	urlStr:        @1 (required)                         URL to fetch
func CardClipboard(
	userAgent string,
	cardType string,
	retryAttempts int,
	retryDelay time.Duration,
	retryMaxDelay time.Duration,
	timeout time.Duration,
	imageTimeout time.Duration,
	requireCover bool,
	qualityReport string,
	urlStr string,
) error {
	cardImg, _, err := generateCardData(
		userAgent,
		cardType,
		retryAttempts,
		retryDelay,
		retryMaxDelay,
		timeout,
		imageTimeout,
		requireCover,
		qualityReport,
		urlStr,
	)
	if err != nil {
		return err
	}

	err = clipboard.Init()
	if err != nil {
		return fmt.Errorf("error initializing clipboard: %w", err)
	}
	var buf bytes.Buffer
	err = png.Encode(&buf, cardImg)
	if err != nil {
		return fmt.Errorf("error encoding png: %w", err)
	}
	clipboard.Write(clipboard.FmtImage, buf.Bytes())
	fmt.Println("Card copied to clipboard.")
	return nil
}

// CardStdout is a subcommand `header card stdout` -- Output destination: standard output (pipe).
//
// Flags:
//
//	userAgent:     --user-agent (default: "Mozilla/5.0") The User-Agent header to use when fetching URLs
//	cardType:      --type (default: "basic")             The type of card generator to use
//	retryAttempts: --retry-attempts (default: 3)         Number of attempts for idempotent HTTP requests (1 means no retries)
//	retryDelay:    --retry-delay (default: 250ms)        Initial delay between retries
//	retryMaxDelay: --retry-max-delay (default: 5s)       Maximum delay between retries
//	timeout:       --timeout (default: 20s)              Maximum total request timeout
//	imageTimeout:  --image-timeout (default: 10s)        Timeout for fetching images
//	requireCover:  --require-cover                       Fail if no usable cover image was fetched
//	qualityReport: --quality-report                      Write machine readable card quality JSON to file
//	urlStr:        @1 (required)                         URL to fetch
func CardStdout(
	userAgent string,
	cardType string,
	retryAttempts int,
	retryDelay time.Duration,
	retryMaxDelay time.Duration,
	timeout time.Duration,
	imageTimeout time.Duration,
	requireCover bool,
	qualityReport string,
	urlStr string,
) error {
	cardImg, _, err := generateCardData(
		userAgent,
		cardType,
		retryAttempts,
		retryDelay,
		retryMaxDelay,
		timeout,
		imageTimeout,
		requireCover,
		qualityReport,
		urlStr,
	)
	if err != nil {
		return err
	}

	err = png.Encode(os.Stdout, cardImg)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}
	return nil
}

func generateCardData(
	userAgent string,
	cardType string,
	retryAttempts int,
	retryDelay time.Duration,
	retryMaxDelay time.Duration,
	timeout time.Duration,
	imageTimeout time.Duration,
	requireCover bool,
	qualityReport string,
	urlStr string,
) (image.Image, *cards.CardQualityReport, error) {
	fetchOpts := header.FetchOptions{
		UserAgent:     userAgent,
		Timeout:       timeout,
		RetryAttempts: retryAttempts,
		RetryDelay:    retryDelay,
		RetryMaxDelay: retryMaxDelay,
		ImageTimeout:  imageTimeout,
	}

	info, err := header.FetchURL(urlStr, fetchOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching URL: %w", err)
	}

	warnOnCardQuality(info)

	generator, err := cards.Get(cardType)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting card generator: %w", err)
	}

	assetOpts := cards.AssetFetchOptions{
		UserAgent:     userAgent,
		Timeout:       imageTimeout,
		RetryAttempts: retryAttempts,
		RetryDelay:    retryDelay,
		RetryMaxDelay: retryMaxDelay,
	}

	assets := cards.FetchCardAssets(context.Background(), info, assetOpts)

	selectedContent := cards.SelectCardContent(info)
	usedPlaceholder := assets.CoverImage == nil

	report := &cards.CardQualityReport{
		InputURL:           cards.SanitizeURL(urlStr),
		FinalURL:           cards.SanitizeURL(info.URL),
		Generator:          generator.Name(),
		SelectedCoverURL:   cards.SanitizeURL(assets.CoverResult.SelectedURL),
		Candidates:         assets.CoverResult.Candidates,
		HTTPStatus:         assets.CoverResult.HTTPStatus,
		Attempts:           assets.CoverResult.Attempts,
		CoverStatus:        assets.CoverResult.Status,
		FaviconStatus:      assets.FaviconResult.Status,
		DescriptionQuality: selectedContent.Quality,
		UsedPlaceholder:    usedPlaceholder,
	}

	if assets.CoverResult.Status != cards.ImageStatusLoaded {
		fmt.Fprintf(os.Stderr, "card-cover: %s; review before publication\n", assets.CoverResult.Status)
	}

	if qualityReport != "" {
		if qerr := cards.WriteQualityReport(qualityReport, *report); qerr != nil {
			return nil, report, fmt.Errorf("error writing quality report: %w", qerr)
		}
	}

	if requireCover && assets.CoverResult.Status != cards.ImageStatusLoaded {
		return nil, report, fmt.Errorf("required cover image was not loaded (status: %s)", assets.CoverResult.Status)
	}

	cardImg, err := cards.GenerateWithAssets(generator, info, assets)
	if err != nil {
		return nil, report, fmt.Errorf("error generating card: %w", err)
	}

	return cardImg, report, nil
}

// warnOnCardQuality emits machine-recognisable review statuses on stderr,
// keeping PNG stdout and the original extraction data unchanged.
func warnOnCardQuality(info *header.PageInfo) {
	selected := cards.SelectCardContent(info)
	if selected.Quality != cards.CardQualityCandidate {
		fmt.Fprintf(os.Stderr, "card-quality: %s (description-source=%s); review before publication\n",
			selected.Quality, selected.DescriptionSource)
	}
}
