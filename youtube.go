package header

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type youtubeExtractor struct{}

func (e youtubeExtractor) Name() string { return "YouTube" }

func (e youtubeExtractor) Init(info *PageInfo) Processor {
	if !isYouTubeURL(info.URL) {
		return &noopProcessor{}
	}
	return &youtubeProcessor{info: info}
}

type noopProcessor struct{}

func (p *noopProcessor) ProcessToken(_ html.Token) bool { return true }

type youtubeProcessor struct {
	info     *PageInfo
	inScript bool
}

func (p *youtubeProcessor) ProcessToken(t html.Token) bool {
	if t.Type == html.StartTagToken {
		if t.Data == "script" {
			var isJSONLD bool
			for _, attr := range t.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" {
					isJSONLD = true
					break
				}
			}
			if !isJSONLD {
				p.inScript = true
			}
		} else if t.Data == "body" {
			return true // No player response script inside body needed usually, but could be safe to not break on body.
		}
	} else if t.Type == html.TextToken && p.inScript {
		script := t.Data
		jsonText, ok := extractAssignedJSONObject(script, "ytInitialPlayerResponse")
		if ok {
			var pr playerResponse
			if err := json.Unmarshal([]byte(jsonText), &pr); err == nil {
				if pr.VideoDetails.LengthSeconds != "" {
					sec, _ := strconv.Atoi(pr.VideoDetails.LengthSeconds)
					p.info.Mutate(func(i *PageInfo) {
						i.YouTubeVideoID = pr.VideoDetails.VideoID
						if i.Title == "" && pr.VideoDetails.Title != "" {
							i.Title = pr.VideoDetails.Title
						}
						i.YouTubeLengthSeconds = sec
						i.YouTubeDuration = time.Duration(sec) * time.Second
					})
					return true // We found what we need, stop processing
				}
			}
		}
	} else if t.Type == html.EndTagToken {
		if t.Data == "script" {
			p.inScript = false
		}
	}
	return false
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
