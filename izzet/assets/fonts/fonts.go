package fonts

import "github.com/go-gl/mathgl/mgl32"

type Glyph struct {
	TextureCoords mgl32.Vec2
	Width         int
	Height        int
}

type Font struct {
	TextureID   uint32
	TotalHeight int
	TotalWidth  int
	Glyphs      map[string]Glyph
}
