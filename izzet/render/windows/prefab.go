package windows

import (
	"fmt"
	"math"
	"strings"

	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/internal/collision/collider"
	"github.com/kkevinchou/izzet/izzet/animation"
	"github.com/kkevinchou/izzet/izzet/assets"
	"github.com/kkevinchou/izzet/izzet/entity"
	"github.com/kkevinchou/izzet/izzet/prefab"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/ui"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/izzet/izzet/types"
)

var showPrefabWindow bool

type prefabEditorState struct {
	Name  string
	Scale float32

	IncludeMesh          bool
	MeshSourceAsset      string
	MeshMaterials        []assets.MaterialID
	IncludeAnimation     bool
	AnimationSourceAsset string
	StateMachineID       animation.StateMachineID
	IncludeCapsule       bool
	CapsuleRadius        float32
	CapsuleLength        float32
	IncludeKinematic     bool
	KinematicGravity     bool
	KinematicSpeed       float32
	IncludeHealth        bool
	HealthAmount         int32
	IncludeAimDownSights bool
	IncludeAttack        bool
	AttackRange          float32
	IncludeAI            bool

	Error string
}

var activePrefabEditor prefabEditorState

const (
	defaultPrefabName            string  = "<new prefab>"
	prefabEditorLabelColumnWidth float32 = 260
	prefabEditorPropertyIndent   float32 = 18
	prefabWindowWidth            float32 = 720
	prefabWindowHeight           float32 = 680
)

var (
	prefabEditorHeaderTextColor   = imgui.Vec4{X: 0.92, Y: 0.92, Z: 0.92, W: 1}
	prefabEditorPropertyTextColor = imgui.Vec4{X: 0.65, Y: 0.65, Z: 0.65, W: 1}
)

func ShowCreatePrefabWindow(app renderiface.App) {
	showPrefabWindow = true
	assignDefaultPrefab(app)
}

func renderPrefabWindow(app renderiface.App) {
	if !showPrefabWindow {
		return
	}

	center := imgui.MainViewport().Center()
	imgui.SetNextWindowPosV(center, imgui.CondAppearing, imgui.Vec2{X: 0.5, Y: 0.5})
	imgui.SetNextWindowSizeV(imgui.Vec2{X: prefabWindowWidth, Y: prefabWindowHeight}, imgui.CondAppearing)
	imgui.PushStyleVarVec2(imgui.StyleVarWindowPadding, imgui.Vec2{X: 12, Y: 12})
	defer imgui.PopStyleVar()

	if imgui.BeginV("Create Prefab", &showPrefabWindow, imgui.WindowFlagsNone) {
		renderPrefabEditor(app)
	}
	imgui.End()
}

