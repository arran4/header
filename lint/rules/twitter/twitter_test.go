package twitter

import (
	"testing"

	"github.com/arran4/header"
)

func TestTwitterRule(t *testing.T) {
	rule := twitterRule{}

	tests := []struct {
		name     string
		info     *header.PageInfo
		expected int
	}{
		{
			name: "No video intent",
			info: &header.PageInfo{
				URL: "https://example.com",
			},
			expected: 0,
		},
		{
			name: "Video intent with og:video but no twitter:player",
			info: &header.PageInfo{
				URL:     "https://example.com/video",
				OGVideo: "https://example.com/video.mp4",
			},
			expected: 1,
		},
		{
			name: "Instagram reels but no twitter:player",
			info: &header.PageInfo{
				URL: "https://www.instagram.com/reels/Day4n9LOhpx/",
			},
			expected: 1,
		},
		{
			name: "Video intent with twitter:player",
			info: &header.PageInfo{
				URL:           "https://example.com/video",
				OGVideo:       "https://example.com/video.mp4",
				TwitterPlayer: "https://example.com/player",
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := rule.Check(tt.info)
			if len(results) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(results))
				if len(results) > 0 {
					t.Logf("Result: %v", results[0].Message)
				}
			}
		})
	}
}
