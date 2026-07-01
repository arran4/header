package header

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type youtubeExtractor struct{}

func (e youtubeExtractor) Name() string { return "YouTube" }

func (e youtubeExtractor) Extract(r io.Reader, info *PageInfo) error {
	if !isYouTubeURL(info.URL) {
		return nil
	}

	z := html.NewTokenizer(r)

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				return nil
			}
			return z.Err()
		}

		if tt == html.StartTagToken {
			t := z.Token()
			if t.Data == "script" {
				var isJSONLD bool
				for _, attr := range t.Attr {
					if attr.Key == "type" && attr.Val == "application/ld+json" {
						isJSONLD = true
						break
					}
				}

				if !isJSONLD {
					tt = z.Next()
					if tt == html.TextToken {
						script := z.Token().Data
						jsonText, ok := extractAssignedJSONObject(script, "ytInitialPlayerResponse")
						if ok {
							var pr playerResponse
							if err := json.Unmarshal([]byte(jsonText), &pr); err == nil {
								if pr.VideoDetails.LengthSeconds != "" {
									sec, _ := strconv.Atoi(pr.VideoDetails.LengthSeconds)
									info.SetYouTubeInfo(pr.VideoDetails.VideoID, pr.VideoDetails.Title, sec)
									return nil // We found what we need, stop parsing early
								}
							}
						}
					}
				}
			}
		}
	}
}

func init() {
	RegisterExtractor(youtubeExtractor{})
}

type playerResponse struct {
	VideoDetails struct {
		LengthSeconds string `json:"lengthSeconds"`
		Title         string `json:"title"`
		VideoID       string `json:"videoId"`
	} `json:"videoDetails"`
}

func extractAssignedJSONObject(script string, variableName string) (string, bool) {
	pos := strings.Index(script, variableName)
	if pos < 0 {
		return "", false
	}

	eq := strings.IndexByte(script[pos:], '=')
	if eq < 0 {
		return "", false
	}

	startSearch := pos + eq + 1

	start := strings.IndexByte(script[startSearch:], '{')
	if start < 0 {
		return "", false
	}

	start += startSearch

	end, ok := findMatchingJSONBrace(script, start)
	if !ok {
		return "", false
	}

	return script[start : end+1], true
}

func findMatchingJSONBrace(s string, start int) (int, bool) {
	if start < 0 || start >= len(s) || s[start] != '{' {
		return 0, false
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		ch := s[i]

		if inString {
			if escaped {
				escaped = false
				continue
			}

			switch ch {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}

			continue
		}

		switch ch {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i, true
			}
		}
	}

	return 0, false
}
