package client

import (
	"github.com/kkevinchou/izzet/izzet/contentbrowser"
)

type Project struct {
	WorldFile      string
	Name           string
	ContentBrowser contentbrowser.ContentBrowser
}
