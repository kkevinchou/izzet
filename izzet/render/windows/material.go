package windows

import (
	"os"
	"path/filepath"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/sqweek/dialog"
)

var (
	activeMaterial     *assets.MaterialAsset
	isCreatingMaterial bool
	backupMaterial     assets.MaterialAsset
	materialWindow     string
)

var showCreateMaterialModel bool

func ShowCreateMaterialWindow() {
	showCreateMaterialModel = true
	isCreatingMaterial = true
	materialWindow = "Create Material"
	assignDefaultMaterial()
}

func ShowEditMaterialWindow(material assets.MaterialAsset) {
	showCreateMaterialModel = true
	isCreatingMaterial = false
	materialWindow = "Edit Material"
	backupMaterial = material
	activeMaterial = &material
}

func init() {
	assignDefaultMaterial()
}

func renderMaterialWindow(app renderiface.App) {
	if !showCreateMaterialModel {
		return
	}

	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	if imgui.BeginV(materialWindow, &showCreateMaterialModel, imgui.WindowFlagsNone) {
		imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		var materialUpdated bool

		panelutils.SetupRow("Name", func() {
			if imgui.InputTextWithHint("##Name", "", &activeMaterial.Name, imgui.InputTextFlagsNone, nil) {
				materialUpdated = true
			}
		}, true)

		panelutils.SetupRow("Diffuse", func() {
			// color := [3]float32{}
			var color [3]float32 = activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorFactor.Vec3()
			if imgui.ColorEdit3V("", &color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel) {
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[0] = color[0]
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[1] = color[1]
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[2] = color[2]
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[3] = 1
				materialUpdated = true
			}
		}, true)

		panelutils.SetupRow("Roughness", func() {
			if imgui.SliderFloatV("", &activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.RoughnessFactor, 0, 1, "%.2f", imgui.SliderFlagsNone) {
				materialUpdated = true
			}
		}, true)
		panelutils.SetupRow("Metallic Factor", func() {
			if imgui.SliderFloatV("", &activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.MetalicFactor, 0, 1, "%.2f", imgui.SliderFlagsNone) {
				materialUpdated = true
			}
		}, true)
		panelutils.SetupRow("Texture", func() {
			imgui.LabelText("", activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName)
		}, true)

		if activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName != "" {
			t := app.AssetManager().GetTexture(activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName)
			texture := imgui.TextureID{Data: uintptr(t.ID)}
			size := imgui.Vec2{X: 150, Y: 150}
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
				// i := 0
				baseFileName := apputils.NameFromAssetFilePath(assetFilePath)
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName = baseFileName
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.RoughnessFactor = 0.55
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.MetalicFactor = 0
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorFactor = mgl32.Vec4{1, 1, 1, 1}
				materialUpdated = true
			}
		}
		imgui.SameLine()
		if imgui.Button("Remove Texture") {
			activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName = ""
			materialUpdated = true
		}

		if materialUpdated {
			if !isCreatingMaterial {
				app.AssetManager().UpdateMaterialAsset(*activeMaterial)
			}
		}

		imgui.EndTable()

		if isCreatingMaterial {
			if imgui.Button("Create Material") {
				app.AssetManager().CreateMaterial(activeMaterial.Name, activeMaterial.Material)
				showCreateMaterialModel = false
				assignDefaultMaterial()
			}
			imgui.SameLine()
		}

		if imgui.Button("Cancel") {
			showCreateMaterialModel = false
			if !isCreatingMaterial {
				app.AssetManager().UpdateMaterialAsset(backupMaterial)
			}
			assignDefaultMaterial()
		}
	}
	imgui.End()
}

func assignDefaultMaterial() {
	activeMaterial = &assets.MaterialAsset{
		Material: modelspec.MaterialSpecification{
			PBRMaterial: modelspec.PBRMaterial{
				PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
					BaseColorFactor: mgl32.Vec4{1, 1, 1, 1},
					RoughnessFactor: 0.55,
					MetalicFactor:   1,
				},
			},
		},
	}
}
