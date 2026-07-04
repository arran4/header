package header

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// PageInfo represents the extracted information from a URL.
type PageInfo struct {
	mu sync.Mutex

	// General Information
	URL   string
	Title string

	// OpenGraph
	OGTitle       string
	OGType        string
	OGImage       string
	OGDescription string
	OGImageAlt    string

	// Twitter
	TwitterCard        string
	TwitterImage       string
	TwitterTitle       string
	TwitterDescription string
	TwitterImageAlt    string

	// YouTube Specific
	YouTubeVideoID       string
	YouTubeLengthSeconds int
	YouTubeDuration      time.Duration

	// JSON-LD (First found entity or merged data)
	JSONLDType string
	JSONLDName string
}

// CoverImages returns a list of potential image URLs.
func (p *PageInfo) CoverImages() []string {
	var images []string
	if p.OGImage != "" {
		images = append(images, p.OGImage)
	}
	if p.TwitterImage != "" && p.TwitterImage != p.OGImage {
		images = append(images, p.TwitterImage)
	}
	return images
}

// GroupedProperty represents a dynamically grouped metadata field for output.
type GroupedProperty struct {
	Label string
	Value string
	Count int
	Order int
}

// GroupedProperties returns fields grouped by value and sorted by number of grouped sources.
func (p *PageInfo) GroupedProperties() []GroupedProperty {
	p.mu.Lock()
	defer p.mu.Unlock()

	type field struct {
		Source   string
		Label    string
		GroupKey string
		Value    interface{}
		Order    int
	}

	var fields []field
	var order int
	add := func(source, label, groupKey string, value interface{}) {
		if s, ok := value.(string); ok && s == "" {
			return
		}
		if d, ok := value.(fmt.Stringer); ok && d.String() == "0s" {
			return
		}
		fields = append(fields, field{source, label, groupKey, value, order})
		order++
	}

	add("OpenGraph", "Title", "Title", p.OGTitle)
	add("OpenGraph", "Type", "Type", p.OGType)
	add("OpenGraph", "Description", "Description", p.OGDescription)
	add("OpenGraph", "Image Alt", "ImageAlt", p.OGImageAlt)
	add("Twitter", "Title", "Title", p.TwitterTitle)
	add("Twitter", "Card", "Type", p.TwitterCard)
	add("Twitter", "Description", "Description", p.TwitterDescription)
	add("Twitter", "Image Alt", "ImageAlt", p.TwitterImageAlt)
	add("YouTube", "Video ID", "VideoID", p.YouTubeVideoID)
	if p.YouTubeDuration > 0 {
		add("YouTube", "Duration", "Duration", p.YouTubeDuration)
	}
	add("JSON-LD", "Type", "Type", p.JSONLDType)
	add("JSON-LD", "Name", "Name", p.JSONLDName)

	type groupKey struct {
		Key   string
		Value interface{}
	}

	groups := make(map[groupKey][]field)
	for _, f := range fields {
		k := groupKey{f.GroupKey, f.Value}
		groups[k] = append(groups[k], f)
	}

	var results []GroupedProperty

	for _, grpFields := range groups {
		sort.Slice(grpFields, func(i, j int) bool {
			return grpFields[i].Order < grpFields[j].Order
		})

		var sources []string
		var labels []string
		seenLabels := make(map[string]bool)

		for _, f := range grpFields {
			sources = append(sources, f.Source)
			if !seenLabels[f.Label] {
				labels = append(labels, f.Label)
				seenLabels[f.Label] = true
			}
		}

		var sourceStr string
		if len(sources) == 1 {
			sourceStr = sources[0]
		} else if len(sources) == 2 {
			sourceStr = sources[0] + " and " + sources[1]
		} else {
			sourceStr = strings.Join(sources[:len(sources)-1], ", ") + " and " + sources[len(sources)-1]
		}

		var labelStr string
		if len(labels) == 1 {
			labelStr = labels[0]
		} else {
			labelStr = strings.Join(labels, "/")
		}

		label := sourceStr + " " + labelStr

		results = append(results, GroupedProperty{
			Label: label,
			Value: fmt.Sprintf("%v", grpFields[0].Value),
			Count: len(grpFields),
			Order: grpFields[0].Order,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Count != results[j].Count {
			return results[i].Count > results[j].Count
		}
		return results[i].Order < results[j].Order
	})

	return results
}
