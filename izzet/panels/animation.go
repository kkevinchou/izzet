package panels

import (
	"fmt"
	"sort"
	"time"

	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/modelspec"
)

var val int32
var inputText string

// var animation string
var currentItem int32
var LoopAnimation bool
var RenderJoints bool

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
	imgui.Checkbox("loop", &LoopAnimation)
	imgui.SameLine()
	imgui.Checkbox("render joints", &RenderJoints)

	if imgui.SliderInt("cool slider", &val, 0, int32(fullAnimationLength.Milliseconds())) {
		entity.AnimationPlayer.UpdateTo(time.Duration(val) * time.Millisecond)
		LoopAnimation = false
	}

	imgui.InputText("some input text", &inputText)
	if imgui.Button("Add Annotation") {
		inputText = ""
	}

	imgui.LabelText("", "Joints")
	drawJoint(entity.Model.RootJoint())

	imgui.End()
}

func drawJoint(joint *modelspec.JointSpec) {
	nodeFlags := imgui.TreeNodeFlagsNone

	if len(joint.Children) == 0 {
		nodeFlags = nodeFlags | imgui.TreeNodeFlagsLeaf
	}
	if imgui.TreeNodeV(fmt.Sprintf("[%d] %s", joint.ID, joint.Name), nodeFlags) {
		for _, child := range joint.Children {
			drawJoint(child)
		}
		imgui.TreePop()
	}

}
