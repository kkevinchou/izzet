package opengl

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/assets/fonts"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var supportedCharacters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ ()[]{}!@#$%^&*=+-_,./?:;'\""

func NewFont(fontFile string, size int) fonts.Font {
	ttfFont, err := ttf.OpenFont(fontFile, size)
	if err != nil {
		panic(err)
	}
	ttfFont.SetHinting(ttf.HINTING_NORMAL)
	ttfFont.SetStyle(ttf.STYLE_NORMAL)

	surface, err := ttfFont.RenderUTF8Blended(supportedCharacters, sdl.Color{R: 0, G: 21, B: 161, A: 255})
	if err != nil {
		panic(err)
	}

	textureWidth := int(surface.ClipRect.W)
	textureHeight := int(surface.ClipRect.H)
	pixels := make([]byte, len(surface.Pixels()))
	surfacePixels := surface.Pixels()

	index := 0
	for j := 0; j < int(textureHeight); j++ {
		for i := 0; i < int(textureWidth); i++ {
			b := surfacePixels[index]
			g := surfacePixels[index+1]
			r := surfacePixels[index+2]
			a := surfacePixels[index+3]

			flippedIndex := (textureHeight-j-1)*4*textureWidth + 4*i
			pixels[flippedIndex] = r
			pixels[flippedIndex+1] = g
			pixels[flippedIndex+2] = b
			pixels[flippedIndex+3] = a
			index += 4
		}
	}

	glyphs := map[string]fonts.Glyph{}
	widthPerGlyph := textureWidth / len(supportedCharacters)
	heightPerGlyph := textureHeight
	textureCoordWidth := float64(widthPerGlyph) / float64(textureWidth)
	for i, c := range supportedCharacters {
		glyphs[string(c)] = fonts.Glyph{
			TextureCoords: mgl64.Vec2{(float64(i) * textureCoordWidth), 0},
			Width:         widthPerGlyph,
			Height:        heightPerGlyph,
		}
	}

	textTexture := NewFontTexture(pixels, surface.ClipRect.W, surface.ClipRect.H)
	return fonts.Font{
		TextureID:   textTexture,
		Glyphs:      glyphs,
		TotalWidth:  textureWidth,
		TotalHeight: textureHeight,
	}
}
