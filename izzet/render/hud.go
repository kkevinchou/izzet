package render

import (
	"github.com/kkevinchou/izzet/izzet/entities"
)

var gun *entities.Entity

func (r *RenderSystem) hud() {
	if gun == nil {
		document := r.app.AssetManager().GetDocumentAsset("gun_anim")
		gun = r.app.CreateEntitiesFromDocumentAsset(document)
	}
}
