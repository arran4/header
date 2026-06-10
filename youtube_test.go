package header

import (
	"testing"
)

func TestExtractYouTubeInfo(t *testing.T) {
	scripts := []string{
		`var ytInitialPlayerResponse = {"videoDetails":{"videoId":"dQw4w9WgXcQ","title":"Rick Astley - Never Gonna Give You Up (Official Music Video)","lengthSeconds":"212"}};`,
	}

	info := &PageInfo{}
	extractYouTubeInfo(scripts, info)

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
