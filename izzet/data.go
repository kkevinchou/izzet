package izzet

import (
	"encoding/json"
	"io"
	"os"
)

type EntityAsset struct {
	Name      string `json:"name"`
	Multipart bool   `json:"multipart"`
}

type Data struct {
	EntityAssets []EntityAsset `json:"entity_assets"`
}

func loadData(dataFilePath string) *Data {
	var data Data
	dataFile, err := os.Open(dataFilePath)
	if err != nil {
		panic(err.Error())
	} else {
		dataBytes, err := io.ReadAll(dataFile)
		if err != nil {
			panic(err)
		}

		if err = dataFile.Close(); err != nil {
			panic(err)
		}

		err = json.Unmarshal(dataBytes, &data)
		if err != nil {
			panic(err)
		}
	}
	return &data
}
