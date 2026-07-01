package diagnostic

import (
	"fmt"
	"net/url"
	"github.com/arran4/header"
)

type youtubeRule struct{}

func (r youtubeRule) Name() string { return "YouTubeCheck" }

func (r youtubeRule) Check(info *header.PageInfo) []Result {
	var results []Result

	u, err := url.Parse(info.URL)
	if err != nil {
		results = append(results, Result{SeverityError, fmt.Sprintf("Failed to parse URL: %v", err)})
		return results
	}

	hostname := u.Hostname()
	if hostname == "youtube.com" || hostname == "www.youtube.com" || hostname == "youtu.be" {
		if info.YouTubeVideoID == "" {
			results = append(results, Result{SeverityWarning, "YouTube URL detected but failed to extract Video ID/Watch time."})
		}
	}

	return results
}

func init() {
	Register(youtubeRule{})
}