func renderPrefabEditor(app renderiface.App) {
	originalColumnWidth := ui.TableColumn0Width
	ui.TableColumn0Width = prefabEditorLabelColumnWidth
	defer func() {
		ui.TableColumn0Width = originalColumnWidth
	}()

	ui.Table("Prefab Editor", func() {
		ui.Row("Prefab Name", func() {
			imgui.InputTextWithHint("##value", defaultPrefabName, &activePrefabEditor.Name, imgui.InputTextFlagsNone, nil)
		})

		sectionHeading("Transform")
		inputFloatRow("Scale", &activePrefabEditor.Scale, 0.1, 5, "%.2f")
		componentSection("Mesh", &activePrefabEditor.IncludeMesh, func() {
			propertyRow("Source Asset", func() {
				renderSourceAssetCombo(app, &activePrefabEditor.MeshSourceAsset)
			})
			renderPrefabMaterialSlots(app)
		})
		componentSection("Animation", &activePrefabEditor.IncludeAnimation, func() {
			propertyRow("Source Asset", func() {
				renderSourceAssetCombo(app, &activePrefabEditor.AnimationSourceAsset)
			})
			propertyRow("State Machine", func() {
				renderStateMachineCombo(&activePrefabEditor.StateMachineID)
			})
		})
		componentSection("Capsule Collider", &activePrefabEditor.IncludeCapsule, func() {
			inputFloatRow("Radius", &activePrefabEditor.CapsuleRadius, 0.01, 10, "%.2f")
			inputFloatRow("Length", &activePrefabEditor.CapsuleLength, 0.01, 10, "%.2f")
		})
		componentSection("Kinematic", &activePrefabEditor.IncludeKinematic, func() {
			checkboxPropertyRow("Gravity", &activePrefabEditor.KinematicGravity)
			inputFloatRow("Speed", &activePrefabEditor.KinematicSpeed, 0.1, 100, "%.2f")
		})
		componentSection("Health", &activePrefabEditor.IncludeHealth, func() {
			inputIntRow("Health Amount", &activePrefabEditor.HealthAmount)
		})
		componentSection("Aim Down Sights", &activePrefabEditor.IncludeAimDownSights, nil)
		componentSection("Attack", &activePrefabEditor.IncludeAttack, func() {
			inputFloatRow("Attack Range", &activePrefabEditor.AttackRange, 0.1, 100, "%.2f")
		})
		componentSection("AI", &activePrefabEditor.IncludeAI, nil)
	})

	if activePrefabEditor.Error != "" {
		imgui.PushStyleColorVec4(imgui.ColText, imgui.Vec4{X: 1, Y: 0.25, Z: 0.25, W: 1})
		imgui.Text(activePrefabEditor.Error)
		imgui.PopStyleColor()
	}

	if imgui.Button("Save") {
		if err := savePrefab(app); err != nil {
			activePrefabEditor.Error = err.Error()
		} else {
			showPrefabWindow = false
			assignDefaultPrefab(app)
		}
	}
	imgui.SameLine()
	if imgui.Button("Cancel") {
		showPrefabWindow = false
		assignDefaultPrefab(app)
	}
}

func savePrefab(app renderiface.App) error {
	name := strings.TrimSpace(activePrefabEditor.Name)
	if name == "" {
		return fmt.Errorf("prefab name is required")
	}
	if activePrefabEditor.Scale <= 0 {
		return fmt.Errorf("scale must be greater than zero")
	}
	if activePrefabEditor.IncludeCapsule && (activePrefabEditor.CapsuleRadius <= 0 || activePrefabEditor.CapsuleLength <= 0) {
		return fmt.Errorf("capsule dimensions must be greater than zero")
	}
	if activePrefabEditor.IncludeKinematic && activePrefabEditor.KinematicSpeed < 0 {
		return fmt.Errorf("kinematic speed cannot be negative")
	}
	if activePrefabEditor.IncludeHealth && activePrefabEditor.HealthAmount < 0 {
		return fmt.Errorf("health cannot be negative")
	}
	if activePrefabEditor.IncludeAttack && activePrefabEditor.AttackRange < 0 {
		return fmt.Errorf("attack range cannot be negative")
	}

	animationHandle := app.AssetManager().GetAnimationHandle(activePrefabEditor.AnimationSourceAsset)
	if activePrefabEditor.IncludeAnimation && !hasAnimations(app.AssetManager(), animationHandle) {
		return fmt.Errorf("source asset [%s] has no animations", activePrefabEditor.AnimationSourceAsset)
	}

	template := buildPrefabTemplate(app, name)
	return prefab.RegisterPrefab(name, template)
}

