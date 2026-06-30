package panels

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/izzet/appmode"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/ui"
)

type ComponentComboOption string

var SelectedComponentComboOption ComponentComboOption = PhysicsComboOption

var PhysicsComboOption ComponentComboOption = "Physics Component"
var LightComboOption ComponentComboOption = "Light Component"
var ImageComboOption ComponentComboOption = "Image Component"
var SpawnPointComboOption ComponentComboOption = "Spawn Point Component"

var componentComboOptions []ComponentComboOption = []ComponentComboOption{
	PhysicsComboOption,
	LightComboOption,
	ImageComboOption,
	SpawnPointComboOption,
}

var (
	animationFilterText string
)

const animationComboListHeight float32 = 200

func EntityProps(e *entity.Entity, app renderiface.App) {
	if imgui.CollapsingHeaderTreeNodeFlagsV("Entity Properties", imgui.TreeNodeFlagsDefaultOpen) {
		entityIDStr := ""
		entityNameStr := ""
		localRotationStr := ""
		localQuaternionStr := ""
		scaleStr := ""
		worldPositionStr := ""
		eulerRotationStr := ""
		parentStr := ""

		if e != nil {
			entityIDStr = fmt.Sprintf("%d", e.ID)
			entityNameStr = e.NameID()

			rotation := e.GetLocalRotation()
			euler := QuatToEuler(rotation)
			localRotationStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z())
			localQuaternionStr = fmt.Sprintf("{%.2f, %.2f, %.2f, %.2f}", rotation.X(), rotation.Y(), rotation.Z(), rotation.W)

			scale := e.Scale()
			scaleStr = fmt.Sprintf("{%.2f, %.2f, %.2f}", scale.X(), scale.Y(), scale.Z())

			worldPosition := e.Position()
			worldPositionStr = fmt.Sprintf("{%.1f, %.1f, %.1f}", worldPosition.X(), worldPosition.Y(), worldPosition.Z())

			euler = QuatToEuler(e.Rotation())
			eulerRotationStr = fmt.Sprintf("{%.0f, %.0f, %.0f}", euler.X(), euler.Y(), euler.Z())

			if e.Parent != nil {
				parentStr = fmt.Sprintf("%s", e.Parent.Name)
			} else {
				parentStr = "nil"
			}

		}

		imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
		ui.InitColumns()
		uiTableRow("ID", entityIDStr)
		uiTableRow("Name", entityNameStr)

		var position *mgl64.Vec3
		var x, y, z float32
		if e != nil {
			position = &e.LocalPosition
			x, y, z = float32(position.X()), float32(position.Y()), float32(position.Z())
		}

		ui.RowV("Local Position", func() {
			if e != nil {
				imgui.PushItemWidth(imgui.ContentRegionAvail().X / 3.0)
				if imgui.InputFloatV("##x", &x, 0, 0, "%.2f", imgui.InputTextFlagsNone) {
					if e != nil {
						position[0] = float64(x)
						entity.SetDirty(e)
					}
				}
				imgui.SameLine()
				if imgui.InputFloatV("##y", &y, 0, 0, "%.2f", imgui.InputTextFlagsNone) {
					if e != nil {
						position[1] = float64(y)
						entity.SetDirty(e)
					}
				}
				imgui.SameLine()
				if imgui.InputFloatV("##z", &z, 0, 0, "%.2f", imgui.InputTextFlagsNone) {
					if e != nil {
						position[2] = float64(z)
						entity.SetDirty(e)
					}
				}
				imgui.PopItemWidth()
			}
		}, false)

		uiTableRow("Local Rotation", localRotationStr)
		uiTableRow("Local Quat", localQuaternionStr)
		uiTableRow("Scale", scaleStr)
		uiTableRow("World Position", worldPositionStr)
		uiTableRow("World Rotation", eulerRotationStr)
		uiTableRow("Parent", parentStr)
		if e != nil {
			uiTableRow("Static", fmt.Sprintf("%v", e.Static))
			uiTableRow("Deadge", fmt.Sprintf("%v", e.Deadge))
		}
		imgui.EndTable()
	}

	if e == nil {
		return
	}

	if e.LightInfo != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Light Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			lightTypeStr := "?"
			if e.LightInfo.Type == entity.LightTypePoint {
				lightTypeStr = "Point Light"
			} else if e.LightInfo.Type == entity.LightTypeDirection {
				lightTypeStr = "Directional Light"
			}
			uiTableRow("Light Type", lightTypeStr)
			ui.RowV("Color", func() {
				imgui.ColorEdit3V("", &e.LightInfo.Diffuse3F, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			}, true)
			ui.RowV("Color Intensity", func() {
				if e.LightInfo.Type == entity.LightTypePoint {
					imgui.SliderFloatV("", &e.LightInfo.PreScaledIntensity, 0, 0.1, "%.3f", imgui.SliderFlagsNone)
				} else if e.LightInfo.Type == entity.LightTypeDirection {
					imgui.SliderFloatV("", &e.LightInfo.PreScaledIntensity, 0, 6, "%.3f", imgui.SliderFlagsNone)
				}
			}, true)

			if e.LightInfo.Type == entity.LightTypePoint {
				ui.RowV("Light Range", func() { imgui.SliderFloatV("", &e.LightInfo.Range, 1, 1500, "%.0f", imgui.SliderFlagsNone) }, true)
			} else if e.LightInfo.Type == entity.LightTypeDirection {
				ui.RowV("Directional Light Direction", func() { imgui.SliderFloat3("", &e.LightInfo.Direction3F, -1, 1) }, true)
			}
			imgui.EndTable()
			imgui.PushIDStr("remove light")
			if imgui.Button("Remove") {
				e.LightInfo = nil
			}
			imgui.PopID()
		}
	}

	if e.ImageComponent != nil {
		imageComponent := e.ImageComponent
		if imgui.CollapsingHeaderTreeNodeFlagsV("Image Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			ui.RowV("Image Name", func() {
				imgui.InputTextWithHint("", "default.png", &imageComponent.ImageName, imgui.InputTextFlagsNone, nil)
			}, true)

			ui.RowV("Scale", func() {
				scale := float32(imageComponent.Scale)
				if imgui.InputFloatV("", &scale, 0.1, 1, "%.2f", imgui.InputTextFlagsNone) {
					imageComponent.Scale = float64(scale)
				}
			}, true)

			ui.RowV("Billboard", func() {
				imgui.Checkbox("", &imageComponent.Billboard)
			}, true)

			imgui.EndTable()
			imgui.PushIDStr("remove image")
			if imgui.Button("Remove") {
				e.ImageComponent = nil
			}
			imgui.PopID()
		}
	}

	originalMeshTriCount := 0

	if e.MeshComponent != nil {
		for _, primitive := range app.AssetManager().GetPrimitives(e.MeshComponent.MeshHandle) {
			originalMeshTriCount += len(primitive.Primitive.VertexIndices) / 3
		}

		if imgui.CollapsingHeaderTreeNodeFlagsV("Mesh Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()
			ui.RowV("Visible", func() { imgui.Checkbox("", &e.MeshComponent.Visible) }, true)
			ui.RowV("Shadow Casting", func() { imgui.Checkbox("", &e.MeshComponent.ShadowCasting) }, true)

			uiTableRow("Original Triangle Count", originalMeshTriCount)
			var materialStrs []string
			for _, handle := range e.MeshComponent.Materials {
				m := app.AssetManager().GetMaterial(handle)
				materialStrs = append(materialStrs, m.Name)
			}
			materialText := "-"
			if len(materialStrs) > 0 {
				materialText = strings.Join(materialStrs, ", ")
			}
			uiTableRow("Materials", materialText)
			imgui.EndTable()
		}
	}

	if e.Physics != nil {
		physicsComponent := e.Physics
		velocity := &physicsComponent.Velocity
		if imgui.CollapsingHeaderTreeNodeFlagsV("Physics Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			var x, y, z int32 = int32(velocity.X()), int32(velocity.Y()), int32(velocity.X())

			ui.RowV("Velocity X", func() {
				imgui.PushIDStr("velocity x")
				if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
					velocity[0] = float64(x)
				}
				imgui.PopID()
			}, true)
			ui.RowV("Velocity Y", func() {
				imgui.PushIDStr("velocity y")
				if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
					velocity[1] = float64(y)
				}
				imgui.PopID()
			}, true)
			ui.RowV("Velocity Z", func() {
				imgui.PushIDStr("velocity z")
				if imgui.InputIntV("", &z, 0, 0, imgui.InputTextFlagsNone) {
					velocity[2] = float64(z)
				}
				imgui.PopID()
			}, true)
			imgui.EndTable()
			imgui.PushIDStr("remove phys")
			if imgui.Button("Remove") {
				e.Physics = nil
			}
			imgui.PopID()
		}
	}
	if e.Kinematic != nil {
		// kinematicComponent := e.Kinematic
		velocity := e.TotalKinematicVelocity()
		if imgui.CollapsingHeaderTreeNodeFlagsV("Kinematic Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			var x, y, z int32 = int32(velocity.X()), int32(velocity.Y()), int32(velocity.Z())

			ui.RowV("Velocity X", func() {
				imgui.PushIDStr("velocity x")
				if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
					velocity[0] = float64(x)
				}
				imgui.PopID()
			}, true)
			ui.RowV("Velocity Y", func() {
				imgui.PushIDStr("velocity y")
				if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
					velocity[1] = float64(y)
				}
				imgui.PopID()
			}, true)
			ui.RowV("Velocity Z", func() {
				imgui.PushIDStr("velocity z")
				if imgui.InputIntV("", &z, 0, 0, imgui.InputTextFlagsNone) {
					velocity[2] = float64(z)
				}
				imgui.PopID()
			}, true)
			ui.RowV("Grounded", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Kinematic.Grounded))
			}, true)
			ui.RowV("Enable Gravity", func() { imgui.Checkbox("", &e.Kinematic.GravityEnabled) }, true)
			imgui.EndTable()
			imgui.PushIDStr("remove kinematic")
			if imgui.Button("Remove") {
				e.Kinematic = nil
			}
			imgui.PopID()
		}
	}

	if e.Collider != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Collider Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			ui.RowV("Collider Type", func() {
				imgui.LabelText("", entity.ColliderFlagToGroupName[e.Collider.ColliderGroup])
			}, true)
			ui.RowV("Capsule", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Collider.CapsuleCollider != nil))
			}, true)
			ui.RowV("Triangular Mesh", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Collider.TriMeshCollider != nil))
			}, true)
			ui.RowV("Bounding Box", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Collider.BoundingBoxCollider != nil))
			}, true)

			simplifiedMeshTriCount := originalMeshTriCount
			if e.Collider.SimplifiedTriMeshCollider != nil {
				simplifiedMeshTriCount = len(e.Collider.SimplifiedTriMeshCollider.Triangles)
			}

			uiTableRow("Triangle Count", simplifiedMeshTriCount)
			imgui.EndTable()
			iterations := app.RuntimeConfig().SimplifyMeshIterations

			parentWidth := imgui.ContentRegionAvail().X
			imgui.PushItemWidth(parentWidth / 2)
			if imgui.InputIntV("##SimplifyMeshIterations", &iterations, 0, 0, imgui.InputTextFlagsNone) {
				app.RuntimeConfig().SimplifyMeshIterations = iterations
			}
			imgui.PopItemWidth()
			imgui.SameLine()
			if imgui.Button("Simplify Mesh") {
				primitives := app.AssetManager().GetPrimitives(e.MeshComponent.MeshHandle)
				specPrimitives := entity.AssetPrimitiveToSpecPrimitive(primitives)
				e.Collider.SimplifiedTriMeshCollider = geometry.SimplifyMesh(specPrimitives[0], int(app.RuntimeConfig().SimplifyMeshIterations))
				e.Collider.SimplifiedTriMeshIterations = int(app.RuntimeConfig().SimplifyMeshIterations)
			}
		}
	}

	if e.Collider != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Debugging Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			collisionObserver := app.CollisionObserver()
			ui.RowV("Entities In Partition", func() { imgui.LabelText("", formatNumber(collisionObserver.SpatialQuery[e.GetID()])) }, true)
			ui.RowV("Bounding Box Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.BoundingBoxCheck[e.GetID()])) }, true)
			ui.RowV("Collision Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheck[e.GetID()])) }, true)
			ui.RowV("Triangle Mesh Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheckTriMesh[e.GetID()])) }, true)
			ui.RowV("Triangle Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheckTriangle[e.GetID()])) }, true)
			ui.RowV("Capsule Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheckCapsule[e.GetID()])) }, true)
			ui.RowV("Collision Resolutions", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionResolution[e.GetID()])) }, true)

			imgui.EndTable()
		}
	}

	if e.Animation != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Animation", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			ui.InitColumns()

			ui.RowV("Current Animation", func() {
				imgui.LabelText("", e.Animation.AnimationPlayer.CurrentAnimation())
			}, true)

			ui.RowV("Animation State", func() {
				var state string

				if app.AppMode() == appmode.Play {
					if e.Animation.Mode == entity.AnimationModeStateMachine {
						state = e.Animation.AnimationStateMachine.CurrentState.Name
					}
				}
				imgui.LabelText("", state)
			}, true)
			ui.RowV("Length (ms)", func() {
				text := ""
				animationName := e.Animation.AnimationPlayer.CurrentAnimation()
				if animation := e.Animation.Animations[animationName]; animation != nil {
					text = fmt.Sprintf("%d", animation.Length.Milliseconds())
				}
				imgui.LabelText("", text)
			}, true)
			ui.RowV("Clip Elapsed Time", func() {
				text := ""
				if e.Animation.AnimationPlayer.CurrentAnimation() != "" {
					elapsedTime := e.Animation.AnimationPlayer.ElapsedTime()
					text = fmt.Sprintf("%d", elapsedTime.Milliseconds())
				}
				imgui.LabelText("", text)
			}, true)
			ui.RowV("Normalized Clip Progress", func() {
				text := ""
				if e.Animation.AnimationPlayer.CurrentAnimation() != "" {
					progress := e.Animation.AnimationPlayer.NormalizedClipProgress()
					text = fmt.Sprintf("%.2f", progress)
				}
				imgui.LabelText("", text)
			}, true)

			ui.RowV("Animation", func() {
				var animationList []string
				for animation := range e.Animation.Animations {
					animationList = append(animationList, animation)
				}
				slices.Sort(animationList)

				comboWidth := imgui.ContentRegionAvail().X
				imgui.SetNextItemWidth(comboWidth)
				imgui.SetNextWindowSizeConstraints(
					imgui.Vec2{X: comboWidth, Y: 0},
					imgui.Vec2{X: comboWidth, Y: float32(math.MaxFloat32)},
				)
				if imgui.BeginComboV("##AnimationCombo", e.Animation.SelectedAnimation, imgui.ComboFlagsPopupAlignLeft) {
					if imgui.IsWindowAppearing() {
						imgui.SetKeyboardFocusHere()
					}
					imgui.SetNextItemWidth(-1)
					imgui.InputTextWithHint("##AnimationComboFilter", "Filter animations", &animationFilterText, imgui.InputTextFlagsNone, nil)
					imgui.Separator()

					filter := strings.ToLower(strings.TrimSpace(animationFilterText))
					matchCount := 0
					imgui.BeginChildStrV("##AnimationComboList", imgui.Vec2{X: 0, Y: animationComboListHeight}, imgui.ChildFlagsNone, imgui.WindowFlagsAlwaysVerticalScrollbar)
					for _, option := range animationList {
						if filter != "" && !strings.Contains(strings.ToLower(option), filter) {
							continue
						}

						matchCount++
						if imgui.SelectableBool(option) {
							e.Animation.SelectedAnimation = option
							e.Animation.SelectedKeyFrame = 0
							animationFilterText = ""
							imgui.CloseCurrentPopup()
						}
					}

					if matchCount == 0 {
						imgui.TextDisabled("No animations found")
					}
					imgui.EndChild()
					imgui.EndCombo()
				}
			}, true)

			ui.RowV("Key Frame", func() {
				currentAnimation := e.Animation.SelectedAnimation
				animations, _, _ := app.AssetManager().GetAnimations(e.Animation.AnimationHandle)
				animation := animations[currentAnimation]

				val := int32(e.Animation.SelectedKeyFrame)
				if animation != nil {
					if imgui.SliderInt("##", &val, 0, int32(len(animation.KeyFrames)-1)) {
						e.Animation.SelectedKeyFrame = int(val)
					}
				}
			}, true)

			ui.RowV("LoopAnimation", func() { imgui.Checkbox("", &e.Animation.LoopAnimation) }, true)

			imgui.EndTable()
		}
	}

	if imgui.BeginCombo("##add_component_combo", string(SelectedComponentComboOption)) {
		for _, option := range componentComboOptions {
			if imgui.SelectableBool(string(option)) {
				SelectedComponentComboOption = option
			}
		}
		imgui.EndCombo()
	}
	imgui.SameLine()
	if imgui.Button("Add Component") {
		selectedEntity := app.SelectedEntity()
		if selectedEntity != nil {
			if SelectedComponentComboOption == LightComboOption {
				selectedEntity.LightInfo = &entity.LightInfo{
					PreScaledIntensity: 0.05,
					Diffuse3F:          [3]float32{1, 1, 1},
					Type:               entity.LightTypePoint,
					Range:              800,
				}
			} else if SelectedComponentComboOption == PhysicsComboOption {
				selectedEntity.Physics = &entity.PhysicsComponent{}
			} else if SelectedComponentComboOption == SpawnPointComboOption {
				selectedEntity.SpawnPointComponent = &entity.SpawnPoint{}
			} else if SelectedComponentComboOption == ImageComboOption {
				selectedEntity.ImageComponent = entity.NewImageComponent("default.png", 1, true)
			}
		}
	}
}

func uiTableInputPosition(e *entity.Entity, text *string) {
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
			entity.SetLocalPosition(e, newPosition)
		}
	}
}

func uiTableInputRow(label string, text *string, cb imgui.InputTextCallback) bool {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Text(label)
	imgui.TableSetColumnIndex(1)

	v := imgui.InputTextWithHint(fmt.Sprintf("##UITableInputRow_%s", label), "", text, imgui.InputTextFlagsCallbackEdit|imgui.InputTextFlagsEnterReturnsTrue, cb)

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
