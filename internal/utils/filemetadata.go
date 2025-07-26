package utils

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
			continue
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

func GetFileMetaDataRecursive(directory string, extensions map[string]any, keyPrefix string, recurse bool, metaDataCollection map[string]FileMetaData) {
	files, err := os.ReadDir(directory)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		extension := filepath.Ext(file.Name())

		if file.IsDir() && string(file.Name()[0]) != "_" {
			if recurse {
				GetFileMetaDataRecursive(filepath.Join(directory, file.Name()), extensions, file.Name()+"/", true, metaDataCollection)
			}
		} else {
			if _, ok := extensions[extension]; !ok {
				continue
			}

			path := filepath.Join(directory, file.Name())
			name := keyPrefix + file.Name()[0:len(file.Name())-len(extension)]

			metaDataCollection[name] = FileMetaData{Name: name, Path: path, Extension: extension}
		}
	}
}
