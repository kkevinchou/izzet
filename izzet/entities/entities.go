package entities

import (
	"github.com/go-gl/mathgl/mgl64"
	"github.com/kkevinchou/kitolib/animation"
	"github.com/kkevinchou/kitolib/model"
)

type Entity struct {
	Name            string
	Position        mgl64.Vec3
	Model           *model.Model
	AnimationPlayer *animation.AnimationPlayer
}
