package entities

func CreateSocket() *Entity {
	entity := InstantiateBaseEntity("socket", id)
	entity.IsSocket = true
	id += 1

	return entity
}
