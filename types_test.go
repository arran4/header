package header

import (
	"reflect"
	"testing"
)

func TestPageInfo_CoverImages(t *testing.T) {
	tests := []struct {
		name     string
		info     *PageInfo
		expected []string
	}{
		{
			name: "both OGImage and distinct OGImageSecureURL are preserved",
			info: &PageInfo{
				OGImage:          "http://example.com/cover.jpg",
				OGImageSecureURL: "https://example.com/cover.jpg",
				TwitterImage:     "https://example.com/tw.jpg",
			},
			expected: []string{
				"http://example.com/cover.jpg",
				"https://example.com/cover.jpg",
				"https://example.com/tw.jpg",
			},
		},
		{
			name: "identical OGImage and OGImageSecureURL are deduplicated",
			info: &PageInfo{
				OGImage:          "https://example.com/cover.jpg",
				OGImageSecureURL: "https://example.com/cover.jpg",
				TwitterImage:     "https://example.com/tw.jpg",
			},
			expected: []string{
				"https://example.com/cover.jpg",
				"https://example.com/tw.jpg",
			},
		},
		{
			name: "OGImage empty, OGImageSecureURL present",
			info: &PageInfo{
				OGImageSecureURL: "https://example.com/secure.jpg",
				TwitterImage:     "https://example.com/tw.jpg",
			},
			expected: []string{
				"https://example.com/secure.jpg",
				"https://example.com/tw.jpg",
			},
		},
		{
			name: "only TwitterImage",
			info: &PageInfo{
				TwitterImage: "https://example.com/tw.jpg",
			},
			expected: []string{
				"https://example.com/tw.jpg",
			},
		},
		{
			name:     "no images",
			info:     &PageInfo{},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.info.CoverImages()
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("CoverImages() = %v, want %v", got, tc.expected)
			}
		})
	}
}
