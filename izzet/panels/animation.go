package panels

import (
	"fmt"
	"sort"
	"time"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

var val int32
var inputText string

// var animation string
var currentItem int32

func BuildAnimation(world World, entity *entities.Entity) {
	imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	fullAnimationLength := entity.AnimationPlayer.Length()

	imgui.BeginV("animation window", &open, imgui.WindowFlagsNone)

	var anims []string
	for name, _ := range entity.Animations {
		anims = append(anims, name)
	}
	sort.Strings(anims)

	if imgui.ListBox("animations", &currentItem, anims) {
		entity.AnimationPlayer.PlayAnimation(anims[currentItem])
		entity.AnimationPlayer.UpdateTo(0)

	}
	imgui.SliderInt("cool slider", &val, 0, int32(fullAnimationLength.Milliseconds()))
	entity.AnimationPlayer.UpdateTo(time.Duration(val) * time.Millisecond)
	imgui.InputText("some input text", &inputText)
	if imgui.Button("Add Annotation") {
		fmt.Println("clicked")
		inputText = ""
	}

	imgui.End()

}
