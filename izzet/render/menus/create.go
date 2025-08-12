package menus

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/windows"
)

func create(app renderiface.App) {
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})

	if imgui.BeginMenu("Create") {
		if imgui.MenuItemBool("Create Material") {
			windows.ShowCreateMaterialWindow(app)
		}
		if imgui.MenuItemBool("Build Nav Mesh") {
			runtimeConfig := app.RuntimeConfig()
			iterations := int(runtimeConfig.NavigationMeshIterations)
			walkableHeight := int(runtimeConfig.NavigationMeshWalkableHeight)
			climbableHeight := int(runtimeConfig.NavigationMeshClimbableHeight)
			minRegionArea := int(runtimeConfig.NavigationMeshMinRegionArea)
			sampleDist := float64(runtimeConfig.NavigationmeshSampleDist)
			maxError := float64(runtimeConfig.NavigationmeshMaxError)
			app.BuildNavMesh(app, iterations, walkableHeight, climbableHeight, minRegionArea, sampleDist, maxError)
		}
		if imgui.MenuItemBool("Bake Static Geometry") {
			app.SetupBatchedStaticRendering()
			app.RuntimeConfig().BatchRenderingEnabled = true
		}
		imgui.EndMenu()
	}
}
