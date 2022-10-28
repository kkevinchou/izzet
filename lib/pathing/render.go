package pathing

import (
	"github.com/go-gl/mathgl/mgl64"
)

type NavMeshRenderData struct {
	ID      string
	Visible bool
}

func (n *NavMeshRenderData) IsVisible() bool {
	return true
}

type RenderComponent struct {
	RenderData *NavMeshRenderData
}

func (n *RenderComponent) Position() mgl64.Vec3 {
	return mgl64.Vec3{}
}

func (n *RenderComponent) SetPosition(v mgl64.Vec3) {
}
