package serialization

import (
	"encoding/json"
	"io"
	"os"

	"github.com/kkevinchou/izzet/izzet/entities"
	"github.com/kkevinchou/izzet/izzet/modellibrary"
	"github.com/kkevinchou/izzet/izzet/world"
	"github.com/kkevinchou/kitolib/animation"
)

type App interface {
	ModelLibrary() *modellibrary.ModelLibrary
}

type GameWorld interface {
	Entities() []*entities.Entity
}

type Relation struct {
	Parent int
	Child  int
}

type WorldIR struct {
	Entities  []*entities.Entity
	Relations []Relation
}

type Serializer struct {
	app App
}

func New(app App, world GameWorld) *Serializer {
	return &Serializer{app: app}
}

func (s *Serializer) WriteToFile(world *world.GameWorld, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	defer f.Close()

	err = s.Write(world, f)
	if err != nil {
		return err
	}

	return nil
}

func (s *Serializer) Write(world *world.GameWorld, writer io.Writer) error {
	entities := world.Entities()

	worldIR := WorldIR{
		Entities: entities,
	}

	for _, e := range entities {
		if e.Parent != nil {
			worldIR.Relations = append(worldIR.Relations, Relation{Parent: e.Parent.ID, Child: e.ID})
		}
	}

	bytes, err := json.MarshalIndent(worldIR, "", "    ")
	if err != nil {
		return err
	}

	_, err = writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (s *Serializer) ReadFromFile(filepath string) (*world.GameWorld, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	gameWorld, err := s.Read(f)
	if err != nil {
		return nil, err
	}

	return gameWorld, err
}

func (s *Serializer) Read(reader io.Reader) (*world.GameWorld, error) {
	bytes, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	var worldIR WorldIR
	err = json.Unmarshal(bytes, &worldIR)
	if err != nil {
		return nil, err
	}

	entityMap := map[int]*entities.Entity{}
	for _, e := range worldIR.Entities {
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
	for _, relation := range worldIR.Relations {
		parent := entityMap[relation.Parent]
		child := entityMap[relation.Child]
		if len(parent.Children) == 0 {
			parent.Children = make(map[int]*entities.Entity)
		}
		parent.Children[relation.Child] = child
		child.Parent = parent
	}

	return world.New(entityMap), nil
}
