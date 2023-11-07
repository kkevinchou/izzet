package entities

type Filter struct {
	filters []FilterFunction
}

type FilterFunction func(entity *Entity) bool

// var ShadowCasting Filter = Filter{
// 	filters: []FilterFunction{ShadowCastingFunc},
// }

var ShadowCasting = func(entity *Entity) bool {
	return entity.MeshComponent != nil && entity.MeshComponent.ShadowCasting
}

var Renderable = func(entity *Entity) bool {
	return entity.MeshComponent != nil && entity.MeshComponent.Visible
}

var EmptyFilter = func(entity *Entity) bool {
	return true
}
