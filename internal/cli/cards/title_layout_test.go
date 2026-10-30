package cards

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font/gofont/goregular"
)

func TestWrappedTitleBaseline(t *testing.T) {
	parsed, err := truetype.Parse(goregular.TTF)
	if err != nil {
		t.Fatal(err)
	}
	face := truetype.NewFace(parsed, &truetype.Options{Size: 20})
	img := image.NewRGBA(image.Rect(0, 0, 800, 400))
	short := drawTextWrap(img, 400, 60, "Short title", color.Black, 780, face)
	long := drawTextWrap(img, 400, 60, strings.Repeat("A long repository title ", 4), color.Black, 780, face)
	if short != 60 || long <= short+45 {
		t.Fatalf("short=%d long=%d; description must follow the measured last title baseline", short, long)
	}
	fallback := drawTextWrap(img, 400, 60, strings.Repeat("A long repository title ", 4), color.Black, 780, nil)
	if fallback <= 60 {
		t.Fatalf("fallback baseline=%d", fallback)
	}
}
