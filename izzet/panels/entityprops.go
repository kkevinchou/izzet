package panels

import (
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/types"
)

type ComponentComboOption string

var MaterialComboOption ComponentComboOption = "Material Component"
var PhysicsComboOption ComponentComboOption = "Physics Component"
var LightComboOption ComponentComboOption = "Light Component"
var SelectedComponentComboOption ComponentComboOption = MaterialComboOption

var componentComboOptions []ComponentComboOption = []ComponentComboOption{
	MaterialComboOption,
	PhysicsComboOption,
	LightComboOption,
}

func entityProps(entity *entities.Entity) {
	if imgui.CollapsingHeaderV("Entity Properties", imgui.TreeNodeFlagsDefaultOpen) {
		entityIDStr := ""
		entityNameStr := ""
		localRotationStr := ""
		localQuaternionStr := ""
		scaleStr := ""
		worldPositionStr := ""
		eulerRotationStr := ""
		parentStr := ""

		if entity != nil {
			entityIDStr = fmt.Sprintf("%d", entity.ID)
			entityNameStr = entity.NameID()

			rotation := entities.GetLocalRotation(entity)
			euler := QuatToEuler(rotation)
			localRotationStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z())
			localQuaternionStr = fmt.Sprintf("{%.2f, %.2f, %.2f, %.2f}", rotation.X(), rotation.Y(), rotation.Z(), rotation.W)

			scale := entities.GetLocalScale(entity)
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
		}

		imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
		uiTableRow("ID", entityIDStr)
		uiTableRow("Name", entityNameStr)

		var position *mgl64.Vec3
		var x, y, z int32
		if entity != nil {
			position = &entity.LocalPosition
			x, y, z = int32(position.X()), int32(position.Y()), int32(position.Z())
		}

		setupRow("Local Position X", func() {
			imgui.PushID("position x")
			if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
				if entity != nil {
					position[0] = float64(x)
					entities.SetDirty(entity)
				}
			}
			imgui.PopID()
		})

		setupRow("Local Position Y", func() {
			imgui.PushID("position y")
			if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
				if entity != nil {
					position[1] = float64(y)
					entities.SetDirty(entity)
				}
			}
			imgui.PopID()
		})

		setupRow("Local Position Z", func() {
			imgui.PushID("position z")
			if imgui.InputIntV("", &z, 0, 0, imgui.InputTextFlagsNone) {
				if entity != nil {
					position[2] = float64(z)
					entities.SetDirty(entity)
				}
			}
			imgui.PopID()
		})

		uiTableRow("Local Rotation", localRotationStr)
		uiTableRow("Local Quat", localQuaternionStr)
		uiTableRow("Scale", scaleStr)
		uiTableRow("World Position", worldPositionStr)
		uiTableRow("World Rotation", eulerRotationStr)
		uiTableRow("Parent", parentStr)
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
			setupRow("Color Intensity", func() {
				imgui.SliderFloatV("", &entity.LightInfo.PreScaledIntensity, 1, 20, "%.1f", imgui.SliderFlagsNone)
			})

			if entity.LightInfo.Type == entities.LightTypePoint {
				setupRow("Light Range", func() { imgui.SliderFloatV("", &entity.LightInfo.Range, 1, 1500, "%.0f", imgui.SliderFlagsNone) })
			} else if entity.LightInfo.Type == entities.LightTypeDirection {
				setupRow("Directional Light Direction", func() { imgui.SliderFloat3("", &entity.LightInfo.Direction3F, -1, 1) })
			}
			imgui.EndTable()
			imgui.PushID("remove light")
			if imgui.Button("Remove") {
				entity.LightInfo = nil
			}
			imgui.PopID()
		}
	}

	if entity.Material != nil {
		if imgui.CollapsingHeaderV("Material Properties", imgui.TreeNodeFlagsDefaultOpen) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)

			setupRow("Diffuse", func() {
				imgui.ColorEdit3V("", &entity.Material.PBR.Diffuse, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			})
			setupRow("Invisible", func() {
				imgui.Checkbox("", &entity.Material.Invisible)
			})

			setupRow("Diffuse Intensity", func() {
				imgui.SliderFloatV("", &entity.Material.PBR.DiffuseIntensity, 1, 20, "%.1f", imgui.SliderFlagsNone)
			})

			setupRow("Roughness", func() { imgui.SliderFloatV("", &entity.Material.PBR.Roughness, 0, 1, "%.2f", imgui.SliderFlagsNone) })
			setupRow("Metallic Factor", func() { imgui.SliderFloatV("", &entity.Material.PBR.Metallic, 0, 1, "%.2f", imgui.SliderFlagsNone) })
			imgui.EndTable()
			imgui.PushID("remove material")
			if imgui.Button("Remove") {
				entity.Material = nil
			}
			imgui.PopID()
		}
	}

	if entity.Physics != nil {
		physicsComponent := entity.Physics
		velocity := &physicsComponent.Velocity
		if imgui.CollapsingHeaderV("Physics Properties", imgui.TreeNodeFlagsDefaultOpen) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)

			var x, y, z int32 = int32(velocity.X()), int32(velocity.Y()), int32(velocity.X())

			setupRow("Velocity X", func() {
				imgui.PushID("velocity x")
				if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
					velocity[0] = float64(x)
				}
				imgui.PopID()
			})
			setupRow("Velocity Y", func() {
				imgui.PushID("velocity y")
				if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
					velocity[1] = float64(y)
				}
				imgui.PopID()
			})
			setupRow("Velocity Z", func() {
				imgui.PushID("velocity z")
				if imgui.InputIntV("", &z, 0, 0, imgui.InputTextFlagsNone) {
					velocity[2] = float64(z)
				}
				imgui.PopID()
			})
			imgui.EndTable()
			imgui.PushID("remove phys")
			if imgui.Button("Remove") {
				entity.Physics = nil
			}
			imgui.PopID()
		}
	}

	if entity.Collider != nil {
		if imgui.CollapsingHeaderV("Collider Properties", imgui.TreeNodeFlagsDefaultOpen) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)

			setupRow("Collider Type", func() {
				imgui.LabelText("", string(entities.ColliderFlagToGroupName[entity.Collider.ColliderGroup]))
			})
			imgui.EndTable()
		}
	}

	imgui.PushID("Component Combo")
	if imgui.BeginCombo("", string(SelectedComponentComboOption)) {
		for _, option := range componentComboOptions {
			if imgui.Selectable(string(option)) {
				SelectedComponentComboOption = option
			}
		}
		imgui.EndCombo()
	}
	imgui.PopID()
	if imgui.Button("Add Component") {
		entity := SelectedEntity()
		if entity != nil {
			if SelectedComponentComboOption == MaterialComboOption {
				entity.Material = &entities.MaterialComponent{
					PBR: types.PBR{
						Roughness:        0.85,
						Metallic:         0,
						Diffuse:          [3]float32{1, 1, 1},
						DiffuseIntensity: 1,
					},
				}
			} else if SelectedComponentComboOption == LightComboOption {
				entity.LightInfo = &entities.LightInfo{
					PreScaledIntensity: 3,
					Diffuse3F:          [3]float32{1, 1, 1},
					Type:               entities.LightTypePoint,
					Range:              800,
				}
			} else if SelectedComponentComboOption == PhysicsComboOption {
				entity.Physics = &entities.PhysicsComponent{}
			}
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
