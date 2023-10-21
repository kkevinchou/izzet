package app

import "github.com/kkevinchou/izzet/izzet/entities"

func IsCapsuleTriMeshCollision(e1, e2 *entities.Entity) (bool, *entities.Entity, *entities.Entity) {
	if e1.Collider.CapsuleCollider != nil {
		if e2.Collider.TriMeshCollider != nil {
			return true, e1, e2
		}
	}

	if e2.Collider.CapsuleCollider != nil {
		if e1.Collider.TriMeshCollider != nil {
			return true, e2, e1
		}
	}

	return false, nil, nil
}

func IsCapsuleCapsuleCollision(e1, e2 *entities.Entity) bool {
	if e1.Collider.CapsuleCollider != nil {
		if e2.Collider.CapsuleCollider != nil {
			return true
		}
	}

	return false
}
