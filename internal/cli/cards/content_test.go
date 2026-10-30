package cards

import (
	"strings"
	"testing"

	"github.com/arran4/header"
)

func TestSelectCardContentGitHubLXA(t *testing.T) {
	// These fields are from the genuine header capture preserved in
	// arran4/public-relations/projects/lxa/recordings/production/header/metadata.json.
	raw := &header.PageInfo{
		URL:                "https://github.com/arran4/lxa",
		Title:              "GitHub - arran4/lxa: ls with extended attributes · GitHub",
		OGTitle:            "GitHub - arran4/lxa: ls with extended attributes",
		OGDescription:      "ls with extended attributes. Contribute to arran4/lxa development by creating an account on GitHub.",
		TwitterDescription: "ls with extended attributes. Contribute to arran4/lxa development by creating an account on GitHub.",
	}
	c := SelectCardContent(raw)
	if c.Title != "arran4/lxa" || c.Description != "ls with extended attributes" {
		t.Fatalf("unexpected content: %+v", c)
	}
	if c.TitleSource != "github-repository-url" || c.DescriptionSource != "page-og-description+github-boilerplate-removed" ||
		c.Quality != CardQualityCandidate {
		t.Fatalf("unexpected provenance: %+v", c)
	}
	// Raw extraction and JSON/template outputs must not be mutated.
	if !strings.Contains(raw.OGDescription, "Contribute to arran4/lxa development") ||
		!strings.HasPrefix(raw.Title, "GitHub - ") {
		t.Fatalf("raw source metadata was modified: %+v", raw)
	}
}

func TestSelectCardContentGitHubFallbackAndEmpty(t *testing.T) {
	tests := []struct {
		name, url, title, og, twitter, want, source, quality string
	}{
		{
			name: "twitter fallback after empty github description",
			url: "https://github.com/arran4/lxa",
			og: "Contribute to arran4/lxa development by creating an account on GitHub.",
			twitter: "A separate, real repository description.",
			want: "A separate, real repository description.",
			source: "page-twitter-description", quality: CardQualityCandidate,
		},
		{
			name: "both github descriptions empty after removing boilerplate",
			url: "https://github.com/arran4/empty",
			og: "Contribute to arran4/empty development by creating an account on GitHub.",
			twitter: "Contribute to arran4/empty development by creating an account on GitHub.",
			want: "", source: "none", quality: CardQualityNeedsOwnerDescription,
		},
		{
			name: "github page with legitimately empty metadata",
			url: "https://github.com/arran4/empty",
			want: "", source: "none", quality: CardQualityNeedsOwnerDescription,
		},
		{
			name: "mismatched boilerplate requires review rather than fabricated text",
			url: "https://github.com/arran4/lxa",
			og: "Contribute to someone/else development by creating an account on GitHub.",
			want: "", source: "none", quality: CardQualityNeedsOwnerDescription,
		},
		{
			name: "non-github source is unchanged",
			title: "An actual page title",
			url: "https://example.com",
			og: "Contribute to arran4/lxa development by creating an account on GitHub.",
			want: "Contribute to arran4/lxa development by creating an account on GitHub.",
			source: "page-og-description", quality: CardQualityCandidate,
		},
		{
			name: "github issue subpage is not normalized",
			title: "An actual page title",
			url: "https://github.com/arran4/lxa/issues/1",
			og: "Contribute to arran4/lxa development by creating an account on GitHub.",
			want: "Contribute to arran4/lxa development by creating an account on GitHub.",
			source: "page-og-description", quality: CardQualityCandidate,
		},
		{
			name: "github lookalike domain is not normalized",
			title: "An actual page title",
			url: "https://github.com.evil.test/arran4/lxa",
			og: "Contribute to arran4/lxa development by creating an account on GitHub.",
			want: "Contribute to arran4/lxa development by creating an account on GitHub.",
			source: "page-og-description", quality: CardQualityCandidate,
		},
		{
			name: "github owner page is not normalized",
			title: "An actual page title",
			url: "https://github.com/arran4",
			og: "Contribute to arran4/lxa development by creating an account on GitHub.",
			want: "Contribute to arran4/lxa development by creating an account on GitHub.",
			source: "page-og-description", quality: CardQualityCandidate,
		},
		{
			name: "legitimate repo description is preserved",
			url: "https://github.com/arran4/lxa",
			og: "An independently written description of lxa.",
			want: "An independently written description of lxa.",
			source: "page-og-description", quality: CardQualityCandidate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SelectCardContent(&header.PageInfo{
				URL: tt.url, Title: tt.title, OGDescription: tt.og, TwitterDescription: tt.twitter,
			})
			if c.Description != tt.want || c.DescriptionSource != tt.source || c.Quality != tt.quality {
				t.Fatalf("selected %+v; expected desc=%q source=%q quality=%q", c, tt.want, tt.source, tt.quality)
			}
			if tt.title != "" && (c.Title != tt.title || c.TitleSource != "page-html-title") {
				t.Fatalf("non-repository title was changed: %+v, expected %q", c, tt.title)
			}
		})
	}
}

// A meaningful description does not make a card with no title publishable.
func TestSelectCardContentMissingTitleNeedsReview(t *testing.T) {
	c := SelectCardContent(&header.PageInfo{
		URL:           "https://example.com",
		OGDescription: "A real page description.",
	})
	if c.Title != "No Title" || c.Quality != CardQualityNeedsReview {
		t.Fatalf("missing title must require review: %+v", c)
	}
}

func TestSelectCardContentLongTitleAndUnknownURL(t *testing.T) {
	long := strings.Repeat("a long repository description ", 10)
	c := SelectCardContent(&header.PageInfo{
		URL: "https://github.com/arran4/long-title",
		Title: "GitHub - arran4/long-title: " + long + " · GitHub",
		OGDescription: long + "Contribute to arran4/long-title development by creating an account on GitHub.",
	})
	if c.Title != "arran4/long-title" || c.Description != strings.TrimSpace(long) {
		t.Fatalf("unexpected long title/description selection: %+v", c)
	}
	if c := SelectCardContent(nil); c.Quality != CardQualityNeedsOwnerDescription {
		t.Fatalf("nil info must request review: %+v", c)
	}
}

func TestSelectCardContentRenderers(t *testing.T) {
	// Both public card styles must accept the same selected metadata and
	// continue emitting valid 800x400 images for a repository with no cover.
	info := &header.PageInfo{
		URL: "https://github.com/arran4/lxa",
		Title: "GitHub - arran4/lxa: ls with extended attributes · GitHub",
		OGDescription: "ls with extended attributes. Contribute to arran4/lxa development by creating an account on GitHub.",
	}
	for _, style := range []string{"basic", "darkmode"} {
		t.Run(style, func(t *testing.T) {
			g, err := Get(style)
			if err != nil {
				t.Fatal(err)
			}
			img, err := g.Generate(info)
			if err != nil {
				t.Fatal(err)
			}
			if img.Bounds().Dx() != 800 || img.Bounds().Dy() != 400 {
				t.Fatalf("unexpected dimensions: %v", img.Bounds())
			}
		})
	}
}
