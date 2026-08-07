package cli

import (
	"bytes"
	"flag"
	"fmt"
	"image/png"
	"os"

	"github.com/arran4/header"
	"github.com/arran4/header/internal/cli/cards"
	"golang.design/x/clipboard"
)

// CardCmd handles the card subcommand.
func CardCmd(args []string) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: header card <subcommand> [flags] <url>\n")
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		fmt.Fprintf(os.Stderr, "  list        List available card generators\n")
		fmt.Fprintf(os.Stderr, "  file        Output destination: file path\n")
		fmt.Fprintf(os.Stderr, "  clipboard   Output destination: copy to system clipboard\n")
		fmt.Fprintf(os.Stderr, "  stdout      Output destination: standard output (pipe)\n")
		return 2
	}

	subcommand := args[0]

	if subcommand == "list" {
		fs := flag.NewFlagSet("header card list", flag.ContinueOnError)
		if err := fs.Parse(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
			return 2
		}
		fmt.Println("Available Card Generators:")
		for _, g := range cards.List() {
			fmt.Printf("  %-10s - %s\n", g.Name(), g.Description())
		}
		return 0
	}

	fs := flag.NewFlagSet("header card "+subcommand, flag.ContinueOnError)
	userAgent := fs.String("user-agent", "Mozilla/5.0", "The User-Agent header to use when fetching URLs")
	cardType := fs.String("type", "basic", "The type of card generator to use")

	var outputFile *string
	if subcommand == "file" {
		outputFile = fs.String("out", "", "Output file path (required)")
	}

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		return 2
	}

	pargs := fs.Args()
	if len(pargs) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: header card %s [flags] <url>\n", subcommand)
		fs.PrintDefaults()
		return 2
	}

	if subcommand == "file" && *outputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -out flag is required for file subcommand\n")
		return 2
	}

	urlStr := pargs[0]

	opts := header.FetchOptions{
		UserAgent: *userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		return 1
	}

	generator, err := cards.Get(*cardType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting card generator: %v\n", err)
		return 1
	}

	cardImg, err := generator.Generate(info)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating card: %v\n", err)
		return 1
	}

	switch subcommand {
	case "clipboard":
		err = clipboard.Init()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing clipboard: %v\n", err)
			return 1
		}
		var buf bytes.Buffer
		err = png.Encode(&buf, cardImg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding png: %v\n", err)
			return 1
		}
		clipboard.Write(clipboard.FmtImage, buf.Bytes())
		fmt.Println("Card copied to clipboard.")
	case "stdout":
		err = png.Encode(os.Stdout, cardImg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to stdout: %v\n", err)
			return 1
		}
	case "file":
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
			return 1
		}
		defer f.Close()
		err = png.Encode(f, cardImg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding png: %v\n", err)
			return 1
		}
		fmt.Printf("Card written to %s\n", *outputFile)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		return 2
	}

	return 0
}
