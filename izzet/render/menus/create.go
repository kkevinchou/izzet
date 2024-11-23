package menus

import (
	"fmt"
	"os"
	"path/filepath"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/material"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/kkevinchou/kitolib/utils"
	"github.com/sqweek/dialog"
)

var materialIDGen int
var defaultMaterial = material.Material{
	ID: fmt.Sprintf("material-%d", materialIDGen),
	PBR: types.PBR{
		Roughness:        0.55,
		Metallic:         0,
		DiffuseIntensity: 1,
	},
}

var (
	WIPMaterial material.Material = defaultMaterial
	tableFlags  imgui.TableFlags  = imgui.TableFlagsBordersInnerV
)

var createMaterialModel bool

func create(app renderiface.App) {
	imgui.SetNextWindowSize(imgui.Vec2{X: 300})

	if imgui.BeginMenu("Create") {
		if imgui.MenuItemBool("Create Material") {
			createMaterialModel = true
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
			createStaticBatch(app)
		}
		imgui.EndMenu()
	}

	if createMaterialModel {
		createMaterial(app)
	}
}

var BATCH_CREATED bool
var BATCH_VAO uint32
var BATCH_NUM_VERTICES int32

func createStaticBatch(app renderiface.App) {
	var primitives []types.MeshHandle

	var modelMatrices []mgl32.Mat4
	for _, entity := range app.World().Entities() {
		if entity.MeshComponent == nil || !entity.Static {
			continue
		}

		meshHandle := entity.MeshComponent.MeshHandle
		primitives = append(primitives, meshHandle)

		modelMatrix := entities.WorldTransform(entity)
		modelMat := utils.Mat4F64ToF32(modelMatrix).Mul4(utils.Mat4F64ToF32(entity.MeshComponent.Transform))

		modelMatrices = append(modelMatrices, modelMat)
	}
	vao, num_vertices := app.AssetManager().CreateBatch(primitives, modelMatrices)
	BATCH_CREATED = true
	BATCH_VAO = vao
	BATCH_NUM_VERTICES = num_vertices
}

func createMaterial(app renderiface.App) {
	mat := &WIPMaterial
	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	imgui.OpenPopupStr("Create Material")
	if imgui.BeginPopupModalV("Create Material", nil, imgui.WindowFlagsAlwaysAutoResize) {
		imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("ID", func() {
			imgui.InputTextWithHint("##MaterialID", "", &mat.ID, imgui.InputTextFlagsNone, nil)
		}, true)

		panelutils.SetupRow("Diffuse", func() {
			imgui.ColorEdit3V("", &mat.PBR.Diffuse, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
		}, true)
		panelutils.SetupRow("Invisible", func() {
			imgui.Checkbox("", &mat.Invisible)
		}, true)

		panelutils.SetupRow("Diffuse Intensity", func() {
			imgui.SliderFloatV("", &mat.PBR.DiffuseIntensity, 1, 20, "%.1f", imgui.SliderFlagsNone)
		}, true)

		panelutils.SetupRow("Roughness", func() { imgui.SliderFloatV("", &mat.PBR.Roughness, 0, 1, "%.2f", imgui.SliderFlagsNone) }, true)
		panelutils.SetupRow("Metallic Factor", func() { imgui.SliderFloatV("", &mat.PBR.Metallic, 0, 1, "%.2f", imgui.SliderFlagsNone) }, true)
		panelutils.SetupRow("Texture", func() { imgui.LabelText("", mat.PBR.TextureName) }, true)

		if mat.PBR.TextureName != "" {
			t := app.AssetManager().GetTexture(mat.PBR.TextureName)
			texture := imgui.TextureID{Data: uintptr(t.ID)}
			size := imgui.Vec2{X: 50, Y: 50}
			imgui.ImageV(texture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
		}

		if imgui.Button("Import Texture") {
			// loading the asset
			d := dialog.File()
			currentDir, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			d = d.SetStartDir(filepath.Join(currentDir, "_assets"))
			d = d.Filter("PNG file", "png")

			assetFilePath, err := d.Load()
			if err != nil {
				if err != dialog.ErrCancelled {
					panic(err)
				}
			} else {
				i := 0
				baseFileName := apputils.NameFromAssetFilePath(assetFilePath)
				mat.PBR.TextureName = baseFileName
				mat.PBR.ColorTextureIndex = &i
				mat.PBR.DiffuseIntensity = 1
				mat.PBR.Diffuse = [3]float32{1, 1, 1}
				mat.PBR.Metallic = 0
				mat.PBR.Roughness = 0.55
			}
		}

		imgui.EndTable()
		if imgui.Button("Create Material") {
			app.CreateMaterial(*mat)
			materialIDGen++
			WIPMaterial = defaultMaterial
			WIPMaterial.ID = fmt.Sprintf("material-%d", materialIDGen)
			createMaterialModel = false
			imgui.CloseCurrentPopup()
		}
		imgui.SameLine()
		if imgui.Button("Cancel") {
			createMaterialModel = false
			imgui.CloseCurrentPopup()
		}
		imgui.EndPopup()
	}

}
