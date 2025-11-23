package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

type CreateEntityComboOption string

const (
	CreateEntityComboOptionVelociraptor    CreateEntityComboOption = CreateEntityComboOption(entity.EntityTypeVelociraptor)
	CreateEntityComboOptionParasaurolophus CreateEntityComboOption = CreateEntityComboOption(entity.EntityTypeParasaurolophus)
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
