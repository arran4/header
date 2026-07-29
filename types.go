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
	URL         string   `json:",omitempty"`
	Trace       []string `json:",omitempty"`
	MetaRefresh string   `json:",omitempty"`
	Title       string   `json:",omitempty"`
	Keywords    string   `json:",omitempty"`

	// OpenGraph
	OGTitle          string `json:",omitempty"`
	OGType           string `json:",omitempty"`
	OGImage          string `json:",omitempty"`
	OGImageSecureURL string `json:",omitempty"`
	OGImageType      string `json:",omitempty"`
	OGImageWidth     string `json:",omitempty"`
	OGImageHeight    string `json:",omitempty"`
	OGDescription    string `json:",omitempty"`
	OGImageAlt       string `json:",omitempty"`
	OGVideo          string `json:",omitempty"`
	OGVideoSecureURL string `json:",omitempty"`
	OGVideoType      string `json:",omitempty"`
	OGVideoWidth     string `json:",omitempty"`
	OGVideoHeight    string `json:",omitempty"`
	OGVideoAlt       string `json:",omitempty"`

	// Twitter
	TwitterCard         string `json:",omitempty"`
	TwitterImage        string `json:",omitempty"`
	TwitterTitle        string `json:",omitempty"`
	TwitterDescription  string `json:",omitempty"`
	TwitterImageAlt     string `json:",omitempty"`
	TwitterPlayer       string `json:",omitempty"`
	TwitterPlayerWidth  string `json:",omitempty"`
	TwitterPlayerHeight string `json:",omitempty"`

	// YouTube Specific
	YouTubeVideoID       string        `json:",omitempty"`
	YouTubeLengthSeconds int           `json:",omitempty"`
	YouTubeDuration      time.Duration `json:",omitempty"`

	// App Links
	AppLinksIOSName        string `json:",omitempty"`
	AppLinksIOSAppStoreID  string `json:",omitempty"`
	AppLinksIOSURL         string `json:",omitempty"`
	AppLinksAndroidName    string `json:",omitempty"`
	AppLinksAndroidPackage string `json:",omitempty"`
	AppLinksAndroidURL     string `json:",omitempty"`

	// Instagram Specific
	InstagramOwnerUserID string `json:",omitempty"`

	// HTML Application
	ApplicationTitle string `json:",omitempty"`
	ThemeColor       string `json:",omitempty"`

	// JSON-LD (First found entity or merged data)
	JSONLDType string `json:",omitempty"`
	JSONLDName string `json:",omitempty"`

	// HTML Links
	CanonicalLink  string          `json:",omitempty"`
	ShortLink      string          `json:",omitempty"`
	AlternateLinks []AlternateLink `json:",omitempty"`
	FeedLinks      []FeedLink      `json:",omitempty"`
	Icons          []Icon          `json:",omitempty"`
}

// FeedLink represents a parsed RSS, Atom or JSON feed link.
type FeedLink struct {
	URL   string `json:",omitempty"`
	Title string `json:",omitempty"`
	Type  string `json:",omitempty"`
}

// AlternateLink represents a parsed alternate link.
type AlternateLink struct {
	URL   string `json:",omitempty"`
	Media string `json:",omitempty"`
}

// Icon represents a parsed icon link.
type Icon struct {
	URL   string `json:",omitempty"`
	Sizes string `json:",omitempty"`
}

