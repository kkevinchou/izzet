package panels

import (
	"fmt"
	"sort"
	"time"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/internal/modelspec"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
)

var val int32
var inputText string

// var animation string
var currentItem int32
var LoopAnimation bool

var RenderJoints bool
var JointHover *int

var JointsToRender []int

func animationUI(app renderiface.App, entity *entities.Entity) {
	// imgui.SetNextWindowPosV(imgui.Vec2{X: 400, Y: 400}, imgui.ConditionFirstUseEver, imgui.Vec2{})
	// imgui.SetNextWindowSizeV(imgui.Vec2{X: 100, Y: 100}, imgui.ConditionFirstUseEver)

	// app.AssetManager().GetAnimations(entity.Animation.AnimationHandle)
	fullAnimationLength := entity.Animation.AnimationPlayer.Length()

	// imgui.BeginV("animation window", &open, imgui.WindowFlagsNone)

	var anims []string

	animations, joints, _ := app.AssetManager().GetAnimations(entity.Animation.AnimationHandle)
	for name, _ := range animations {
		anims = append(anims, name)
	}
	sort.Strings(anims)

	imgui.Text(entity.NameID())
	if imgui.ListBoxStrarr("animations", &currentItem, anims, int32(len(anims))) {
		entity.Animation.AnimationPlayer.PlayAnimation(anims[currentItem])
		entity.Animation.AnimationPlayer.UpdateTo(0)
	}
	imgui.Checkbox("loop", &LoopAnimation)
	imgui.SameLine()
	imgui.Checkbox("render all joints", &RenderJoints)

	if imgui.SliderInt("cool slider", &val, 0, int32(fullAnimationLength.Milliseconds())) {
		entity.Animation.AnimationPlayer.UpdateTo(time.Duration(val) * time.Millisecond)
		LoopAnimation = false
	}

	if imgui.Button("Add Annotation") {
		inputText = ""
	}

	imgui.LabelText("", "Joints")
	JointHover = nil
	JointsToRender = nil
	drawJointTree(app, entity, joints[entity.Animation.RootJointID])

	if RenderJoints {
		// for jid, _ := range entity.Model.JointMap() {
		// 	JointsToRender = append(JointsToRender, jid)
		// }
	} else if JointHover != nil {
		JointsToRender = append(JointsToRender, *JointHover)
	}

	// imgui.End()
}

func drawJointTree(app renderiface.App, parent *entities.Entity, joint *modelspec.JointSpec) {
	var nodeFlags imgui.TreeNodeFlags = imgui.TreeNodeFlagsNone

	if len(joint.Children) == 0 {
		nodeFlags = nodeFlags | imgui.TreeNodeFlagsLeaf
	}

	opened := imgui.TreeNodeExStrV(fmt.Sprintf("[%d] %s", joint.ID, joint.Name), nodeFlags)

	imgui.PushIDStr(joint.Name)
	setupMenu(app, parent, joint)
	imgui.PopID()
	if imgui.IsItemHovered() {
		JointHover = &joint.ID
	}

	if opened {
		for _, child := range joint.Children {
			drawJointTree(app, parent, child)
		}
		imgui.TreePop()
	}
}

func setupMenu(app renderiface.App, parent *entities.Entity, joint *modelspec.JointSpec) {
	imgui.PushStyleColorVec4(imgui.ColButton, imgui.Vec4{X: 66. / 255, Y: 17. / 255, Z: 212. / 255, W: 1})
	imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1})
	if imgui.BeginPopupContextItemV("NULL", imgui.PopupFlagsMouseButtonRight) {
		if imgui.Button("Create Socket") {
			// socket := entities.CreateSocket()
			// app.AddEntity(socket)
			// entities.BuildRelation(SelectedEntity(), socket)
			// socket.ParentJoint = joint
			imgui.CloseCurrentPopup()
		}

		if imgui.BeginMenu("Assign Socket") {
			// socketCount := 0
			// for _, entity := range app.Entities() {
			// 	if entity.IsSocket {
			// 		socketCount++

			// 		isParented := entity.ParentJoint != nil && entity.ParentJoint.ID == joint.ID
			// 		if imgui.MenuItemBoolV(entity.NameID(), "", isParented, true) {
			// 			// toggle parented status
			// 			if isParented {
			// 				entities.RemoveParent(entity)
			// 				entity.ParentJoint = nil
			// 			} else {
			// 				entities.BuildRelation(parent, entity)
			// 				entity.ParentJoint = joint
			// 			}
			// 		}
			// 	}
			// }

			// if socketCount == 0 {
			// 	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: 0.5, Y: 0.5, Z: 0.5, W: 0.5})
			// 	imgui.MenuItem("none")
			// 	imgui.PopStyleColor()
			// }
			imgui.EndMenu()
		}
		imgui.EndPopup()
	}

	imgui.PopStyleColor()
	imgui.PopStyleColor()
}
