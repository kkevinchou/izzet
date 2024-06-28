package materialbrowser

import (
	"fmt"

	"github.com/kkevinchou/izzet/izzet/material"
)

type MaterialBrowser struct {
	Items []material.Material
}

func (m *MaterialBrowser) AddMaterial(material material.Material) {
	m.Items = append(m.Items, material)
	fmt.Println(len(m.Items), "total materials")
}
