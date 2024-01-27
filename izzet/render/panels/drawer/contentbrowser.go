package drawer

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/render/renderiface"
	"github.com/kkevinchou/izzet/izzet/render/renderutils"
	"github.com/kkevinchou/kitolib/collision/collider"
	"github.com/kkevinchou/kitolib/modelspec"
	"github.com/kkevinchou/kitolib/utils"
	"github.com/sqweek/dialog"
)

func contentBrowser(app renderiface.App, world renderiface.GameWorld) bool {
	var menuOpen bool
	if imgui.BeginTabItem("Content Browser") {
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
				texture := renderutils.CreateUserSpaceTextureHandle(t.ID)
				documentTexture = &texture
			}

			imgui.ImageV(*documentTexture, size, imgui.Vec2{X: 0, Y: 1}, imgui.Vec2{X: 1, Y: 0}, imgui.Vec4{X: 1, Y: 1, Z: 1, W: 1}, imgui.Vec4{X: 0, Y: 0, Z: 0, W: 0})
			if imgui.BeginPopupContextItem() {
				menuOpen = true
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
	return menuOpen
}
