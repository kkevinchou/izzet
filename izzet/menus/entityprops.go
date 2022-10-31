package menus

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

func entityProps(entity *entities.Entity) {
	parentWindowSize := imgui.WindowSize()
	// imgui.SetNextWindowPos(imgui.Vec2{})
	imgui.SetNextWindowBgAlpha(0.8)
	windowSize := imgui.Vec2{X: parentWindowSize.X, Y: parentWindowSize.Y * 0.5}
	imgui.BeginChildV("entityProps", windowSize, false, imgui.WindowFlagsNoMove|imgui.WindowFlagsNoResize)

	if entity != nil {
		positionStr := fmt.Sprintf("%v", entity.Position)
		text := &positionStr

		imgui.BeginTableV("", 2, imgui.TableFlagsBorders, imgui.Vec2{}, 0)
		uiTableRow("Entity Name", entity.Name)
		if uiTableInputRow("Position", entity.Position, text, nil) {
			textCopy := *text
			r := regexp.MustCompile(`\[(?P<x>\d+) (?P<y>\d+) (?P<z>\d+)\]`)
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
