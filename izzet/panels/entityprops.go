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
	if imgui.CollapsingHeaderV("Entity Properties", imgui.TreeNodeFlagsDefaultOpen) {
		entityNameStr := ""
		positionStr := ""
		localRotationStr := ""
		localQuaternionStr := ""
		scaleStr := ""
		worldPositionStr := ""
		eulerRotationStr := ""
		parentStr := ""
		parentJointStr := ""

		if entity != nil {
			entityNameStr = entity.NameID()
			position := entities.LocalPosition(entity)
			positionStr = fmt.Sprintf("{%.1f, %.1f, %.1f}", position.X(), position.Y(), position.Z())

			rotation := entities.LocalRotation(entity)
			euler := QuatToEuler(rotation)
			localRotationStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z())
			localQuaternionStr = fmt.Sprintf("{%.2f, %.2f, %.2f, %.2f}", rotation.X(), rotation.Y(), rotation.Z(), rotation.W)

			scale := entities.Scale(entity)
			scaleStr = fmt.Sprintf("{%.2f, %.2f, %.2f}", scale.X(), scale.Y(), scale.Z())

			worldPosition := entity.WorldPosition()
			worldPositionStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", worldPosition.X(), worldPosition.Y(), worldPosition.Z())

			euler = QuatToEuler(entity.WorldRotation())
			eulerRotationStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z())

			if entity.Parent != nil {
				parentStr = fmt.Sprintf("%s", entity.Parent.Name)
			} else {
				parentStr = "nil"
			}

			if entity.ParentJoint != nil {
				parentJointStr = entity.ParentJoint.Name
			} else {
				parentJointStr = "nil"
			}
		}

		imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
		uiTableRow("Entity Name", entityNameStr)

		// if entity != nil {
		// 	// if uiTableInputRow("Local Position", &positionStr, nil) {
		// 	// 	uiTableInputPosition(entity, &positionStr)
		// 	// }
		// }
		uiTableRow("Local Position", positionStr)

		uiTableRow("Local Rotation", localRotationStr)
		uiTableRow("Local Quat", localQuaternionStr)
		uiTableRow("Scale", scaleStr)
		uiTableRow("World Position", worldPositionStr)
		uiTableRow("World Rotation", eulerRotationStr)
		uiTableRow("Parent", parentStr)
		uiTableRow("Parent Joint", parentJointStr)
		imgui.EndTable()
	}

	if entity == nil {
		return
	}

	if entity.LightInfo != nil {
		if imgui.CollapsingHeaderV("Light Properties", imgui.TreeNodeFlagsDefaultOpen) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)

			lightTypeStr := "?"
			if entity.LightInfo.Type == entities.LightTypePoint {
				lightTypeStr = "Point Light"
			} else if entity.LightInfo.Type == entities.LightTypeDirection {
				lightTypeStr = "Directional Light"
			}
			uiTableRow("Light Type", lightTypeStr)
			setupRow("Color", func() {
				imgui.ColorEdit3V("", &entity.LightInfo.Diffuse3F, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			})
			setupRow("Color Intensity", func() { imgui.SliderFloat("", &entity.LightInfo.PreScaledIntensity, 1, 20) })
			imgui.EndTable()
		}
	}
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

	// imgui.PushItemWidth(imgui.WindowWidth())
	v := imgui.InputTextV("", text, imgui.ImGuiInputTextFlagsCallbackEdit|imgui.InputTextFlagsEnterReturnsTrue, cb)
	// v := imgui.InputTextV("", text, imgui.InputTextFlagsNone, cb)
	// imgui.PopItemWidth()
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
