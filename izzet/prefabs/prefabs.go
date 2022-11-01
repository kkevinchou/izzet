package prefabs

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/model"
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
	ModelRefs []*ModelRef
}

func CreatePrefab(name string, models []*model.Model) *Prefab {
	modelRefs := []*ModelRef{}
	for i, model := range models {
		modelRef := &ModelRef{
			Name:  fmt.Sprintf("%s-mesh-%d", name, i),
			Model: model,

			Position: mgl64.Vec3{},
			Rotation: mgl64.QuatIdent(),
			Scale:    mgl64.Vec3{},
		}

		modelRefs = append(modelRefs, modelRef)
	}

	pf := &Prefab{
		ID:        id,
		Name:      name,
		ModelRefs: modelRefs,
	}

	id += 1

	return pf
}
