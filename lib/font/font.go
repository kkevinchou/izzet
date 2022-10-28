package font

import (
	"github.com/go-gl/mathgl/mgl64"
)

type Glyph struct {
	TextureCoords mgl64.Vec2
	Width         int
	Height        int
}

type Font struct {
	TextureID   uint32
	TotalHeight int
	TotalWidth  int
	Glyphs      map[string]Glyph
}
