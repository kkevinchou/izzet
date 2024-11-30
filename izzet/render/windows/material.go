package windows

import (
	"fmt"
	"os"
	"path/filepath"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/sqweek/dialog"
)

var materialIDGen int

var (
	activeMaterial *modelspec.MaterialSpecification
	tableFlags     imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

var showCreateMaterialModel bool

func ShowCreateMaterialWindow(material *modelspec.MaterialSpecification) {
	showCreateMaterialModel = true
	if material != nil {
		activeMaterial = material
	} else {
		assignDefaultMaterial()
	}
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

	if imgui.BeginV("Create/Update Material", &showCreateMaterialModel, imgui.WindowFlagsNone) {
		imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("ID", func() {
			imgui.InputTextWithHint("##MaterialID", "", &activeMaterial.ID, imgui.InputTextFlagsNone, nil)
		}, true)

		panelutils.SetupRow("Diffuse", func() {
			// color := [3]float32{}
			var color [3]float32 = activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorFactor.Vec3()
			if imgui.ColorEdit3V("", &color, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel) {
				activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[0] = color[0]
				activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[1] = color[1]
				activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[2] = color[2]
				activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorFactor[3] = 1
			}
		}, true)

		panelutils.SetupRow("Roughness", func() {
			imgui.SliderFloatV("", &activeMaterial.PBRMaterial.PBRMetallicRoughness.RoughnessFactor, 0, 1, "%.2f", imgui.SliderFlagsNone)
		}, true)
		panelutils.SetupRow("Metallic Factor", func() {
			imgui.SliderFloatV("", &activeMaterial.PBRMaterial.PBRMetallicRoughness.MetalicFactor, 0, 1, "%.2f", imgui.SliderFlagsNone)
		}, true)
		panelutils.SetupRow("Texture", func() { imgui.LabelText("", activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName) }, true)

		if activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName != "" {
			t := app.AssetManager().GetTexture(activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName)
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
				activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorTextureName = baseFileName
				activeMaterial.PBRMaterial.PBRMetallicRoughness.RoughnessFactor = 0.55
				activeMaterial.PBRMaterial.PBRMetallicRoughness.MetalicFactor = 0
				activeMaterial.PBRMaterial.PBRMetallicRoughness.BaseColorFactor = mgl32.Vec4{1, 1, 1, 1}
			}
		}

		imgui.EndTable()
		if imgui.Button("Save As New Material") {
			// materialIDGen++
			// activeMaterial.ID = fmt.Sprintf("material-%d", materialIDGen)
			showCreateMaterialModel = false
			assignDefaultMaterial()
		}
		imgui.SameLine()
		if imgui.Button("Cancel") {
			showCreateMaterialModel = false
			assignDefaultMaterial()
		}
	}
	imgui.End()
}

func assignDefaultMaterial() {
	activeMaterial = &modelspec.MaterialSpecification{
		ID: fmt.Sprintf("material-%d", materialIDGen),
		PBRMaterial: modelspec.PBRMaterial{
			PBRMetallicRoughness: modelspec.PBRMetallicRoughness{
				BaseColorFactor: mgl32.Vec4{1, 1, 1, 1},
				RoughnessFactor: 0.55,
				MetalicFactor:   1,
			},
		},
	}
}
