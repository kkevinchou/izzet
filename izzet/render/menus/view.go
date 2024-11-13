package menus

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type DebugComboOption string

const (
	ComboOptionFinalRender    DebugComboOption = "FINALRENDER"
	ComboOptionColorPicking   DebugComboOption = "COLORPICKING"
	ComboOptionHDR            DebugComboOption = "HDR (bloom only)"
	ComboOptionBloom          DebugComboOption = "BLOOMTEXTURE (bloom only)"
	ComboOptionShadowDepthMap DebugComboOption = "SHADOW DEPTH MAP"
	ComboOptionCameraDepthMap DebugComboOption = "CAMERA DEPTH MAP"
	ComboOptionCubeDepthMap   DebugComboOption = "CUBE DEPTH MAP"
	ComboOptionVolumetric     DebugComboOption = "VOLUMETRIC"
	ComboOptionKuwahara       DebugComboOption = "KUWAHARA"
)

var SelectedDebugComboOption DebugComboOption = ComboOptionFinalRender

var (
	DebugComboOptions []DebugComboOption = []DebugComboOption{
		ComboOptionFinalRender,
		ComboOptionColorPicking,
		ComboOptionHDR,
		ComboOptionBloom,
		ComboOptionShadowDepthMap,
		ComboOptionCameraDepthMap,
		ComboOptionVolumetric,
		ComboOptionKuwahara,
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

		if imgui.MenuItemBoolV("Show Debug Texture", "", runtimeConfig.ShowDebugTexture, true) {
			runtimeConfig.ShowDebugTexture = !runtimeConfig.ShowDebugTexture

		}

		imgui.EndMenu()
	}
}
