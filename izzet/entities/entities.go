package entities

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
)

var id int

type Entity struct {
	ID     int
	Name   string
	Prefab *prefabs.Prefab

	// each Entity has their own transforms and animation player
	Position mgl64.Vec3
	Rotation mgl64.Quat
	Scale    mgl64.Vec3

	AnimationPlayer *animation.AnimationPlayer
}

func InstantiateFromPrefab(prefab *prefabs.Prefab) *Entity {
	e := &Entity{
		ID:   id,
		Name: fmt.Sprintf("%s-%d", prefab.Name, id),

		Prefab: prefab,
	}

	id += 1
	return e
}
