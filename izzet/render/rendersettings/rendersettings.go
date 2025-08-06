package rendersettings

import "github.com/go-gl/gl/v4.1-core/gl"

const (
	// mipsCount             int = 6
	// MaxBloomTextureWidth  int = 1920
	// MaxBloomTextureHeight int = 1080

	// materialTextureWidth  int32 = 512
	// materialTextureHeight int32 = 512
	// this internal type should support floats in order for us to store HDR values for bloom
	// could change this to gl.RGB16F or gl.RGB32F for less color banding if we want
	RenderFormatRGB uint32 = gl.RGB
	// RenderFormatRGBA uint32 = gl.RGBA
	// internalTextureColorFormatRGB    int32  = gl.RGB32F
	// internalTextureColorFormatRGBA   int32  = gl.RGBA32F
	InternalTextureColorFormat16RGBA int32 = gl.RGBA16F

	// uiWidthRatio float32 = 0.2
)
