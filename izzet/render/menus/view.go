package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type DebugComboOption string

const (
	ComboOptionFinalRender     DebugComboOption = "FINALRENDER"
	ComboOptionColorPicking    DebugComboOption = "COLORPICKING"
	ComboOptionPreBloomHDR     DebugComboOption = "PRE BLOOM HDR (bloom only)"
	ComboOptionBloom           DebugComboOption = "BLOOMTEXTURE (bloom only)"
	ComboOptionShadowDepthMap  DebugComboOption = "SHADOW DEPTH MAP"
	ComboOptionCameraDepthMap  DebugComboOption = "CAMERA DEPTH MAP"
	ComboOptionCubeDepthMap    DebugComboOption = "CUBE DEPTH MAP"
	ComboOptionVolumetric      DebugComboOption = "VOLUMETRIC"
	ComboOptionSSAO            DebugComboOption = "SSAO"
	ComboOptionGBufferPosition DebugComboOption = "GBUFFER - POSITION"
	ComboOptionGBufferNormal   DebugComboOption = "GBUFFER - NORMAL"
	ComboOptionGBufferDebug    DebugComboOption = "GBUFFER - DEBUG"
)

var SelectedDebugComboOption DebugComboOption = ComboOptionSSAO

var (
	DebugComboOptions []DebugComboOption = []DebugComboOption{
		ComboOptionFinalRender,
		ComboOptionColorPicking,
		ComboOptionPreBloomHDR,
		ComboOptionBloom,
		ComboOptionShadowDepthMap,
		ComboOptionCameraDepthMap,
		ComboOptionVolumetric,
		ComboOptionSSAO,
		ComboOptionGBufferPosition,
		ComboOptionGBufferNormal,
		ComboOptionGBufferDebug,
	}
)

func view(app renderiface.App, renderContext RenderContext) {
	runtimeConfig := app.RuntimeConfig()
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})
	if imgui.BeginMenu("View") {
		if imgui.MenuItemBoolV("Show Colliders", "", runtimeConfig.ShowColliders, true) {
			runtimeConfig.ShowColliders = !runtimeConfig.ShowColliders
		}

		if imgui.MenuItemBoolV("Show Selection Bounding Box", "", runtimeConfig.ShowSelectionBoundingBox, true) {
			runtimeConfig.ShowSelectionBoundingBox = !runtimeConfig.ShowSelectionBoundingBox
		}

		if imgui.MenuItemBoolV("Show UI", "", runtimeConfig.UIEnabled, true) {
			app.ConfigureUI(!runtimeConfig.UIEnabled)
		}

		if imgui.MenuItemBoolV("ShowImguiDemo", "", runtimeConfig.ShowImguiDemo, true) {
			runtimeConfig.ShowImguiDemo = !runtimeConfig.ShowImguiDemo
		}

		if imgui.MenuItemBoolV("Show Spatial Partition", "", runtimeConfig.RenderSpatialPartition, true) {
			runtimeConfig.RenderSpatialPartition = !runtimeConfig.RenderSpatialPartition
		}

		if imgui.MenuItemBoolV("Show Texture Viewer", "", runtimeConfig.ShowTextureViewer, true) {
			runtimeConfig.ShowTextureViewer = !runtimeConfig.ShowTextureViewer
		}

		imgui.EndMenu()
	}
}
