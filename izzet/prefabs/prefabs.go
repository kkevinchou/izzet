package prefabs

import (
	"github.com/kkevinchou/kitolib/modelspec"
)

var id int

// if we update the prefab, instances should be updated as well

type Prefab struct {
	ID       int
	Name     string
	Document *modelspec.Document
}

func CreatePrefab(document *modelspec.Document) *Prefab {
	pf := &Prefab{
		ID:       id,
		Name:     document.Name,
		Document: document,
	}

	id += 1

	return pf
}
