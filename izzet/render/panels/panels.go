package panels

import (
	"github.com/inkyblackness/imgui-go/v4"
	"github.com/kkevinchou/izzet/izzet/entities"
)

const (
	propertiesWidth float32 = 450

	tableColumn0Width float32          = 180
	tableFlags        imgui.TableFlags = imgui.TableFlagsBordersInnerV
)

type RenderContext interface {
	Width() int
	Height() int
	AspectRatio() float64
}

type GameWorld interface {
	Entities() []*entities.Entity
	AddEntity(entity *entities.Entity)
	GetEntityByID(id int) *entities.Entity
}

// createUserSpaceTextureHandle creates a handle to a user space texture
// that the imgui renderer is able to render
func CreateUserSpaceTextureHandle(texture uint32) imgui.TextureID {
	handle := 1<<63 | uint64(texture)
	return imgui.TextureID(handle)
}

func setupRow(label string, item func(), fillWidth bool) {
	imgui.TableNextRow()
	imgui.TableNextColumn()
	imgui.Text(label)
	imgui.TableNextColumn()
	imgui.PushID(label)
	if fillWidth {
		imgui.PushItemWidth(-1)
	}
	item()
	if fillWidth {
		imgui.PopItemWidth()
	}
	imgui.PopID()
}

func initColumns() {
	imgui.TableSetupColumnV("0", imgui.TableColumnFlagsWidthFixed|imgui.TableColumnFlagsNoResize, tableColumn0Width, 0)
}
