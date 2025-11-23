package panels

import (
	"github.com/AllenDang/cimgui-go/imgui"
	"github.com/kkevinchou/izzet/internal/spatialpartition"
	"github.com/kkevinchou/izzet/izzet/entity"
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
	Entities() []*entity.Entity
	AddEntity(entity *entity.Entity)
	GetEntityByID(id int) *entity.Entity
	SpatialPartition() *spatialpartition.SpatialPartition
}
