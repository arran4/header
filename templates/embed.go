// Package templates contains HTML templates for output rendering.
package templates

import "embed"

// FS contains the embedded template files.
//
//go:embed *.tmpl
var FS embed.FS
