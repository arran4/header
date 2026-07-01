package header

import "time"

// Setters and Getters for thread-safe access

func (p *PageInfo) SetTitleIfEmpty(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.Title == "" {
		p.Title = title
	}
}

func (p *PageInfo) SetOGType(t string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGType = t
}

func (p *PageInfo) SetOGImage(img string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImage = img
}

func (p *PageInfo) SetOGDescription(desc string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGDescription = desc
}

func (p *PageInfo) SetTwitterCard(c string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterCard = c
}

func (p *PageInfo) SetTwitterImage(img string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterImage = img
}

func (p *PageInfo) SetTwitterTitle(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterTitle = title
	if p.Title == "" {
		p.Title = title
	}
}

func (p *PageInfo) SetTwitterDescription(desc string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterDescription = desc
}

func (p *PageInfo) SetJSONLDTypeIfEmpty(t string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.JSONLDType == "" {
		p.JSONLDType = t
	}
}

func (p *PageInfo) SetJSONLDNameIfEmpty(n string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.JSONLDName == "" {
		p.JSONLDName = n
	}
}

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
