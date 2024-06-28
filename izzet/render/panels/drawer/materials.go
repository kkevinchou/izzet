package drawer

import (
	"fmt"
	"os"
	"path/filepath"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/material"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/sqweek/dialog"
)

var (
	tableFlags  imgui.TableFlags = imgui.TableFlagsBordersInnerV
	WIPMaterial material.Material
)

func materialssUI(app renderiface.App, ps []*prefabs.Prefab) bool {
	mat := &WIPMaterial
	var menuOpen bool
	if imgui.BeginTabItem("Materials") {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Materials", imgui.TreeNodeFlagsDefaultOpen) {
			imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
			panelutils.InitColumns()
			// imgui.Tablepanelutils.SetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)

			// Frame Profiling
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
				}
			}

			imgui.EndTable()
			if imgui.Button("Create Material") {
				app.CreateMaterial(*mat)
				WIPMaterial = material.Material{}
			}
		}
		imgui.EndTabItem()
		if imgui.CollapsingHeaderTreeNodeFlagsV("Materials List", imgui.TreeNodeFlagsNone) {
			for i, _ := range app.MaterialBrowser().Items {
				var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone //| imgui.TreeNodeFlagsLeaf
				open := imgui.TreeNodeExStrV(fmt.Sprintf("material-%d", i), nodeFlags)
				if open {
					imgui.TreePop()
				}
			}
		}
	}
	return menuOpen
}
