package main

import (
	"fmt"
	"strings"

	"github.com/arran4/header"
)

func runDiagnostics(info *header.PageInfo) {
	fmt.Println("--- Diagnostics ---")

	var missing []string
	if info.Title == "" {
		missing = append(missing, "Title")
	}
	if info.OGType == "" {
		missing = append(missing, "OpenGraph Data")
	}
	if info.TwitterCard == "" {
		missing = append(missing, "Twitter Card Data")
	}
	if info.JSONLDType == "" {
		missing = append(missing, "JSON-LD Metadata")
	}
	if len(info.CoverImages()) == 0 {
		missing = append(missing, "Cover Images")
	}

	if len(missing) == 0 {
		fmt.Println("All common metadata fields are present.")
	} else {
		fmt.Println("Missing fields:")
		for _, m := range missing {
			fmt.Printf("  - %s\n", m)
		}
	}

	if strings.Contains(info.URL, "youtube.com") || strings.Contains(info.URL, "youtu.be") {
		if info.YouTubeVideoID == "" {
			fmt.Println("  - Warning: YouTube URL detected but failed to extract Video ID/Watch time.")
		}
	}
	fmt.Println("-------------------")
}
