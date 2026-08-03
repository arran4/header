package header

import "sync"

// KnownMetaTracker tracks all seen and known metadata tags during processing.
type KnownMetaTracker struct {
	mu sync.Mutex

	RawTags   []UnknownMeta
	KnownKeys map[string]bool
}

// AddRaw adds a raw metadata tag that has been seen.
func (t *KnownMetaTracker) AddRaw(key, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.RawTags = append(t.RawTags, UnknownMeta{Key: key, Value: value})
}

// MarkKnown registers a tag key as explicitly handled by a processor.
func (t *KnownMetaTracker) MarkKnown(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.KnownKeys == nil {
		t.KnownKeys = make(map[string]bool)
	}
	t.KnownKeys[key] = true
}

// ProcessUnknown returns a list of tags that are raw but have not been marked known.
func (t *KnownMetaTracker) ProcessUnknown(excludedKeys []string) []UnknownMeta {
	t.mu.Lock()
	defer t.mu.Unlock()

	excluded := make(map[string]bool)
	for _, key := range excludedKeys {
		excluded[key] = true
	}

	var unknown []UnknownMeta
	for _, tag := range t.RawTags {
		if !t.KnownKeys[tag.Key] && !excluded[tag.Key] {
			unknown = append(unknown, tag)
		}
	}
	return unknown
}
