package entities

func CreateSocket() *Entity {
	entity := InstantiateBaseEntity("socket", entityIDGen)
	entity.IsSocket = true
	entityIDGen += 1

	return entity
}
