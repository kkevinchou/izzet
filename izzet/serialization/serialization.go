package serialization

import (
	"encoding/json"
	"io"
	"os"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/prefabs"
	"github.com/kkevinchou/kitolib/animation"
)

type App interface {
	GetPrefabByID(id int) *prefabs.Prefab
	ModelLibrary() *modellibrary.ModelLibrary
}

type GameWorld interface {
	Entities() []*entities.Entity
}

type Relation struct {
	Parent int
	Child  int
}

type SerializedWorld struct {
	Entities  []*entities.Entity
	Relations []Relation
}

type Serializer struct {
	app             App
	world           GameWorld
	serializedWorld SerializedWorld
}

func New(app App, world GameWorld) *Serializer {
	return &Serializer{app: app, world: world}
}

func (s *Serializer) WriteOut(filepath string) {
	entities := s.world.Entities()

	serializedWorld := SerializedWorld{
		Entities: entities,
	}

	for _, e := range entities {
		if e.Parent != nil {
			serializedWorld.Relations = append(serializedWorld.Relations, Relation{Parent: e.Parent.ID, Child: e.ID})
		}
	}

	if err := writeToFile(serializedWorld, filepath); err != nil {
		panic(err)
	}
}

func (s *Serializer) ReadIn(filepath string) error {
	serializedWorld, err := readFile(filepath)
	if err != nil {
		return err
	}

	s.serializedWorld = *serializedWorld
	entityMap := map[int]*entities.Entity{}
	for _, e := range s.serializedWorld.Entities {
		entityMap[e.ID] = e

		// set dirty flags
		e.DirtyTransformFlag = true

		// rebuild animation player
		if e.Animation != nil {
			handle := e.Animation.AnimationHandle
			animations, joints := s.app.ModelLibrary().GetAnimations(handle)
			e.Animation.AnimationPlayer = animation.NewAnimationPlayer()
			e.Animation.AnimationPlayer.Initialize(animations, joints[e.Animation.RootJointID])
		}

		// rebuild collider
	}

	// rebuild relations
	for _, relation := range s.serializedWorld.Relations {
		parent := entityMap[relation.Parent]
		child := entityMap[relation.Child]
		if len(parent.Children) == 0 {
			parent.Children = make(map[int]*entities.Entity)
		}
		parent.Children[relation.Child] = child
		child.Parent = parent
	}

	return nil
}

func writeToFile(app SerializedWorld, filepath string) error {
	bytes, err := json.MarshalIndent(app, "", "    ")
	if err != nil {
		return err
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func readFile(filepath string) (*SerializedWorld, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var serializedWorld SerializedWorld
	err = json.Unmarshal(bytes, &serializedWorld)
	if err != nil {
		return nil, err
	}

	return &serializedWorld, err
}

func (s *Serializer) Entities() []*entities.Entity {
	return s.serializedWorld.Entities
}
