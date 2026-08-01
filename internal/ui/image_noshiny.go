//go:build !shiny
// +build !shiny

package ui

import "io"

// HasShiny indicates whether the shiny viewer is available in this build.
func HasShiny() bool {
	return false
}

// ShowShiny displays an image using the shiny viewer, if available.
func ShowShiny(r io.Reader, urlStr string) bool {
	return false
}
