package menus

import (
	"fmt"
	"os"
	"path/filepath"

	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
)

var errorModal error

func file(app renderiface.App) {
	if imgui.BeginMenu("File") {
		imgui.InputTextWithHint("##WorldName", "", &worldName, imgui.InputTextFlagsNone, nil)

		center := imgui.MainViewport().Center()
		imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})

		imgui.SameLine()
		if imgui.Button("Save") {
			fmt.Println("Save to", worldName)
			if err := app.SaveProject(worldName); err != nil {
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
			if _, ok := ignoredJsonFiles[file.Name()]; ok {
				continue
			}

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

		imgui.EndMenu()
	}
}
