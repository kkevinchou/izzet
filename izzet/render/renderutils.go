package render

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/render/context"
	"github.com/kkevinchou/izzet/izzet/render/rendersettings"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var pickingBuffer []byte

type TriangleVAO struct {
	VAO    uint32
	length int
}

type RenderData struct {
	Primitive   *modelspec.PrimitiveSpecification
	Transform   mgl32.Mat4
	VAO         uint32
	GeometryVAO uint32
}

type TextureFn func() (int, int, []uint32)

func RGBtoHSV(rgb mgl32.Vec3) mgl32.Vec3 {
	// Normalize RGB values to be between 0 and 1
	r := rgb.X()
	g := rgb.Y()
	b := rgb.Z()

	// Determine maximum and minimum values among R, G, and B
	maxVal := float32(math.Max(math.Max(float64(r), float64(g)), float64(b)))
	minVal := float32(math.Min(math.Min(float64(r), float64(g)), float64(b)))

	// Calculate value (V) as maximum of R, G, and B
	v := maxVal

	// Calculate saturation (S)
	var s float32
	if maxVal == 0 {
		s = 0
	} else {
		s = (maxVal - minVal) / maxVal
	}

	// Calculate hue (H)
	var h float32
	if maxVal == minVal {
		h = 0
	} else if maxVal == r && g >= b {
		h = 60 * (g - b) / (maxVal - minVal)
	} else if maxVal == r && g < b {
		h = 60*(g-b)/(maxVal-minVal) + 360
	} else if maxVal == g {
		h = 60*(b-r)/(maxVal-minVal) + 120
	} else { // maxVal == B
		h = 60*(r-g)/(maxVal-minVal) + 240
	}

	// Return HSV values as an mgl32.Vec3
	return mgl32.Vec3{h, s, v}
}

func HSVtoRGB(hsv mgl32.Vec3) mgl32.Vec3 {
	// Extract H, S, and V values from input Vec3
	h := hsv.X()
	s := hsv.Y()
	v := hsv.Z()

	// Calculate chroma (C)
	c := v * s

	// Calculate h' (hPrime)
	hPrime := h / 60

	// Calculate x
	x := c * float32(1-math.Abs(float64(math.Mod(float64(hPrime), 2)-1)))

	// Calculate m
	m := v - c

	// Initialize RGB values to m
	r := m
	g := m
	b := m

	// Determine which sector of the color wheel h' falls in and set RGB values accordingly
	if hPrime < 1 {
		r += c
		g += x
	} else if hPrime < 2 {
		r += x
		g += c
	} else if hPrime < 3 {
		g += c
		b += x
	} else if hPrime < 4 {
		g += x
		b += c
	} else if hPrime < 5 {
		r += x
		b += c
	} else {
		r += c
		b += x
	}

	// Create and return RGB Vec3
	return mgl32.Vec3{r, g, b}
}

func (r *RenderSystem) createCircleTexture(width, height int) (uint32, uint32) {
	circleTextureFn := textureFn(width, height, []int32{rendersettings.InternalTextureColorFormatRGBA}, []uint32{rendersettings.RenderFormatRGBA}, []uint32{gl.UNSIGNED_BYTE})
	fbo, textures := r.initFrameBuffer(circleTextureFn)
	return fbo, textures[0]
}

func textureFn(width int, height int, internalFormat []int32, format []uint32, xtype []uint32) func() (int, int, []uint32) {
	return func() (int, int, []uint32) {
		count := len(internalFormat)
		var textures []uint32
		for i := 0; i < count; i++ {
			texture := createTexture(width, height, internalFormat[i], format[i], xtype[i], gl.LINEAR)
			attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, attachment, gl.TEXTURE_2D, texture, 0)

			textures = append(textures, texture)
		}
		return width, height, textures
	}
}

func (r *RenderSystem) initFrameBuffer(tf TextureFn) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	width, height, textures := tf()
	var drawBuffers []uint32

	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	var rbo uint32
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT, int32(width), int32(height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func (r *RenderSystem) initFrameBufferNoDepth(tf TextureFn) (uint32, []uint32) {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	_, _, textures := tf()
	var drawBuffers []uint32

	textureCount := len(textures)
	for i := 0; i < textureCount; i++ {
		attachment := gl.COLOR_ATTACHMENT0 + uint32(i)
		drawBuffers = append(drawBuffers, attachment)
	}

	gl.DrawBuffers(int32(textureCount), &drawBuffers[0])

	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic(errors.New("failed to initalize frame buffer"))
	}

	return fbo, textures
}

func createTexture(width, height int, internalFormat int32, format uint32, xtype uint32, filtering int32) uint32 {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(gl.TEXTURE_2D, 0, internalFormat,
		int32(width), int32(height), 0, format, xtype, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, filtering)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return texture
}

func (r *RenderSystem) CameraViewerContext() context.ViewerContext {
	return r.cameraViewerContext
}

