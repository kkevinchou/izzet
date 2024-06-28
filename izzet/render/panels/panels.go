package panels

import (
	imgui "github.com/AllenDang/cimgui-go"
	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/kitolib/spatialpartition"
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
	SpatialPartition() *spatialpartition.SpatialPartition
}
