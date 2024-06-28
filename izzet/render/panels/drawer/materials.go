package drawer

import (
	"fmt"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/material"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/types"
)

var (
	tableFlags imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

func materialssUI(app renderiface.App, ps []*prefabs.Prefab) bool {
	var menuOpen bool
	if imgui.BeginTabItem("Materials") {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Materials", imgui.TreeNodeFlagsDefaultOpen) {
			imgui.BeginTableV("Material Editor", 2, tableFlags, imgui.Vec2{}, 0)
			panelutils.InitColumns()
			// imgui.Tablepanelutils.SetupColumnV("0", imgui.TableColumnFlagsWidthFixed, tableColumn0Width, 0)

			material := material.Material{
				PBR: types.PBR{},
			}

			// Frame Profiling
			panelutils.SetupRow("Diffuse", func() {
				imgui.ColorEdit3V("", &material.PBR.Diffuse, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			}, true)
			panelutils.SetupRow("Invisible", func() {
				imgui.Checkbox("", &material.Invisible)
			}, true)

			panelutils.SetupRow("Diffuse Intensity", func() {
				imgui.SliderFloatV("", &material.PBR.DiffuseIntensity, 1, 20, "%.1f", imgui.SliderFlagsNone)
			}, true)

			panelutils.SetupRow("Roughness", func() { imgui.SliderFloatV("", &material.PBR.Roughness, 0, 1, "%.2f", imgui.SliderFlagsNone) }, true)
			panelutils.SetupRow("Metallic Factor", func() { imgui.SliderFloatV("", &material.PBR.Metallic, 0, 1, "%.2f", imgui.SliderFlagsNone) }, true)

			imgui.EndTable()
			if imgui.Button("Create Material") {
				app.CreateMaterial(material)
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
