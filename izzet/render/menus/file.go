package menus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/apputils"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
	"github.com/sqweek/dialog"
)

var errorModal error
var showImportAssetModal bool

var defaultColliderType types.ColliderType = types.ColliderTypeMesh
var defaultColliderGroup types.ColliderGroup = types.ColliderGroupTerrain

var SelectedColliderType types.ColliderType = defaultColliderType
var SelectedColliderGroup types.ColliderGroup = defaultColliderGroup

func file(app renderiface.App) {
	if imgui.BeginMenu("File") {
		imgui.InputTextWithHint("##WorldName", "", &worldName, imgui.InputTextFlagsNone, nil)

		center := imgui.MainViewport().Center()
		imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

		imgui.SameLine()
		if imgui.Button("Save") {
			fmt.Println("Save to", worldName)
			if err := app.SaveProjectAs(worldName); err != nil {
				errorModal = err
			} else {
				imgui.CloseCurrentPopup()
			}
		}

		if errorModal != nil {
			imgui.OpenPopupStr("Error")
			if imgui.BeginPopupModalV("Error", nil, imgui.WindowFlagsAlwaysAutoResize) {
				imgui.LabelText("##", errorModal.Error())
				if imgui.Button("OK") {
					errorModal = nil
					imgui.CloseCurrentPopup()
				}
				imgui.EndPopup()
			}
		}
		err := os.MkdirAll(filepath.Join(settings.ProjectsDirectory), os.ModePerm)
		if err != nil {
			panic(err)
		}
		files, err := os.ReadDir(settings.ProjectsDirectory)
		if err != nil {
			panic(err)
		}

		var savedWorlds []string
		for _, file := range files {
			extension := filepath.Ext(file.Name())

			name := file.Name()[0 : len(file.Name())-len(extension)]
			savedWorlds = append(savedWorlds, name)
		}

		if imgui.BeginCombo("##", selectedWorldName) {
			for _, worldName := range savedWorlds {
				if imgui.SelectableBool(worldName) {
					selectedWorldName = worldName
				}
			}
			imgui.EndCombo()
		}
		imgui.SameLine()
		if imgui.Button("Load") {
			fmt.Println("Load from", selectedWorldName)
			if app.LoadProject(selectedWorldName) {
				worldName = selectedWorldName
			}
			imgui.CloseCurrentPopup()
		}
		if imgui.MenuItemBool("New Project") {
			app.NewProject(settings.NewProjectName)
			selectedWorldName = ""
			worldName = settings.NewProjectName
		}

		if imgui.MenuItemBool("Import Asset") {
			wipImportAssetConfig = assets.AssetConfig{
				ColliderType:  string(defaultColliderType),
				ColliderGroup: string(defaultColliderGroup),
			}
			showImportAssetModal = true
		}

		imgui.EndMenu()
	}

	if showImportAssetModal {
		importAssetModal(app)
	}
}

var wipImportAssetConfig assets.AssetConfig

func importAssetModal(app renderiface.App) {
	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})
	// imgui.SetNextWindowSize(imgui.Vec2{X: 600, Y: 500})

	imgui.OpenPopupStr("Import Asset")
	if imgui.BeginPopupModalV("Import Asset", nil, imgui.WindowFlagsAlwaysAutoResize) {
		imgui.BeginTableV("Import Asset", 2, tableFlags, imgui.Vec2{}, 0)
		panelutils.InitColumns()

		panelutils.SetupRow("ID", func() {
			imgui.SetNextItemWidth(200)
			imgui.InputTextWithHint("##ID", "", &wipImportAssetConfig.Name, imgui.InputTextFlagsNone, nil)
		}, true)
		panelutils.SetupRow("File Path", func() {
			imgui.SetNextItemWidth(200)
			imgui.InputTextWithHint("##FilePath", "", &wipImportAssetConfig.FilePath, imgui.InputTextFlagsNone, nil)
			imgui.SameLine()
			if imgui.Button("...") {
				d := dialog.File()
				currentDir, err := os.Getwd()
				if err != nil {
					panic(err)
				}
				d = d.SetStartDir(filepath.Join(currentDir, settings.BuiltinAssetsDir, "gltf"))
				d = d.Filter("GLTF file", "gltf")

				assetFilePath, err := d.Load()
				if err != nil {
					if err != dialog.ErrCancelled {
						panic(err)
					}
				} else {
					wipImportAssetConfig.FilePath = assetFilePath
					wipImportAssetConfig.Name = apputils.NameFromAssetFilePath(assetFilePath)
				}
			}
		}, true)
		panelutils.SetupRow("Collider Type", func() {
			if imgui.BeginCombo("##", string(SelectedColliderType)) {
				for _, option := range types.ColliderTypes {
					if imgui.SelectableBool(string(option)) {
						SelectedColliderType = option
						wipImportAssetConfig.ColliderType = string(option)
					}
				}
				imgui.EndCombo()
			}
		}, true)
		panelutils.SetupRow("Collider Group", func() {
			if imgui.BeginCombo("##", string(SelectedColliderGroup)) {
				for _, option := range types.ColliderGroups {
					if imgui.SelectableBool(string(option)) {
						SelectedColliderGroup = option
						wipImportAssetConfig.ColliderGroup = string(option)
					}
				}
				imgui.EndCombo()
			}
		}, true)
		panelutils.SetupRow("Static", func() {
			imgui.Checkbox("##", &wipImportAssetConfig.Static)
		}, true)
		panelutils.SetupRow("Physics", func() {
			imgui.Checkbox("##", &wipImportAssetConfig.Physics)
		}, true)
		panelutils.SetupRow("Single Entity", func() {
			imgui.Checkbox("##", &wipImportAssetConfig.SingleEntity)
		}, true)

		imgui.EndTable()

		if imgui.Button("Import") {
			app.ImportAsset(wipImportAssetConfig)
			imgui.CloseCurrentPopup()
			showImportAssetModal = false
			SelectedColliderType = defaultColliderType
			SelectedColliderGroup = defaultColliderGroup
		}
		imgui.SameLine()
		if imgui.Button("Cancel") {
			imgui.CloseCurrentPopup()
			showImportAssetModal = false
		}

		imgui.EndPopup()
	}
}
