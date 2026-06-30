package entity

func CreateSocket() *Entity {
	entity := InstantiateBaseEntity("socket", entityIDGen)
	entityIDGen += 1

	return entity
}
