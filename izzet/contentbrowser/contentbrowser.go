package contentbrowser

import (
	"path/filepath"
	"strings"

	"github.com/kkevinchou/kitolib/modelspec"
)

type ContentType string

const (
	ContentTypeGLTFModel = "GLTFModel"
)

type ContentBrowser struct {
	Items []Content
}

type Content struct {
	Name                 string
	Type                 ContentType
	SavedToProjectFolder bool

	InFilePath      string
	PeripheralFiles []string
}

func (cb *ContentBrowser) AddGLTFModel(sourceFile string, document *modelspec.Document) {
	baseFileName := strings.Split(filepath.Base(sourceFile), ".")[0]
	cb.Items = append(cb.Items,
		Content{
			Name:                 baseFileName,
			Type:                 ContentTypeGLTFModel,
			SavedToProjectFolder: false,

			InFilePath:      sourceFile,
			PeripheralFiles: document.PeripheralFiles,
		},
	)
}
