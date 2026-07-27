//go:build !shiny
// +build !shiny

package ui

import "io"

func HasShiny() bool {
	return false
}

func ShowShiny(r io.Reader, urlStr string) bool {
	return false
}
