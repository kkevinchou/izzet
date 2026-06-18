package sdl

import (
	sdllib "github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/assets/fonts"
)

var supportedCharacters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ ()[]{}!@#$%^&*=+-_,./?:;'\""

type FontAtlas struct {
	Pixels []byte
	Width  int32
	Height int32
	Glyphs map[string]fonts.Glyph
}

func RasterizeFontAtlas(fontFile string, size int) FontAtlas {
	ttfFont, err := ttf.OpenFont(fontFile, float32(size))
	if err != nil {
		panic(err)
	}
	defer ttfFont.Close()
	ttfFont.SetHinting(ttf.HINTING_NORMAL)
	ttfFont.SetStyle(ttf.FontStyleFlags(0))

	surface, err := ttfFont.RenderTextBlended(supportedCharacters, sdllib.Color{R: 0, G: 21, B: 161, A: 255})
	if err != nil {
		panic(err)
	}
	defer surface.Destroy()

	textureWidth := int(surface.W)
	textureHeight := int(surface.H)
	pixels := make([]byte, textureWidth*textureHeight*4)

	for j := 0; j < textureHeight; j++ {
		for i := 0; i < textureWidth; i++ {
			r, g, b, a, err := surface.ReadPixel(int32(i), int32(j))
			if err != nil {
				panic(err)
			}

			flippedIndex := (textureHeight-j-1)*4*textureWidth + 4*i
			pixels[flippedIndex] = r
			pixels[flippedIndex+1] = g
			pixels[flippedIndex+2] = b
			pixels[flippedIndex+3] = a
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

	return FontAtlas{
		Pixels: pixels,
		Width:  surface.W,
		Height: surface.H,
		Glyphs: glyphs,
	}
}
