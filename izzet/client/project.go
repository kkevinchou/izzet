package client

import (
	"github.com/kkevinchou/izzet/izzet/contentbrowser"
	"github.com/kkevinchou/izzet/izzet/materialbrowser"
)

// Project contains engine data that's meant to be persisted and can be reloaded
type Project struct {
	WorldFile string
	Name      string

	ContentBrowser  *contentbrowser.ContentBrowser
	MaterialBrowser *materialbrowser.MaterialBrowser
}

func NewProject() *Project {
	return &Project{
		ContentBrowser:  &contentbrowser.ContentBrowser{},
		MaterialBrowser: &materialbrowser.MaterialBrowser{},
	}
}