func buildPrefabTemplate(app renderiface.App, prefabName string) *entity.Entity {
	template := entity.InstantiateBaseEntity(prefabName, 0)
	scale := float64(activePrefabEditor.Scale)
	entity.SetScale(template, mgl64.Vec3{scale, scale, scale})

	if activePrefabEditor.IncludeMesh {
		template.MeshComponent = &entity.MeshComponent{
			MeshHandle:    app.AssetManager().GetSingleEntityMeshHandle(activePrefabEditor.MeshSourceAsset),
			Materials:     append([]assets.MaterialID{}, activePrefabEditor.MeshMaterials...),
			Transform:     mgl64.Rotate3DY(math.Pi).Mat4(),
			Visible:       true,
			ShadowCasting: true,
		}
	}

	if activePrefabEditor.IncludeAnimation {
		animationHandle := app.AssetManager().GetAnimationHandle(activePrefabEditor.AnimationSourceAsset)
		template.Animation = entity.NewAnimationComponent(app.AssetManager(), animationHandle, activePrefabEditor.StateMachineID, entity.AnimationModeStateMachine)
	}

	if activePrefabEditor.IncludeCapsule {
		radius := float64(activePrefabEditor.CapsuleRadius)
		length := float64(activePrefabEditor.CapsuleLength)
		capsule := collider.Capsule{
			Radius: radius,
			Top:    mgl64.Vec3{0, radius + length, 0},
			Bottom: mgl64.Vec3{0, radius, 0},
		}
		template.Collider = entity.CreateCapsuleColliderComponent(types.ColliderGroupFlagPlayer, types.ColliderGroupFlagTerrain|types.ColliderGroupFlagPlayer, capsule)
	}

	if activePrefabEditor.IncludeKinematic {
		template.Kinematic = &entity.KinematicComponent{
			GravityEnabled: activePrefabEditor.KinematicGravity,
			Speed:          float64(activePrefabEditor.KinematicSpeed),
		}
	}
	if activePrefabEditor.IncludeHealth {
		template.HealthComponent = &entity.HealthComponent{Amount: int(activePrefabEditor.HealthAmount)}
	}
	if activePrefabEditor.IncludeAimDownSights {
		template.AimDownSightsComponent = &entity.AimDownSightsComponent{}
	}
	if activePrefabEditor.IncludeAttack {
		template.AttackComponent = &entity.AttackComponent{AttackRange: float64(activePrefabEditor.AttackRange)}
	}
	if activePrefabEditor.IncludeAI {
		template.AIComponent = &entity.AIComponent{}
	}

	return template
}

func assignDefaultPrefab(app renderiface.App) {
	var sourceAssetName string
	if documents := app.AssetManager().GetDocuments(); len(documents) > 0 {
		sourceAssetName = documents[0].ID
	}

	scale := float32(1)
	activePrefabEditor = prefabEditorState{
		Name:                 defaultPrefabName,
		Scale:                scale,
		IncludeMesh:          true,
		MeshSourceAsset:      sourceAssetName,
		IncludeAnimation:     true,
		AnimationSourceAsset: sourceAssetName,
		StateMachineID:       animation.StateMachineIDPlayer,
		IncludeCapsule:       true,
		CapsuleRadius:        float32(settings.EntityCapsuleColliderRadius) * (1 / scale),
		CapsuleLength:        float32(settings.EntityCapsuleColliderLength) * (1 / scale),
		IncludeKinematic:     true,
		KinematicGravity:     true,
		KinematicSpeed:       7,
		IncludeHealth:        true,
		HealthAmount:         100,
		IncludeAimDownSights: true,
		IncludeAttack:        true,
		AttackRange:          3,
		IncludeAI:            true,
	}
}

func sectionHeading(label string) {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	sectionHeaderText(label)
	imgui.TableSetColumnIndex(1)
	imgui.Separator()
}

func componentSection(label string, enabled *bool, body func()) {
	imgui.PushIDStr(label)
	defer imgui.PopID()

	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	imgui.Checkbox("##enabled", enabled)
	imgui.SameLine()
	sectionHeaderText(label)
	imgui.TableSetColumnIndex(1)
	imgui.Separator()

	if *enabled && body != nil {
		body()
	}
}

func sectionHeaderText(label string) {
	imgui.PushStyleColorVec4(imgui.ColText, prefabEditorHeaderTextColor)
	imgui.Text(label)
	imgui.PopStyleColor()
}

