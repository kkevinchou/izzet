package prefabs

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/model"
	"github.com/kkevinchou/izzet/izzet/settings"
	"github.com/kkevinchou/kitolib/modelspec"
)

// if we update the prefab, instances should be updated as well

var id int

type ModelRef struct {
	Name  string
	Model *model.Model

	// each ModelRef has their own transforms and animation player
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Scale    mgl64.Vec3
}

type Prefab struct {
	ID        int
	Name      string
	modelRefs []*ModelRef

	loaded bool
	scene  *modelspec.Scene
}

func CreatePrefab(name string, scene *modelspec.Scene) *Prefab {
	pf := &Prefab{
		ID:    id,
		Name:  name,
		scene: scene,
	}

	id += 1

	return pf
}

func (p *Prefab) Load() {
	if p.loaded {
		return
	}

	modelConfig := &model.ModelConfig{MaxAnimationJointWeights: settings.MaxAnimationJointWeights}
	models := model.CreateModelsFromScene(p.scene, modelConfig)
	modelRefs := []*ModelRef{}
	for _, model := range models {
		modelRef := &ModelRef{
			Name:  model.Name(),
			Model: model,

			Position: mgl64.Vec3{},
			Rotation: mgl64.QuatIdent(),
			Scale:    mgl64.Vec3{},
		}

		modelRefs = append(modelRefs, modelRef)
	}
	p.modelRefs = modelRefs
	p.loaded = true
}

func (p *Prefab) ModelRefs() []*ModelRef {
	p.Load()
	return p.modelRefs
}
