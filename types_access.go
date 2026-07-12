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

// SetKeywords sets the meta keywords.
func (p *PageInfo) SetKeywords(keywords string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Keywords = keywords
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

// SetOGImageSecureURL sets the OpenGraph image secure URL.
func (p *PageInfo) SetOGImageSecureURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImageSecureURL = url
}

// SetOGImageType sets the OpenGraph image type.
func (p *PageInfo) SetOGImageType(t string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImageType = t
}

// SetOGImageWidth sets the OpenGraph image width.
func (p *PageInfo) SetOGImageWidth(w string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImageWidth = w
}

// SetOGImageHeight sets the OpenGraph image height.
func (p *PageInfo) SetOGImageHeight(h string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGImageHeight = h
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

// SetTwitterPlayer sets the Twitter player.
func (p *PageInfo) SetTwitterPlayer(player string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterPlayer = player
}

// SetTwitterPlayerWidth sets the Twitter player width.
func (p *PageInfo) SetTwitterPlayerWidth(width string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterPlayerWidth = width
}

// SetTwitterPlayerHeight sets the Twitter player height.
func (p *PageInfo) SetTwitterPlayerHeight(height string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.TwitterPlayerHeight = height
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

// SetShortLink sets the short link.
func (p *PageInfo) SetShortLink(link string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ShortLink = link
}

// SetAlternateHandheldLink sets the alternate handheld link.
func (p *PageInfo) SetAlternateHandheldLink(link string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.AlternateHandheldLink = link
}

// AddIcon adds an icon.
func (p *PageInfo) AddIcon(url, sizes string) {
	if url == "" {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, existing := range p.Icons {
		if existing.URL == url && existing.Sizes == sizes {
			return
		}
	}
	p.Icons = append(p.Icons, Icon{URL: url, Sizes: sizes})
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

// SetOGVideo adds an OpenGraph video.
func (p *PageInfo) SetOGVideo(vid string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGVideo = vid
}

// SetOGVideoSecureURL sets the OpenGraph video secure URL.
func (p *PageInfo) SetOGVideoSecureURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGVideoSecureURL = url
}

// SetOGVideoType sets the OpenGraph video type.
func (p *PageInfo) SetOGVideoType(t string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGVideoType = t
}

// SetOGVideoWidth sets the OpenGraph video width.
func (p *PageInfo) SetOGVideoWidth(w string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGVideoWidth = w
}

// SetOGVideoHeight sets the OpenGraph video height.
func (p *PageInfo) SetOGVideoHeight(h string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGVideoHeight = h
}

// SetOGVideoAlt sets the OpenGraph video alt text.
func (p *PageInfo) SetOGVideoAlt(alt string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.OGVideoAlt = alt
}
