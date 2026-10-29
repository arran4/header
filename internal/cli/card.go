package cli

import (
	"bytes"
	"fmt"
	"image/png"
	"os"

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
//	userAgent: --user-agent (default: "Mozilla/5.0") The User-Agent header to use when fetching URLs
//	cardType:  --type (default: "basic")             The type of card generator to use
//	out:       --out (required)                      Output file path
//	urlStr:    @1 (required)                         URL to fetch
func CardFile(userAgent string, cardType string, out string, urlStr string) (err error) {
	opts := header.FetchOptions{
		UserAgent: userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	generator, err := cards.Get(cardType)
	if err != nil {
		return fmt.Errorf("error getting card generator: %w", err)
	}

	cardImg, err := generator.Generate(info)
	if err != nil {
		return fmt.Errorf("error generating card: %w", err)
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() {
		cerr := f.Close()
		if err == nil {
			if cerr != nil {
				err = fmt.Errorf("error closing output file: %w", cerr)
			} else {
				fmt.Printf("Card written to %s\n", out)
			}
		}
	}()

	err = png.Encode(f, cardImg)
	if err != nil {
		return fmt.Errorf("error encoding png: %w", err)
	}

	return nil
}

// CardClipboard is a subcommand `header card clipboard` -- Output destination: copy to system clipboard.
//
// Flags:
//
//	userAgent: --user-agent (default: "Mozilla/5.0") The User-Agent header to use when fetching URLs
//	cardType:  --type (default: "basic")             The type of card generator to use
//	urlStr:    @1 (required)                         URL to fetch
func CardClipboard(userAgent string, cardType string, urlStr string) error {
	opts := header.FetchOptions{
		UserAgent: userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	generator, err := cards.Get(cardType)
	if err != nil {
		return fmt.Errorf("error getting card generator: %w", err)
	}

	cardImg, err := generator.Generate(info)
	if err != nil {
		return fmt.Errorf("error generating card: %w", err)
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
//	userAgent: --user-agent (default: "Mozilla/5.0") The User-Agent header to use when fetching URLs
//	cardType:  --type (default: "basic")             The type of card generator to use
//	urlStr:    @1 (required)                         URL to fetch
func CardStdout(userAgent string, cardType string, urlStr string) error {
	opts := header.FetchOptions{
		UserAgent: userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		return fmt.Errorf("error fetching URL: %w", err)
	}

	generator, err := cards.Get(cardType)
	if err != nil {
		return fmt.Errorf("error getting card generator: %w", err)
	}

	cardImg, err := generator.Generate(info)
	if err != nil {
		return fmt.Errorf("error generating card: %w", err)
	}

	err = png.Encode(os.Stdout, cardImg)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %w", err)
	}
	return nil
}
