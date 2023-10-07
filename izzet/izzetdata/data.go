package izzetdata

import (
	"encoding/json"
	"io"
	"os"
)

type EntityAsset struct {
	// Name         string `json:"name"`
	Multipart    bool `json:"multipart"`
	SingleEntity bool `json:"single_entity"`
	// Handle       string
}

type Data struct {
	EntityAssets map[string]EntityAsset `json:"entity_assets"`
}

func LoadData(dataFilePath string) *Data {
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
