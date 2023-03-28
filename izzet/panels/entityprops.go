package panels

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func entityProps(entity *entities.Entity) {
	// parentWindowSize := imgui.WindowSize()
	// windowSize := imgui.Vec2{X: parentWindowSize.X, Y: parentWindowSize.Y * 0.5}
	// imgui.BeginChildV("entityProps", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)
	if imgui.CollapsingHeaderV("Entity Properties", imgui.TreeNodeFlagsDefaultOpen) {
		// imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .95, Y: .91, Z: 0.81, W: 1})
		// imgui.PopStyleColor()

		if entity != nil {
			position := entities.LocalPosition(entity)
			positionStr := fmt.Sprintf("{%.1f, %.1f, %.1f}", position.X(), position.Y(), position.Z())
			text := &positionStr

			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			uiTableRow("Entity Name", entity.NameID())
			if uiTableInputRow("Local Position", text, nil) {
				uiTableInputPosition(entity, text)
			}

			rotation := entities.LocalRotation(entity)
			euler := QuatToEuler(rotation)
			uiTableRow("Local Rotation", fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z()))
			uiTableRow("Local Quat", fmt.Sprintf("{%.2f, %.2f, %.2f, %.2f}", rotation.X(), rotation.Y(), rotation.Z(), rotation.W))

			scale := entities.Scale(entity)
			uiTableRow("Scale", fmt.Sprintf("{%.0f, %.0f, %.0f}", scale.X(), scale.Y(), scale.Z()))

			position = entity.WorldPosition()
			positionStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", position.X(), position.Y(), position.Z())
			uiTableRow("World Position", positionStr)

			euler = QuatToEuler(entity.WorldRotation())
			uiTableRow("World Rotation", fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z()))

			parentStr := "nil"
			if entity.Parent != nil {
				parentStr = fmt.Sprintf("%s", entity.Parent.Name)
			}
			uiTableRow("Parent", parentStr)

			parentJointStr := "nil"
			if entity.ParentJoint != nil {
				parentJointStr = entity.ParentJoint.Name
			}
			uiTableRow("Parent Joint", parentJointStr)

			imgui.EndTable()
		}
	}
	// imgui.EndChild()

}

func uiTableInputPosition(entity *entities.Entity, text *string) {
	textCopy := *text
	r := regexp.MustCompile(`\{(?P<x>-?\d+), (?P<y>-?\d+), (?P<z>-?\d+)\}`)
	matches := r.FindStringSubmatch(textCopy)
	if matches != nil {
		var parseErr bool
		var newPosition mgl64.Vec3
		for i, name := range r.SubexpNames() {
			// https://pkg.go.dev/regexp#Regexp.SubexpNames
			// first name is always the empty string since the regexp as a whole cannot be named
			if i == 0 {
				continue
			}

			if i < 1 || i > 3 {
				parseErr = true
				continue
			}

			value, err := strconv.Atoi(matches[r.SubexpIndex(name)])
			if err != nil {
				parseErr = true
				continue
			}

			newPosition[i-1] = float64(value)
		}

		if !parseErr {
			entities.SetLocalPosition(entity, newPosition)
		}
	}
}

func uiTableInputRow(label string, text *string, cb imgui.InputTextCallback) bool {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(label)
	imgui.TableSetColumnIndex(1)

	imgui.PushItemWidth(imgui.WindowWidth())
	v := imgui.InputTextV("", text, imgui.ImGuiInputTextFlagsCallbackEdit|imgui.InputTextFlagsEnterReturnsTrue, cb)
	imgui.PopItemWidth()
	return v
}

func uiTableRow(label string, value any) {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(label)
	imgui.TableSetColumnIndex(1)
	imgui.Text(fmt.Sprintf("%v", value))
}

func QuatToEuler(q mgl64.Quat) mgl64.Vec3 {
	// Convert a quaternion into euler angles (roll, pitch, yaw)
	// roll is rotation around x in radians (counterclockwise)
	// pitch is rotation around y in radians (counterclockwise)
	// yaw is rotation around z in radians (counterclockwise)
	x := q.X()
	y := q.Y()
	z := q.Z()
	w := q.W

	t0 := 2.0 * (w*x + y*z)
	t1 := 1.0 - 2.0*(x*x+y*y)
	roll_x := math.Atan2(t0, t1)

	t2 := 2.0 * (w*y - z*x)
	if t2 > 1 {
		t2 = 1
	}

	if t2 < -1 {
		t2 = -1
	}
	pitch_y := math.Asin(t2)

	t3 := +2.0 * (w*z + x*y)
	t4 := +1.0 - 2.0*(y*y+z*z)
	yaw_z := math.Atan2(t3, t4)

	return mgl64.Vec3{mgl64.RadToDeg(roll_x), mgl64.RadToDeg(pitch_y), mgl64.RadToDeg(yaw_z)} // in radians
}
