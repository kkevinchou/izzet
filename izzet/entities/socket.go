package entities

func CreateSocket() *Entity {
	entity := InstantiateBaseEntity("socket", id)
	id += 1

	return entity
}
