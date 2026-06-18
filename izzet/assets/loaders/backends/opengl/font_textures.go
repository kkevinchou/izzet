package opengl

import (
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
	"github.com/kkevinchou/izzet/izzet/assets/loaders/backends/sdl"
)

// NewFont uploads an SDL_ttf-rasterized atlas as an OpenGL texture.
func NewFont(fontFile string, size int) fonts.Font {
	atlas := sdl.RasterizeFontAtlas(fontFile, size)
	textTexture := NewFontTexture(atlas.Pixels, atlas.Width, atlas.Height)
	return fonts.Font{
		TextureID:   textTexture,
		Glyphs:      atlas.Glyphs,
		TotalWidth:  int(atlas.Width),
		TotalHeight: int(atlas.Height),
	}
}
