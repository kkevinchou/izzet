package libutils

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	_ "image/png"
)

type FileMetaData struct {
	Name      string
	Path      string
	Extension string
}

func GetFileMetaData(directory string, subDirectories []string, extensions map[string]any) map[string]FileMetaData {
	var subPaths []string
	for _, subDir := range subDirectories {
		subPaths = append(subPaths, path.Join(directory, subDir))
	}
	if len(subPaths) == 0 {
		subPaths = append(subPaths, directory)
	}

	metaDataCollection := map[string]FileMetaData{}

	for _, subDir := range subPaths {
		files, err := os.ReadDir(subDir)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		for _, file := range files {
			extension := filepath.Ext(file.Name())
			if _, ok := extensions[extension]; !ok {
				continue
			}

			path := filepath.Join(subDir, file.Name())
			name := file.Name()[0 : len(file.Name())-len(extension)]

			metaDataCollection[name] = FileMetaData{Name: name, Path: path, Extension: extension}
		}
	}

	return metaDataCollection
}
