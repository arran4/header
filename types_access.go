package header

import "time"

// Setters and Getters for thread-safe access

// SetTitleIfEmpty sets the title if it is not already set.
func (p *PageInfo) SetTitleIfEmpty(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Title == "" {
		p.Title = title
	}
}

// SetOGTitle sets the OpenGraph title.
func (p *PageInfo) SetOGTitle(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGTitle = title
	if p.Title == "" {
		p.Title = title
	}
}

// SetOGType sets the OpenGraph type.
func (p *PageInfo) SetOGType(t string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGType = t
}

// SetOGImage adds an OpenGraph image.
func (p *PageInfo) SetOGImage(img string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImage = img
}

// SetOGImageAlt adds an OpenGraph image alt text.
func (p *PageInfo) SetOGImageAlt(alt string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImageAlt = alt
}

// SetOGDescription sets the OpenGraph description.
func (p *PageInfo) SetOGDescription(desc string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGDescription = desc
}

// SetTwitterCard sets the Twitter card type.
func (p *PageInfo) SetTwitterCard(c string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterCard = c
}

// SetTwitterImage adds a Twitter image.
func (p *PageInfo) SetTwitterImage(img string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterImage = img
}

// SetTwitterImageAlt adds a Twitter image alt text.
func (p *PageInfo) SetTwitterImageAlt(alt string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterImageAlt = alt
}

// SetTwitterTitle sets the Twitter title.
func (p *PageInfo) SetTwitterTitle(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterTitle = title
	if p.Title == "" {
		p.Title = title
	}
}

// SetTwitterDescription sets the Twitter description.
func (p *PageInfo) SetTwitterDescription(desc string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterDescription = desc
}

// SetCanonicalLink sets the canonical link.
func (p *PageInfo) SetCanonicalLink(link string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.CanonicalLink = link
}

// SetJSONLDTypeIfEmpty sets the JSON-LD type if it is not already set.
func (p *PageInfo) SetJSONLDTypeIfEmpty(t string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.JSONLDType == "" {
		p.JSONLDType = t
	}
}

// SetJSONLDNameIfEmpty sets the JSON-LD name if it is not already set.
func (p *PageInfo) SetJSONLDNameIfEmpty(n string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.JSONLDName == "" {
		p.JSONLDName = n
	}
}

// SetYouTubeInfo sets the YouTube video information.
func (p *PageInfo) SetYouTubeInfo(id string, title string, lengthSecs int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.YouTubeVideoID = id
	if p.Title == "" && title != "" {
		p.Title = title
	}
	p.YouTubeLengthSeconds = lengthSecs
	p.YouTubeDuration = time.Duration(lengthSecs) * time.Second
}