// NOTE: this method should only be called from within the render loop. if the frame
// buffer is swapped, then the data in the buffer can be undefined. so, we should make
// sure this is called in the render loop and before we swap frame buffers. that said,
// this might be handled automatically by the graphics driver so it may not actually
// be necessary.
//
// some changes that i've made to attempt crash fixes is moving the picking buffer into
// a package variable outside of the getEntityByPixelPosition method. in addition, i've
// done better vao caching for our gizmos which previously were recreated whenever the
// camera moves
func (r *RenderSystem) getEntityByPixelPosition(fbo uint32, pixelPosition mgl64.Vec2) *int {
	if r.app.Minimized() || !r.app.WindowFocused() {
		return nil
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	gl.ReadBuffer(gl.COLOR_ATTACHMENT1)
	defer gl.ReadBuffer(gl.COLOR_ATTACHMENT0)

	_, windowHeight := r.app.WindowSize()
	gl.PixelStorei(gl.PACK_ALIGNMENT, 1)

	if len(pickingBuffer) == 0 {
		pickingBuffer = make([]byte, 4)
	}

	var footerSize int32 = 0
	if r.app.RuntimeConfig().UIEnabled {
		footerSize = int32(apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled))
	}

	// in OpenGL, the mouse origin is the bottom left corner, so we need to offset by the footer size if it's present
	// SDL, on the other hand, has the mouse origin in the top left corner
	var weirdOffset float32 = -1 // Weirdge
	gl.ReadPixels(int32(pixelPosition[0]), int32(windowHeight)-int32(pixelPosition[1])-footerSize+int32(weirdOffset), 1, 1, gl.RED_INTEGER, gl.UNSIGNED_INT, gl.Ptr(pickingBuffer))

	uintID := binary.LittleEndian.Uint32(pickingBuffer)
	if uintID == settings.EmptyColorPickingID {
		return nil
	}

	id := int(uintID)
	return &id
}

func calculateFrustumPoints(position mgl64.Vec3, rotation mgl64.Quat, near, far, fovX, fovY, aspectRatio float64, nearPlaneOffset float64) []mgl64.Vec3 {
	viewerViewMatrix := rotation.Mat4()

	viewTranslationMatrix := mgl64.Translate3D(position.X(), position.Y(), position.Z())
	viewMatrix := viewTranslationMatrix.Mul4(viewerViewMatrix)

	halfY := math.Tan(mgl64.DegToRad(fovY / 2))
	halfX := math.Tan(mgl64.DegToRad(fovX / 2))

	var verts []mgl64.Vec3

	corners := []float64{-1, 1}
	nearFar := []float64{near, far}
	offsets := []float64{nearPlaneOffset, 0}

	for k, distance := range nearFar {
		for _, i := range corners {
			for _, j := range corners {
				vert := viewMatrix.Mul4x1(mgl64.Vec3{i * halfX * distance, j * halfY * distance, -distance + offsets[k]}.Vec4(1)).Vec3()
				verts = append(verts, vert)
			}
		}
	}

	return verts
}

func (r *RenderSystem) iztDrawArrays(first, count int32) {
	r.app.RuntimeConfig().TriangleDrawCount += int(count / 3)
	r.app.RuntimeConfig().DrawCount += 1
	gl.DrawArrays(gl.TRIANGLES, first, count)
}

func CalculateMenuBarHeight() float32 {
	style := imgui.CurrentStyle()
	return settings.FontSize + style.FramePadding().Y*2
}

func (r *RenderSystem) GameWindowSize() (int, int) {
	menuBarSize := CalculateMenuBarHeight()
	footerSize := apputils.CalculateFooterSize(r.app.RuntimeConfig().UIEnabled)

	windowWidth, windowHeight := r.app.WindowSize()

	width := windowWidth
	height := windowHeight - int(menuBarSize) - int(footerSize)

	if r.app.RuntimeConfig().UIEnabled {
		width = int(math.Ceil(float64(1-uiWidthRatio) * float64(windowWidth)))
	}

	return width, height
}

// returns the orthographic projection matrix for the directional light as well as the "position" of the light
func ComputeDirectionalLightProps(lightRotationMatrix mgl64.Mat4, frustumPoints []mgl64.Vec3, shadowMapZOffset float32) (mgl64.Vec3, mgl64.Mat4) {
	var lightSpacePoints []mgl64.Vec3
	invLightRotationMatrix := lightRotationMatrix.Inv()

	for _, point := range frustumPoints {
		lightSpacePoint := invLightRotationMatrix.Mul4x1(point.Vec4(1)).Vec3()
		lightSpacePoints = append(lightSpacePoints, lightSpacePoint)
	}

	var minX, maxX, minY, maxY, minZ, maxZ float64

	minX = lightSpacePoints[0].X()
	maxX = lightSpacePoints[0].X()
	minY = lightSpacePoints[0].Y()
	maxY = lightSpacePoints[0].Y()
	minZ = lightSpacePoints[0].Z()
	maxZ = lightSpacePoints[0].Z()

	for _, point := range lightSpacePoints {
		if point.X() < minX {
			minX = point.X()
		}
		if point.X() > maxX {
			maxX = point.X()
		}
		if point.Y() < minY {
			minY = point.Y()
		}
		if point.Y() > maxY {
			maxY = point.Y()
		}
		if point.Z() < minZ {
			minZ = point.Z()
		}
		if point.Z() > maxZ {
			maxZ = point.Z()
		}
	}
	maxZ += float64(shadowMapZOffset)

	halfX := (maxX - minX) / 2
	halfY := (maxY - minY) / 2
	halfZ := (maxZ - minZ) / 2
	position := mgl64.Vec3{minX + halfX, minY + halfY, maxZ}
	position = lightRotationMatrix.Mul4x1(position.Vec4(1)).Vec3() // bring position back into world space
	orthoProjMatrix := mgl64.Ortho(-halfX, halfX, -halfY, halfY, 0, halfZ*2)
	return position, orthoProjMatrix
}
