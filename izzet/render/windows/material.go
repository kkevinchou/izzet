package windows

import (
	"os"
	"path/filepath"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/sqweek/dialog"
)

var (
	activeMaterial     assets.MaterialAsset
	isCreatingMaterial bool
	backupMaterial     assets.MaterialAsset
	materialWindow     string
)

const (
	defaultMaterialName string = "<my new material>"
)

func ShowCreateMaterialWindow(app renderiface.App) {
	app.RuntimeConfig().ShowMaterialEditor = true
	isCreatingMaterial = true
	materialWindow = "Create Material"
	assignDefaultMaterial()
}

func ShowEditMaterialWindow(app renderiface.App, material assets.MaterialAsset) {
	app.RuntimeConfig().ShowMaterialEditor = true
	isCreatingMaterial = false
	materialWindow = "Edit Material"
	backupMaterial = material
	activeMaterial = material
}

func init() {
	assignDefaultMaterial()
}

func renderMaterialWindow(app renderiface.App) {
	if !app.RuntimeConfig().ShowMaterialEditor {
		return
	}

	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

	if imgui.BeginV(materialWindow, &app.RuntimeConfig().ShowMaterialEditor, imgui.WindowFlagsNone) {
		imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		var materialUpdated bool

		panelutils.SetupRow("Name", func() {
			if imgui.InputTextWithHint("##Name", "", &activeMaterial.Name, imgui.InputTextFlagsNone, nil) {
				materialUpdated = true
			}
		}, true)

		panelutils.SetupRow("Diffuse", func() {
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
			texture := imgui.TextureID(t.ID)
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
				activeMaterial.Material.PBRMaterial.PBRMetallicRoughness.RoughnessFactor = 1
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
				app.AssetManager().UpdateMaterialAsset(activeMaterial)
			}
		}

		imgui.EndTable()

		if isCreatingMaterial {
			if imgui.Button("Save") {
				if activeMaterial.Name != "" {
					newMaterialHandle := app.AssetManager().CreateCustomMaterial(activeMaterial.Name, activeMaterial.Material)
					app.QueueCreateMaterialTexture(newMaterialHandle)
					app.RuntimeConfig().ShowMaterialEditor = false
					assignDefaultMaterial()
				} else {
					activeMaterial.Name = defaultMaterialName
				}
			}
		} else {
			if imgui.Button("Save") {
				app.RuntimeConfig().ShowMaterialEditor = false
				app.QueueCreateMaterialTexture(activeMaterial.Handle)
			}
		}
		imgui.SameLine()

		if imgui.Button("Restore") {
			app.AssetManager().UpdateMaterialAsset(backupMaterial)
			activeMaterial = backupMaterial
		}
		imgui.SameLine()

		if imgui.Button("Cancel") {
			app.RuntimeConfig().ShowMaterialEditor = false
			if !isCreatingMaterial {
				app.AssetManager().UpdateMaterialAsset(backupMaterial)
			}
			assignDefaultMaterial()
		}
	}
	imgui.End()
}

func assignDefaultMaterial() {
	activeMaterial = assets.MaterialAsset{
		Name: defaultMaterialName,
		Material: modelspec.MaterialSpecification{
			PBRMaterial: modelspec.PBRMaterial{
				PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
					BaseColorFactor: mgl32.Vec4{1, 1, 1, 1},
					RoughnessFactor: 1,
					MetalicFactor:   0,
				},
			},
		},
	}
}
