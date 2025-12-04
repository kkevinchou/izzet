package panels

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/geometry"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/render/panels/panelutils"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/types"
)

type ComponentComboOption string

var MaterialComboOption ComponentComboOption = "Material Component"
var PhysicsComboOption ComponentComboOption = "Physics Component"
var LightComboOption ComponentComboOption = "Light Component"
var SpawnPointComboOption ComponentComboOption = "Spawn Point Component"
var SelectedComponentComboOption ComponentComboOption = MaterialComboOption

var componentComboOptions []ComponentComboOption = []ComponentComboOption{
	MaterialComboOption,
	PhysicsComboOption,
	LightComboOption,
	SpawnPointComboOption,
}

var (
	selectedMaterialHandle types.MaterialHandle
	selectedMaterialName   string
)

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
		panelutils.InitColumns()
		uiTableRow("ID", entityIDStr)
		uiTableRow("Name", entityNameStr)

		var position *mgl64.Vec3
		var x, y, z float32
		if e != nil {
			position = &e.LocalPosition
			x, y, z = float32(position.X()), float32(position.Y()), float32(position.Z())
		}

		panelutils.SetupRow("Local Position", func() {
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
		}
		imgui.EndTable()
	}

	if e == nil {
		return
	}

	if e.CameraComponent != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Camera Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			panelutils.InitColumns()

			var position *mgl64.Vec3
			var x, y, z int32
			if e != nil {
				position = &e.CameraComponent.TargetPositionOffset
				x, y, z = int32(position.X()), int32(position.Y()), int32(position.Z())
			}

			panelutils.SetupRow("Target Position Offset", func() {
				imgui.PushItemWidth(imgui.ContentRegionAvail().X / 3.0)
				imgui.PushIDStr("position x")
				if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
					if e != nil {
						position[0] = float64(x)
						entity.SetDirty(e)
					}
				}
				imgui.PopID()
				imgui.SameLine()
				imgui.PushIDStr("position y")
				if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
					if e != nil {
						position[1] = float64(y)
						entity.SetDirty(e)
					}
				}
				imgui.PopID()
				imgui.SameLine()
				imgui.PushIDStr("position z")
				if imgui.InputIntV("", &z, 0, 0, imgui.InputTextFlagsNone) {
					if e != nil {
						position[2] = float64(z)
						entity.SetDirty(e)
					}
				}
				imgui.PopID()
				imgui.PopItemWidth()
			}, false)

			panelutils.SetupRow("Target ID", func() {
				var target int32

				if e.CameraComponent.Target != nil {
					target = int32(*e.CameraComponent.Target)
				}

				if imgui.InputInt("", &target) {
					intTarget := int(target)
					e.CameraComponent.Target = &intTarget
				}
			}, true)

			imgui.EndTable()
		}
	}

	if e.LightInfo != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Light Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			panelutils.InitColumns()

			lightTypeStr := "?"
			if e.LightInfo.Type == entity.LightTypePoint {
				lightTypeStr = "Point Light"
			} else if e.LightInfo.Type == entity.LightTypeDirection {
				lightTypeStr = "Directional Light"
			}
			uiTableRow("Light Type", lightTypeStr)
			panelutils.SetupRow("Color", func() {
				imgui.ColorEdit3V("", &e.LightInfo.Diffuse3F, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			}, true)
			panelutils.SetupRow("Color Intensity", func() {
				if e.LightInfo.Type == entity.LightTypePoint {
					imgui.SliderFloatV("", &e.LightInfo.PreScaledIntensity, 0, 0.1, "%.3f", imgui.SliderFlagsNone)
				} else if e.LightInfo.Type == entity.LightTypeDirection {
					imgui.SliderFloatV("", &e.LightInfo.PreScaledIntensity, 0, 6, "%.3f", imgui.SliderFlagsNone)
				}
			}, true)

			if e.LightInfo.Type == entity.LightTypePoint {
				panelutils.SetupRow("Light Range", func() { imgui.SliderFloatV("", &e.LightInfo.Range, 1, 1500, "%.0f", imgui.SliderFlagsNone) }, true)
			} else if e.LightInfo.Type == entity.LightTypeDirection {
				panelutils.SetupRow("Directional Light Direction", func() { imgui.SliderFloat3("", &e.LightInfo.Direction3F, -1, 1) }, true)
			}
			imgui.EndTable()
			imgui.PushIDStr("remove light")
			if imgui.Button("Remove") {
				e.LightInfo = nil
			}
			imgui.PopID()
		}
	}

	if e.Material != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Material Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			panelutils.InitColumns()

			// panelutils.SetupRow("Diffuse", func() {
			// 	imgui.ColorEdit3V("", &entity.Material.Material.PBR.Diffuse, imgui.ColorEditFlagsNoInputs|imgui.ColorEditFlagsNoLabel)
			// }, true)
			// panelutils.SetupRow("Invisible", func() {
			// 	imgui.Checkbox("", &entity.Material.Material.Invisible)
			// }, true)

			// panelutils.SetupRow("Diffuse Intensity", func() {
			// 	imgui.SliderFloatV("", &entity.Material.Material.PBR.DiffuseIntensity, 1, 100, "%.1f", imgui.SliderFlagsNone)
			// }, true)

			// panelutils.SetupRow("Roughness", func() {
			// 	imgui.SliderFloatV("", &entity.Material.Material.PBR.Roughness, 0, 1, "%.2f", imgui.SliderFlagsNone)
			// }, true)
			// panelutils.SetupRow("Metallic Factor", func() {
			// 	imgui.SliderFloatV("", &entity.Material.Material.PBR.Metallic, 0, 1, "%.2f", imgui.SliderFlagsNone)
			// }, true)
			panelutils.SetupRow("Current Material", func() {
				materialName := app.AssetManager().GetMaterial(e.Material.MaterialHandle).Name
				imgui.LabelText("", materialName)
			}, true)
			imgui.EndTable()
			imgui.PushIDStr("Material Combo")
			if imgui.BeginCombo("", selectedMaterialName) {
				for i, material := range app.AssetManager().GetMaterials() {
					if imgui.SelectableBool(fmt.Sprintf("%s##%d", material.Name, i)) {
						selectedMaterialHandle = material.Handle
						selectedMaterialName = material.Name
					}
				}
				imgui.EndCombo()
			}
			imgui.PopID()
			imgui.PushIDStr("assign")
			if imgui.Button("Assign") {
				// material := app.AssetManager().GetMaterial(selectedMaterialHandle)
				e.Material.MaterialHandle = selectedMaterialHandle
			}
			imgui.PopID()
			imgui.SameLine()
			imgui.PushIDStr("remove material")
			if imgui.Button("Remove") {
				e.Material = nil
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
			panelutils.InitColumns()
			panelutils.SetupRow("Visible", func() { imgui.Checkbox("", &e.MeshComponent.Visible) }, true)
			panelutils.SetupRow("Shadow Casting", func() { imgui.Checkbox("", &e.MeshComponent.ShadowCasting) }, true)

			uiTableRow("Original Triangle Count", originalMeshTriCount)
			imgui.EndTable()
		}
	}

	if e.Physics != nil {
		physicsComponent := e.Physics
		velocity := &physicsComponent.Velocity
		if imgui.CollapsingHeaderTreeNodeFlagsV("Physics Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			panelutils.InitColumns()

			var x, y, z int32 = int32(velocity.X()), int32(velocity.Y()), int32(velocity.X())

			panelutils.SetupRow("Velocity X", func() {
				imgui.PushIDStr("velocity x")
				if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
					velocity[0] = float64(x)
				}
				imgui.PopID()
			}, true)
			panelutils.SetupRow("Velocity Y", func() {
				imgui.PushIDStr("velocity y")
				if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
					velocity[1] = float64(y)
				}
				imgui.PopID()
			}, true)
			panelutils.SetupRow("Velocity Z", func() {
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
			panelutils.InitColumns()

			var x, y, z int32 = int32(velocity.X()), int32(velocity.Y()), int32(velocity.Z())

			panelutils.SetupRow("Velocity X", func() {
				imgui.PushIDStr("velocity x")
				if imgui.InputIntV("", &x, 0, 0, imgui.InputTextFlagsNone) {
					velocity[0] = float64(x)
				}
				imgui.PopID()
			}, true)
			panelutils.SetupRow("Velocity Y", func() {
				imgui.PushIDStr("velocity y")
				if imgui.InputIntV("", &y, 0, 0, imgui.InputTextFlagsNone) {
					velocity[1] = float64(y)
				}
				imgui.PopID()
			}, true)
			panelutils.SetupRow("Velocity Z", func() {
				imgui.PushIDStr("velocity z")
				if imgui.InputIntV("", &z, 0, 0, imgui.InputTextFlagsNone) {
					velocity[2] = float64(z)
				}
				imgui.PopID()
			}, true)
			panelutils.SetupRow("Grounded", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Kinematic.Grounded))
			}, true)
			panelutils.SetupRow("Enable Gravity", func() { imgui.Checkbox("", &e.Kinematic.GravityEnabled) }, true)
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
			panelutils.InitColumns()

			panelutils.SetupRow("Collider Type", func() {
				imgui.LabelText("", string(types.ColliderFlagToGroupName[e.Collider.ColliderGroup]))
			}, true)
			panelutils.SetupRow("Capsule", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Collider.CapsuleCollider != nil))
			}, true)
			panelutils.SetupRow("Triangular Mesh", func() {
				imgui.LabelText("", fmt.Sprintf("%t", e.Collider.TriMeshCollider != nil))
			}, true)
			panelutils.SetupRow("Bounding Box", func() {
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
				e.SimplifiedTriMeshIterations = int(app.RuntimeConfig().SimplifyMeshIterations)
			}
		}
	}

	imgui.PushIDStr("Component Combo")
	if imgui.BeginCombo("", string(SelectedComponentComboOption)) {
		for _, option := range componentComboOptions {
			if imgui.SelectableBool(string(option)) {
				SelectedComponentComboOption = option
			}
		}
		imgui.EndCombo()
	}
	imgui.PopID()
	imgui.SameLine()
	if imgui.Button("Add Component") {
		selectedEntity := app.SelectedEntity()
		if selectedEntity != nil {
			if SelectedComponentComboOption == MaterialComboOption {
				selectedEntity.Material = &entity.MaterialComponent{
					MaterialHandle: assets.DefaultMaterialHandle,
				}
			} else if SelectedComponentComboOption == LightComboOption {
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
			}
		}
	}

	if e.Collider != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Debugging Properties", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			panelutils.InitColumns()

			collisionObserver := app.CollisionObserver()
			panelutils.SetupRow("Entities In Partition", func() { imgui.LabelText("", formatNumber(collisionObserver.SpatialQuery[e.GetID()])) }, true)
			panelutils.SetupRow("Bounding Box Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.BoundingBoxCheck[e.GetID()])) }, true)
			panelutils.SetupRow("Collision Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheck[e.GetID()])) }, true)
			panelutils.SetupRow("Triangle Mesh Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheckTriMesh[e.GetID()])) }, true)
			panelutils.SetupRow("Triangle Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheckTriangle[e.GetID()])) }, true)
			panelutils.SetupRow("Capsule Checks", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionCheckCapsule[e.GetID()])) }, true)
			panelutils.SetupRow("Collision Resolutions", func() { imgui.LabelText("", formatNumber(collisionObserver.CollisionResolution[e.GetID()])) }, true)

			imgui.EndTable()
		}
	}

	if e.Animation != nil {
		if imgui.CollapsingHeaderTreeNodeFlagsV("Animation", imgui.TreeNodeFlagsNone) {
			imgui.BeginTableV("", 2, imgui.TableFlagsBorders|imgui.TableFlagsResizable, imgui.Vec2{}, 0)
			panelutils.InitColumns()

			panelutils.SetupRow("Animation", func() {
				var animationList []string
				for animation := range e.Animation.Animations {
					animationList = append(animationList, animation)
				}
				slices.Sort(animationList)

				imgui.PushIDStr("Animation Combo")
				if imgui.BeginCombo("", app.RuntimeConfig().SelectedAnimation) {
					for _, option := range animationList {
						if imgui.SelectableBool(option) {
							app.RuntimeConfig().SelectedAnimation = option
							app.RuntimeConfig().SelectedKeyFrame = 0
						}
					}
					imgui.EndCombo()
				}
				imgui.PopID()
			}, true)

			panelutils.SetupRow("Key Frame", func() {
				currentAnimation := app.RuntimeConfig().SelectedAnimation
				animations, _, _ := app.AssetManager().GetAnimations(e.Animation.AnimationHandle)
				animation := animations[currentAnimation]

				val := int32(app.RuntimeConfig().SelectedKeyFrame)
				if animation != nil {
					if imgui.SliderInt("##", &val, 0, int32(len(animation.KeyFrames)-1)) {
						app.RuntimeConfig().SelectedKeyFrame = int(val)
					}
				}
			}, true)

			panelutils.SetupRow("LoopAnimation", func() { imgui.Checkbox("", &app.RuntimeConfig().LoopAnimation) }, true)

			imgui.EndTable()
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
