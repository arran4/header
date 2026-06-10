package header

import (
	"encoding/json"
	"strings"
	"strconv"
	"time"
)

type playerResponse struct {
	VideoDetails struct {
		LengthSeconds string `json:"lengthSeconds"`
		Title         string `json:"title"`
		VideoID       string `json:"videoId"`
	} `json:"videoDetails"`
}

// extractYouTubeInfo tries to parse YouTube specific data from scripts
func extractYouTubeInfo(scripts []string, info *PageInfo) {
	for _, script := range scripts {
		jsonText, ok := extractAssignedJSONObject(script, "ytInitialPlayerResponse")
		if !ok {
			continue
		}

		var pr playerResponse
		if err := json.Unmarshal([]byte(jsonText), &pr); err != nil {
			continue
		}

		if pr.VideoDetails.LengthSeconds != "" {
			info.YouTubeVideoID = pr.VideoDetails.VideoID
			if info.Title == "" {
				info.Title = pr.VideoDetails.Title
			}

			sec, err := strconv.Atoi(pr.VideoDetails.LengthSeconds)
			if err == nil {
				info.YouTubeLengthSeconds = sec
				info.YouTubeDuration = time.Duration(sec) * time.Second
			}
			return
		}
	}
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
