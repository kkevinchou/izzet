package izzetdata

import (
	"encoding/json"
	"io"
	"os"

	"github.com/go-gl/mathgl/mgl32"
)

type EntityAsset struct {
	SingleEntity bool `json:"single_entity"`
	Collider     *Collider
	Static       bool
	Physics      *Physics

	Translation *mgl32.Vec3
	Rotation    *mgl32.Quat
	Scale       *mgl32.Vec3
}

type Physics struct {
	Velocity mgl32.Vec3
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
