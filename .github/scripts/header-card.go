// Generate the repository card from project-owned text, without fetching a
// private GitHub page or redistributing GitHub's error-page artwork.
package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	"github.com/arran4/header"
	"github.com/arran4/header/internal/cli/cards"
)

func main() {
	kind := flag.String("type", "basic", "card generator")
	flag.Parse()
	generator, err := cards.Get(*kind)
	if err != nil {
		log.Fatal(err)
	}
	img, err := generator.Generate(&header.PageInfo{
		URL:           "https://github.com/arran4/header",
		Title:         "header",
		OGDescription: "Extract and inspect web metadata from your terminal. OpenGraph, Twitter cards, JSON-LD, AppLinks and more.",
	})
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create("header-card.png")
	if err != nil {
		log.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
