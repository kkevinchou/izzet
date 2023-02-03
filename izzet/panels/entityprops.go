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
	parentWindowSize := imgui.WindowSize()
	windowSize := imgui.Vec2{X: parentWindowSize.X, Y: parentWindowSize.Y * 0.5}
	imgui.BeginChildV("entityProps", windowSize, true, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	imgui.PushStyleColor(imgui.StyleColorText, imgui.Vec4{X: .95, Y: .91, Z: 0.81, W: 1})
	imgui.Text("Entity Properties")
	imgui.PopStyleColor()

	if entity != nil {
		positionStr := fmt.Sprintf("%v", entity.Position)
		text := &positionStr

		imgui.BeginTableV("", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0)
		uiTableRow("Entity Name", entity.Name)
		if uiTableInputRow("Position", entity.Position, text, nil) {
			textCopy := *text
			r := regexp.MustCompile(`\[(?P<x>-?\d+) (?P<y>-?\d+) (?P<z>-?\d+)\]`)
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
					entity.Position = newPosition
				}
			}
		}

		euler := QuatToEuler(entity.Rotation)
		uiTableRow("Rotation", fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z()))
		imgui.EndTable()
	}

	imgui.EndChild()
}

func uiTableInputRow(label string, value any, text *string, cb imgui.InputTextCallback) bool {
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
