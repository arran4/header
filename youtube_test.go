package header

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed testdata/youtube.html
var youtubeHTML string

func TestExtractYouTubeInfo(t *testing.T) {
	info := &PageInfo{
		URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	}
	extractor := youtubeExtractor{}
	err := extractor.Extract(strings.NewReader(youtubeHTML), info)
	if err != nil {
		t.Fatalf("Extractor returned error: %v", err)
	}

	if info.YouTubeVideoID != "dQw4w9WgXcQ" {
		t.Errorf("Expected VideoID 'dQw4w9WgXcQ', got '%s'", info.YouTubeVideoID)
	}
	if info.Title != "Rick Astley - Never Gonna Give You Up (Official Music Video)" {
		t.Errorf("Expected Title 'Rick Astley - Never Gonna Give You Up (Official Music Video)', got '%s'", info.Title)
	}
	if info.YouTubeLengthSeconds != 212 {
		t.Errorf("Expected LengthSeconds 212, got %d", info.YouTubeLengthSeconds)
	}
}
