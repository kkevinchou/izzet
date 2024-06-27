package client

import (
	"github.com/kkevinchou/izzet/izzet/contentbrowser"
	"github.com/kkevinchou/izzet/izzet/materialbrowser"
)

type Project struct {
	WorldFile       string
	Name            string
	ContentBrowser  contentbrowser.ContentBrowser
	MaterialBrowser materialbrowser.MaterialBrowser
}
