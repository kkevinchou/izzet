package panels

import (
	"fmt"
	"time"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

var val int32
var inputText string

func BuildAnimation(world World, entity *entities.Entity) {
	// var heightRatio float32 = 0.15
	// _ = heightRatio
	// imgui.SetNextWindowBgAlpha(0.8)

	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	fullAnimationLength := entity.AnimationPlayer.Length()

	imgui.BeginV("animation window", &open, imgui.WindowFlagsNone)
	imgui.SliderInt("cool slider", &val, 0, int32(fullAnimationLength.Milliseconds()))
	entity.AnimationPlayer.UpdateTo(time.Duration(val) * time.Millisecond)
	imgui.InputText("some input text", &inputText)
	if imgui.Button("Add Annotation") {
		fmt.Println("clicked")
		inputText = ""
	}
	// imgui.BeginChildV("prefab", imgui.Vec2{}, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	// imgui.EndChild()
	imgui.End()

}
