package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/prefabs"
)

var open bool

func BuildTabsSet(world World, renderContext RenderContext, menuBarSize imgui.Vec2, ps []*prefabs.Prefab, es []*entities.Entity) {
	rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}
	width := rect.X * 0.20
	height := rect.Y * 0.5

	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: width, Y: height}, imgui.ConditionFirstUseEver)
	imgui.BeginV("Fixed Tab Set", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoMove|imgui.WindowFlagsNoCollapse)

	if imgui.BeginTabBar("Scene") {
		if imgui.BeginTabItem("Scene Hierarchy") {
			sceneUI(es, world)
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()

	imgui.SetNextWindowBgAlpha(0.8)
	imgui.SetNextWindowPosV(imgui.Vec2{Y: menuBarSize.Y + rect.Y*0.5}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: width, Y: height}, imgui.ConditionFirstUseEver)
	imgui.BeginV("Free Tab Set", &open, imgui.WindowFlagsNoTitleBar|imgui.WindowFlagsNoCollapse)

	if imgui.BeginTabBar("Main") {
		if imgui.BeginTabItem("World Properties") {
			worldProps(renderContext)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Entity Properties") {
			entityProps(SelectedEntity())
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Prefabs") {
			prefabsUI(ps)
			imgui.EndTabItem()
		}
		if imgui.BeginTabItem("Animation") {
			entity := SelectedEntity()
			if entity != nil && entity.AnimationPlayer != nil {
				animationUI(world, entity)
			} else {
				imgui.Text("<select an entity with animations>")
			}
			imgui.EndTabItem()
		}
		imgui.EndTabBar()
	}

	imgui.End()
}
