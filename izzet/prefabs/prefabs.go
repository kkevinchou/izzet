package prefabs

import (
	"github.com/kkevinchou/izzet/izzet/izzetdata"
	"github.com/kkevinchou/kitolib/modelspec"
)

var id int

// if we update the prefab, instances should be updated as well

type Prefab struct {
	ID        int
	Name      string
	Document  *modelspec.Document
	IzzetData *izzetdata.Data
}

func CreatePrefab(document *modelspec.Document, data *izzetdata.Data) *Prefab {
	pf := &Prefab{
		ID:        id,
		Name:      document.Name,
		Document:  document,
		IzzetData: data,
	}

	id += 1

	return pf
}
