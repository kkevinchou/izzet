package menus

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/settings"
)

func file(app renderiface.App) {
	if imgui.BeginMenu("File") {
		imgui.PushID("World Name")
		imgui.InputText("", &worldName)
		imgui.PopID()

		imgui.SameLine()
		if imgui.Button("Save") {
			fmt.Println("Save to", worldName)
			app.SaveProjectAs(worldName)
		}

		files, err := os.ReadDir(settings.ProjectsDirectory)
		if err != nil {
			panic(err)
		}

		var savedWorlds []string
		for _, file := range files {
			extension := filepath.Ext(file.Name())
			// if extension != ".json" {
			// 	continue
			// }

			if _, ok := ignoredJsonFiles[file.Name()]; ok {
				continue
			}

			name := file.Name()[0 : len(file.Name())-len(extension)]
			savedWorlds = append(savedWorlds, name)
		}

		if len(savedWorlds) == 0 {
			savedWorlds = append(savedWorlds, selectedWorldName)
		}

		if imgui.BeginCombo("", selectedWorldName) {
			for _, worldName := range savedWorlds {
				if imgui.Selectable(worldName) {
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
		}

		imgui.EndMenu()
	}
}