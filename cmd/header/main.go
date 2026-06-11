package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/arran4/header"
	"github.com/arran4/header/diagnostic"
	"github.com/arran4/header/templates"
	"github.com/fatih/color"
)

func main() {
	// Create a new flag set so we can call main multiple times in tests
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	testMode := fs.Bool("test", false, "Run in test mode to show diagnostics/missing data")
	showImage := fs.Bool("show-image", false, "Display the cover image (if available)")
	style := fs.String("style", "default", "The output style template to use (e.g. default, plain)")
	userAgent := fs.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36", "The User-Agent header to use when fetching URLs")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(2)
	}

	args := fs.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <url>\n", os.Args[0])
		fs.PrintDefaults()
		os.Exit(2)
	}

	urlStr := args[0]

	opts := header.FetchOptions{
		UserAgent: *userAgent,
	}

	info, err := header.FetchURL(urlStr, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		os.Exit(1)
	}

	tmplContent, err := templates.FS.ReadFile(*style + ".tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading template '%s': %v\n", *style, err)
		os.Exit(1)
	}

	funcMap := template.FuncMap{
		"bold":   color.New(color.Bold).SprintFunc(),
		"green":  color.New(color.FgGreen).SprintFunc(),
		"cyan":   color.New(color.FgCyan).SprintFunc(),
		"yellow": color.New(color.FgYellow).SprintFunc(),
	}

	tmpl, err := template.New("output").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, info); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
	}
	fmt.Print(buf.String())
	fmt.Println()

	if *testMode {
		diagnostic.Run(info)
	}

	if *showImage {
		images := info.CoverImages()
		if len(images) > 0 {
			displayImage(images[0])
		} else {
			fmt.Println("No cover image found to display.")
		}
	}
}
