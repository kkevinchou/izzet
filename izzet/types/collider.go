package types

type ColliderType string

const (
	ColliderTypeNone ColliderType = "NONE"
	ColliderTypeMesh ColliderType = "MESH"
)

var ColliderTypes []ColliderType = []ColliderType{
	ColliderTypeNone,
	ColliderTypeMesh,
}

type ColliderGroup string

const (
	ColliderGroupNone    ColliderGroup = "NONE"
	ColliderGroupTerrain ColliderGroup = "TERRAIN"
	ColliderGroupPlayer  ColliderGroup = "PLAYER"
)

var ColliderGroups []ColliderGroup = []ColliderGroup{
	ColliderGroupNone,
	ColliderGroupTerrain,
	ColliderGroupPlayer,
}

var ColliderGroupMap map[ColliderGroup]ColliderGroupFlag = map[ColliderGroup]ColliderGroupFlag{
	ColliderGroupTerrain: ColliderGroupFlagTerrain,
	ColliderGroupPlayer:  ColliderGroupFlagPlayer,
}

var ColliderFlagToGroupName map[ColliderGroupFlag]ColliderGroup

func init() {
	ColliderFlagToGroupName = map[ColliderGroupFlag]ColliderGroup{}
	for k, v := range ColliderGroupMap {
		ColliderFlagToGroupName[v] = k
	}
}

func ConvertGroupToFlag(group ColliderGroup) ColliderGroupFlag {
	return ColliderGroupMap[group]
}

type ColliderGroupFlag uint64

const (
	ColliderGroupFlagTerrain ColliderGroupFlag = 1 << 0
	ColliderGroupFlagPlayer  ColliderGroupFlag = 2 << 0
)
