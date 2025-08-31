package opengl

import (
	"image"
	"log"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
)

type TextureInfo struct {
	Width  int32
	Height int32
	Data   []uint8
}

func ReadTextureInfo(file string) TextureInfo {
	f, err := os.Open(file)
	if err != nil {
		log.Fatalf("texture %q not found on disk: %v\n", file, err)
	}
	defer func() {
		if e := f.Close(); e != nil {
			log.Printf("warning: texture %q failed to close: %v\n", file, e)
		}
	}()

	// Decode (this returns an image.Image; we’ll keep it straight alpha)
	src, _, err := image.Decode(f)
	if err != nil {
		log.Fatalf("decode failed for %q: %v", file, err)
	}

	// Flip vertically. imaging functions return *image.NRGBA (straight alpha).
	flipped := imaging.FlipV(src) // *image.NRGBA

	// Ensure tight stride (w*4). If not, repack row-by-row.
	b := flipped.Bounds()
	w, h := b.Dx(), b.Dy()
	wantStride := w * 4

	var pix []byte
	if flipped.Stride == wantStride {
		// Already tightly packed; we can use the backing slice directly.
		// NOTE: The Pix slice may be larger than w*h*4; slice it to exactly that size.
		pix = flipped.Pix[:wantStride*h]
	} else {
		// Repack into tightly-packed NRGBA buffer
		pix = make([]byte, wantStride*h)
		for y := 0; y < h; y++ {
			srcOff := flipped.PixOffset(b.Min.X, b.Min.Y+y)
			dstOff := y * wantStride
			copy(pix[dstOff:dstOff+wantStride], flipped.Pix[srcOff:srcOff+flipped.Stride])
		}
	}

	return TextureInfo{
		Width:  int32(w),
		Height: int32(h),
		Data:   pix, // straight alpha bytes (NRGBA)
	}
}

// TODO : RGB values don't look right based on what renderdoc sees. much darker in renderdoc
// than what it should be based on color picking from gimp. top right corner
func CreateOpenGLTexture(textureInfo TextureInfo) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		rendersettings.InternalTextureColorFormatSRGBA,
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
