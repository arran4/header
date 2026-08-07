package cards

import (
	"fmt"
	"image"
	"sort"
	"strings"

	"github.com/arran4/header"
)

// Generator defines an interface for a card generator
type Generator interface {
	Name() string
	Description() string
	Generate(info *header.PageInfo) (image.Image, error)
}

var registry = make(map[string]Generator)

// Register adds a new card generator to the registry
func Register(g Generator) {
	registry[strings.ToLower(g.Name())] = g
}

// Get retrieves a card generator by name
func Get(name string) (Generator, error) {
	if name == "" {
		name = "basic"
	}
	g, ok := registry[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("card generator %q not found", name)
	}
	return g, nil
}

// List returns a sorted slice of available card generators
func List() []Generator {
	var list []Generator
	for _, g := range registry {
		list = append(list, g)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].Name() < list[j].Name()
	})
	return list
}
