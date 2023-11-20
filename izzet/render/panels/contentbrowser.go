package panels

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
	"github.com/sqweek/dialog"
)

const (
	maxContentBrowserHeight float32 = 300
)

// TODO
// 1 - Create materials
// 2 - Load 3d meshes
// 3 - Load 3d animations

// In Progress
// 1 - Load 3d Meshes
// 2 - Create thumbnail

// Use cases
// - Drop down to specify import type (mesh, animation, texture)
// - Click Import, filters files basd on import type
// - load 3d mesh
// - create thumbnail
// - render thumbnail to content browser
// - drag and drop from content browser to world space (otherwise, just right click -> instantiate)

var documentTexture *imgui.TextureID

func BuildContentBrowser(app renderiface.App, world GameWorld, renderContext RenderContext, ps []*prefabs.Prefab, x, y float32, height *float32, expanded *bool) {
	// rect := imgui.Vec2{X: float32(renderContext.Width()), Y: float32(renderContext.Height()) - menuBarSize.Y}

	imgui.SetNextWindowBgAlpha(1)
	// imgui.SetNextWindowPosV(imgui.Vec2{X: menuBarSize.X - propertiesWidth, Y: menuBarSize.Y}, imgui.ConditionNone, imgui.Vec2{})
	// imgui.SetNextWindowFocus()
	r := imgui.ContentRegionAvail()
	imgui.SetNextWindowPosV(imgui.Vec2{x, y}, imgui.ConditionNone, imgui.Vec2{})
	imgui.SetNextWindowSize(imgui.Vec2{X: r.X, Y: *height})
	// imgui.SetNextWindowSizeV(imgui.Vec2{X: propertiesWidth, Y: rect.Y}, imgui.ConditionNone)

	// imgui.BeginV("Content Browser", nil, imgui.WindowFlagsNone)
	var open bool = true
	flags := imgui.WindowFlagsNoResize | imgui.WindowFlagsNoMove | imgui.WindowFlagsNoCollapse | imgui.WindowFlagsNoTitleBar
	if !*expanded {
		flags |= imgui.WindowFlagsNoScrollbar
	}
	imgui.BeginV("Content Browser", &open, flags)

	if imgui.BeginTabBarV("Content Browser Tab Bar", imgui.TabBarFlagsFittingPolicyScroll|imgui.TabBarFlagsReorderable) {
		if imgui.BeginTabItem("Content Browser") {
			if imgui.IsItemClicked() {
				*expanded = true
				*height = maxContentBrowserHeight
			}

			if imgui.Button("Import") {
				// loading the asset
				assetFilePath, err := dialog.File().Filter("GLTF file", "gltf").Load()
				if err != nil {
					if err != dialog.ErrCancelled {
						panic(err)
					}
				} else {
					app.ImportToContentBrowser(assetFilePath)
				}
			}

			imgui.EndTabItem()

			for i, item := range app.ContentBrowser().Items {
				size := imgui.Vec2{X: 100, Y: 100}
				// invert the Y axis since opengl vs texture coordinate systems differ
				// https://learnopengl.com/Getting-started/Textures
				imgui.BeginGroup()
				imgui.PushID(fmt.Sprintf("image %d", i))

				if documentTexture == nil {
					t := app.AssetManager().GetTexture("document")
					texture := CreateUserSpaceTextureHandle(t.ID)
					documentTexture = &texture
				}

				imgui.ImageV(*documentTexture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
				if imgui.BeginPopupContextItem() {
					if imgui.Button("Instantiate") {
						document := app.AssetManager().GetDocument(item.Name)
						handle := modellibrary.NewGlobalHandle(item.Name)
						if len(document.Scenes) != 1 {
							panic("single entity asset loading only supports a singular scene")
						}

						scene := document.Scenes[0]
						node := scene.Nodes[0]

						entity := entities.InstantiateEntity(document.Name)
						entity.MeshComponent = &entities.MeshComponent{MeshHandle: handle, Transform: mgl64.Ident4(), Visible: true, ShadowCasting: true}
						var vertices []modelspec.Vertex
						entities.VerticesFromNode(node, document, &vertices)
						entity.InternalBoundingBox = collider.BoundingBoxFromVertices(utils.ModelSpecVertsToVec3(vertices))
						entities.SetLocalPosition(entity, utils.Vec3F32ToF64(node.Translation))
						entities.SetLocalRotation(entity, utils.QuatF32ToF64(node.Rotation))
						entities.SetScale(entity, utils.Vec3F32ToF64(node.Scale))

						world.AddEntity(entity)
						imgui.CloseCurrentPopup()
					}
					imgui.EndPopup()
				}
				imgui.PopID()

				if imgui.BeginDragDropSource(imgui.DragDropFlagsSourceAllowNullID) {
					imgui.SetDragDropPayload("content_browser_item", []byte(item.Name), imgui.ConditionNone)
					imgui.EndDragDropSource()
					fmt.Println("START DRAGGING", item.Name)
				}

				imgui.Text(item.Name)
				imgui.EndGroup()
				imgui.SameLine()
			}
		}
		imgui.EndTabBar()
	}

	if imgui.IsWindowHovered() && imgui.IsMouseClicked(0) {
		*expanded = true
		*height = maxContentBrowserHeight
	}

	imgui.End()
}
