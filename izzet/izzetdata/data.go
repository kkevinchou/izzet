package izzetdata

import (
	"encoding/json"
	"io"
	"os"

	"github.com/go-gl/mathgl/mgl64"
)

type EntityAsset struct {
	SingleEntity bool      `json:"single_entity"`
	Collider     *Collider `json:"collider"`
	Physics      *Physics  `json:"physics"`

	Translation *mgl64.Vec3
	Rotation    *mgl64.Quat
	Scale       *mgl64.Vec3
}

type Physics struct {
	Velocity mgl64.Vec3
}

type Collider struct {
	ColliderGroup   string
	TriMeshCollider bool `json:"trimeshcollider"`
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
