package opengl

import (
	"image"
	"image/draw"
	"log"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v4.1-core/gl"
)

type TextureInfo struct {
	Width  int32
	Height int32
	Data   []uint8
}

func ReadTextureInfo(file string) TextureInfo {
	imgFile, err := os.Open(file)
	defer func() {
		e := imgFile.Close()
		if e != nil {
			log.Fatalf("texture %q failed to close: %v\n", file, err)
		}
	}()

	if err != nil {
		log.Fatalf("texture %q not found on disk: %v\n", file, err)
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		panic(err)
	}

	// is vertically flipped if directly read into opengl texture
	nrgba := imaging.FlipV(img)

	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		panic("unsupported stride")
	}

	draw.Draw(rgba, rgba.Bounds(), nrgba, image.Point{0, 0}, draw.Src)

	size := rgba.Rect.Size()
	return TextureInfo{Width: int32(size.X), Height: int32(size.Y), Data: rgba.Pix}
}

func CreateOpenGLTexture(textureInfo TextureInfo) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.SRGB8_ALPHA8,
		textureInfo.Width,
		textureInfo.Height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(textureInfo.Data),
	)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)

	var maxAniso float32
	gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &maxAniso)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAX_ANISOTROPY, maxAniso)
	gl.GenerateMipmap(gl.TEXTURE_2D)

	return texture
}

func NewFontTexture(pixels []byte, width, height int32) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		width,
		height,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(pixels),
	)

	return texture
}