// CoverImages returns a list of potential image URLs.
func (p *PageInfo) CoverImages() []string {
	var images []string
	seen := make(map[string]bool)

	add := func(img string) {
		if img != "" && !seen[img] {
			seen[img] = true
			images = append(images, img)
		}
	}

	add(p.OGImage)
	if p.OGImage == "" {
		add(p.OGImageSecureURL)
	}
	add(p.TwitterImage)

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

	// We only show Canonical Link if it differs from the requested URL.
	// Otherwise, we don't add it as a separate property.
	if p.CanonicalLink != "" && p.CanonicalLink != p.URL {
		add("HTML", "Canonical Link", "CanonicalLink", p.CanonicalLink)
	}
	if p.ShortLink != "" && p.ShortLink != p.URL {
		add("HTML", "Short Link", "ShortLink", p.ShortLink)
	}
	for _, link := range p.AlternateLinks {
		val := link.URL
		if link.Media != "" {
			val = fmt.Sprintf("%s (%s)", link.URL, link.Media)
		}
		add("HTML", "Alternate Link", "AlternateLink", val)
	}
	for _, link := range p.FeedLinks {
		val := link.URL
		var extras []string
		if link.Title != "" {
			extras = append(extras, fmt.Sprintf("%q", link.Title))
		}
		if link.Type != "" {
			extras = append(extras, link.Type)
		}
		if len(extras) > 0 {
			val = fmt.Sprintf("%s (%s)", link.URL, strings.Join(extras, ", "))
		}
		add("HTML", "Feed Link", "FeedLink", val)
	}
	for _, icon := range p.Icons {
		val := icon.URL
		if icon.Sizes != "" {
			val = fmt.Sprintf("%s (%s)", icon.URL, icon.Sizes)
		}
		add("HTML", "Icon", "Icon", val)
	}

	add("HTML", "Keywords", "Keywords", p.Keywords)
	add("HTML", "Application Title", "ApplicationTitle", p.ApplicationTitle)
	add("HTML", "Theme Color", "ThemeColor", p.ThemeColor)
	add("AppLinks", "iOS Name", "AppLinksIOSName", p.AppLinksIOSName)
	add("AppLinks", "iOS App Store ID", "AppLinksIOSAppStoreID", p.AppLinksIOSAppStoreID)
	add("AppLinks", "iOS URL", "AppLinksIOSURL", p.AppLinksIOSURL)
	add("AppLinks", "Android Name", "AppLinksAndroidName", p.AppLinksAndroidName)
	add("AppLinks", "Android Package", "AppLinksAndroidPackage", p.AppLinksAndroidPackage)
	add("AppLinks", "Android URL", "AppLinksAndroidURL", p.AppLinksAndroidURL)
	add("Instagram", "Owner User ID", "InstagramOwnerUserID", p.InstagramOwnerUserID)
	add("OpenGraph", "Title", "Title", p.OGTitle)
	add("OpenGraph", "Type", "Type", p.OGType)
	add("OpenGraph", "Image", "Image", p.OGImage)
	add("OpenGraph", "Image Secure URL", "ImageSecureURL", p.OGImageSecureURL)
	add("OpenGraph", "Image Type", "ImageType", p.OGImageType)
	add("OpenGraph", "Image Width", "ImageWidth", p.OGImageWidth)
	add("OpenGraph", "Image Height", "ImageHeight", p.OGImageHeight)
	add("OpenGraph", "Description", "Description", p.OGDescription)
	add("OpenGraph", "Image Alt", "ImageAlt", p.OGImageAlt)
	add("OpenGraph", "Video", "Video", p.OGVideo)
	add("OpenGraph", "Video Secure URL", "VideoSecureURL", p.OGVideoSecureURL)
	add("OpenGraph", "Video Type", "VideoType", p.OGVideoType)
	add("OpenGraph", "Video Width", "VideoWidth", p.OGVideoWidth)
	add("OpenGraph", "Video Height", "VideoHeight", p.OGVideoHeight)
	add("OpenGraph", "Video Alt", "VideoAlt", p.OGVideoAlt)
	add("Twitter", "Title", "Title", p.TwitterTitle)
	add("Twitter", "Card", "Type", p.TwitterCard)
	add("Twitter", "Description", "Description", p.TwitterDescription)
	add("Twitter", "Image Alt", "ImageAlt", p.TwitterImageAlt)
	add("Twitter", "Player", "Player", p.TwitterPlayer)
	add("Twitter", "Player Width", "PlayerWidth", p.TwitterPlayerWidth)
	add("Twitter", "Player Height", "PlayerHeight", p.TwitterPlayerHeight)
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

// Mutate executes the given function while holding the PageInfo mutex lock,
// allowing thread-safe modifications to the PageInfo fields.
func (p *PageInfo) Mutate(fn func(*PageInfo)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fn(p)
}
