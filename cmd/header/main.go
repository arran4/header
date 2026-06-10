package main

import (
	"flag"
	"fmt"
	"os"
	"text/template"

	"github.com/arran4/header"
)

const outputTmpl = `URL: {{.URL}}
Title: {{.Title}}
{{if .OGType}}OpenGraph Type: {{.OGType}}{{end}}
{{if .OGDescription}}OpenGraph Description: {{.OGDescription}}{{end}}
{{if .TwitterCard}}Twitter Card: {{.TwitterCard}}{{end}}
{{if .TwitterDescription}}Twitter Description: {{.TwitterDescription}}{{end}}
{{if .YouTubeVideoID}}YouTube Video ID: {{.YouTubeVideoID}}{{end}}
{{if .YouTubeLengthSeconds}}YouTube Duration: {{.YouTubeDuration}}{{end}}
{{if .JSONLDType}}JSON-LD Type: {{.JSONLDType}}{{end}}
{{if .JSONLDName}}JSON-LD Name: {{.JSONLDName}}{{end}}
{{if .CoverImages}}
Cover Images:
{{range .CoverImages}}  - {{.}}
{{end}}{{end}}`

func main() {
	// Create a new flag set so we can call main multiple times in tests
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	testMode := fs.Bool("test", false, "Run in test mode to show diagnostics/missing data")
	showImage := fs.Bool("show-image", false, "Display the cover image (if available)")

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
	info, err := header.FetchURL(urlStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
		os.Exit(1)
	}

	tmpl, err := template.New("output").Parse(outputTmpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	if err := tmpl.Execute(os.Stdout, info); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
	}
	fmt.Println()

	if *testMode {
		runDiagnostics(info)
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
