package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type CreateEntityComboOption string

const (
	CreateEntityComboOptionVelociraptor    CreateEntityComboOption = CreateEntityComboOption(entities.EntityTypeVelociraptor)
	CreateEntityComboOptionParasaurolophus CreateEntityComboOption = CreateEntityComboOption(entities.EntityTypeParasaurolophus)
)

var SelectedCreateEntityComboOption CreateEntityComboOption = CreateEntityComboOptionVelociraptor

var (
	createEntityComboOptions []CreateEntityComboOption = []CreateEntityComboOption{
		CreateEntityComboOptionVelociraptor,
		CreateEntityComboOptionParasaurolophus,
	}
)

func controls(app renderiface.App, renderContext RenderContext) {
	if imgui.CollapsingHeaderTreeNodeFlagsV("Entity Type", imgui.TreeNodeFlagsDefaultOpen) {
		if imgui.BeginCombo("##", string(SelectedCreateEntityComboOption)) {
			for _, option := range createEntityComboOptions {
				if imgui.SelectableBool(string(option)) {
					SelectedCreateEntityComboOption = option
				}
			}
			imgui.EndCombo()
		}
	}
}