func propertyRow(label string, body func()) {
	imgui.TableNextRow()
	imgui.TableSetColumnIndex(0)
	cursor := imgui.CursorPos()
	imgui.SetCursorPosX(cursor.X + prefabEditorPropertyIndent)
	imgui.PushStyleColorVec4(imgui.ColText, prefabEditorPropertyTextColor)
	imgui.Text(label)
	imgui.PopStyleColor()
	imgui.TableSetColumnIndex(1)
	imgui.PushIDStr(label)
	imgui.PushItemWidth(-1)
	body()
	imgui.PopItemWidth()
	imgui.PopID()
}

func checkboxPropertyRow(label string, value *bool) {
	propertyRow(label, func() {
		imgui.Checkbox("##value", value)
	})
}

func renderSourceAssetCombo(app renderiface.App, sourceAsset *string) {
	documents := app.AssetManager().GetDocuments()
	if len(documents) == 0 {
		imgui.TextDisabled("No documents")
		return
	}

	if *sourceAsset == "" {
		*sourceAsset = documents[0].ID
	}

	if imgui.BeginCombo("##value", *sourceAsset) {
		for _, document := range documents {
			if imgui.SelectableBool(document.ID) {
				*sourceAsset = document.ID
			}
		}
		imgui.EndCombo()
	}
}

func renderStateMachineCombo(selected *animation.StateMachineID) {
	ids := animation.StateMachineIDs()
	if len(ids) == 0 {
		imgui.TextDisabled("No state machines")
		return
	}

	if *selected == "" {
		*selected = ids[0]
	}

	if imgui.BeginCombo("##value", string(*selected)) {
		for _, id := range ids {
			if imgui.SelectableBool(string(id)) {
				*selected = id
			}
		}
		imgui.EndCombo()
	}
}

func renderPrefabMaterialSlots(app renderiface.App) {
	materials := app.AssetManager().GetMaterials()

	propertyRow("Material Slots", func() {
		if imgui.Button("Append") {
			activePrefabEditor.MeshMaterials = append(
				activePrefabEditor.MeshMaterials,
				app.AssetManager().DefaultMaterialID(),
			)
		}
	})

	removeIndex := -1
	for i := range activePrefabEditor.MeshMaterials {
		index := i
		propertyRow(fmt.Sprintf("Material Slot %d", index), func() {
			comboWidth := imgui.ContentRegionAvail().X - 90
			if comboWidth < 120 {
				comboWidth = imgui.ContentRegionAvail().X
			}
			imgui.SetNextItemWidth(comboWidth)
			renderMaterialIDCombo("##material", &activePrefabEditor.MeshMaterials[index], materials)
			imgui.SameLine()
			if imgui.Button("Remove") {
				removeIndex = index
			}
		})
	}

	if removeIndex >= 0 {
		activePrefabEditor.MeshMaterials = append(
			activePrefabEditor.MeshMaterials[:removeIndex],
			activePrefabEditor.MeshMaterials[removeIndex+1:]...,
		)
	}
}

func renderMaterialIDCombo(id string, selected *assets.MaterialID, materials []assets.Material) {
	if len(materials) == 0 {
		imgui.TextDisabled("No materials")
		return
	}

	if imgui.BeginCombo(id, materialIDName(*selected, materials)) {
		for _, material := range materials {
			if imgui.SelectableBool(material.Name) {
				*selected = material.ID
			}
		}
		imgui.EndCombo()
	}
}

func materialIDName(id assets.MaterialID, materials []assets.Material) string {
	for _, material := range materials {
		if material.ID == id {
			return material.Name
		}
	}
	return ""
}

func inputFloatRow(label string, value *float32, step float32, fastStep float32, format string) {
	propertyRow(label, func() {
		imgui.InputFloatV("##value", value, step, fastStep, format, imgui.InputTextFlagsNone)
	})
}

func inputIntRow(label string, value *int32) {
	propertyRow(label, func() {
		imgui.InputIntV("##value", value, 0, 0, imgui.InputTextFlagsNone)
	})
}

func hasAnimations(assetManager *assets.AssetManager, animationHandle assets.AnimationHandle) bool {
	animations, joints, _ := assetManager.GetAnimations(animationHandle)
	return len(animations) > 0 && len(joints) > 0
}
