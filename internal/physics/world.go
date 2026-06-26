package physics

import (
	"math"
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

const (
	DefaultVelocityIterations = 8
	DefaultPositionIterations = 3
	DefaultMaxSubstep         = 1.0 / 60.0
	DefaultMaxSubsteps        = 8
)

type WorldOption func(*World)

type World struct {
	gravity mgl64.Vec3

	nextBodyID BodyID
	bodies     map[BodyID]*Body
	bodyOrder  []BodyID

	VelocityIterations int
	PositionIterations int
	MaxSubstep         float64
	MaxSubsteps        int
}

func NewWorld(options ...WorldOption) *World {
	world := &World{
		gravity:            mgl64.Vec3{0, -9.81, 0},
		nextBodyID:         1,
		bodies:             map[BodyID]*Body{},
		VelocityIterations: DefaultVelocityIterations,
		PositionIterations: DefaultPositionIterations,
		MaxSubstep:         DefaultMaxSubstep,
		MaxSubsteps:        DefaultMaxSubsteps,
	}

	for _, option := range options {
		option(world)
	}

	return world
}

func WithGravity(gravity mgl64.Vec3) WorldOption {
	return func(world *World) {
		world.gravity = gravity
	}
}

func WithSolverIterations(velocityIterations, positionIterations int) WorldOption {
	return func(world *World) {
		if velocityIterations > 0 {
			world.VelocityIterations = velocityIterations
		}
		if positionIterations > 0 {
			world.PositionIterations = positionIterations
		}
	}
}

func WithSubsteps(maxSubstep float64, maxSubsteps int) WorldOption {
	return func(world *World) {
		if maxSubstep > 0 {
			world.MaxSubstep = maxSubstep
		}
		if maxSubsteps > 0 {
			world.MaxSubsteps = maxSubsteps
		}
	}
}

func (w *World) Gravity() mgl64.Vec3 {
	return w.gravity
}

func (w *World) SetGravity(gravity mgl64.Vec3) {
	w.gravity = gravity
}

func (w *World) CreateSphere(radius float64, position mgl64.Vec3, mass float64) (BodyID, error) {
	options := SphereOptions{
		BodyOptions: DefaultBodyOptions(mass),
		Radius:      radius,
	}
	options.Position = position
	return w.CreateSphereWithOptions(options)
}

func (w *World) CreateSphereWithOptions(options SphereOptions) (BodyID, error) {
	id := w.nextBodyID
	body, err := newSphere(id, options)
	if err != nil {
		return 0, err
	}
	w.addBody(body)
	return id, nil
}

func (w *World) CreateCube(size float64, position mgl64.Vec3, mass float64) (BodyID, error) {
	return w.CreateBox(mgl64.Vec3{size, size, size}, position, mass)
}

func (w *World) CreateBox(size, position mgl64.Vec3, mass float64) (BodyID, error) {
	options := CubeOptions{
		BodyOptions: DefaultBodyOptions(mass),
		Size:        size,
	}
	options.Position = position
	return w.CreateCubeWithOptions(options)
}

func (w *World) CreateCubeWithOptions(options CubeOptions) (BodyID, error) {
	id := w.nextBodyID
	body, err := newCube(id, options)
	if err != nil {
		return 0, err
	}
	w.addBody(body)
	return id, nil
}

func (w *World) addBody(body *Body) {
	w.bodies[body.id] = body
	w.bodyOrder = append(w.bodyOrder, body.id)
	w.nextBodyID++
}

func (w *World) RemoveBody(id BodyID) bool {
	if _, ok := w.bodies[id]; !ok {
		return false
	}

	delete(w.bodies, id)
	for i, bodyID := range w.bodyOrder {
		if bodyID == id {
			w.bodyOrder = append(w.bodyOrder[:i], w.bodyOrder[i+1:]...)
			break
		}
	}
	return true
}

func (w *World) Body(id BodyID) (*Body, bool) {
	body, ok := w.bodies[id]
	return body, ok
}

func (w *World) BodyIDs() []BodyID {
	ids := make([]BodyID, 0, len(w.bodyOrder))
	for _, id := range w.bodyOrder {
		if _, ok := w.bodies[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func (w *World) Bodies() []*Body {
	bodies := make([]*Body, 0, len(w.bodyOrder))
	for _, id := range w.bodyOrder {
		if body, ok := w.bodies[id]; ok {
			bodies = append(bodies, body)
		}
	}
	return bodies
}

func (w *World) Position(id BodyID) (mgl64.Vec3, bool) {
	body, ok := w.Body(id)
	if !ok {
		return mgl64.Vec3{}, false
	}
	return body.Position(), true
}

func (w *World) Rotation(id BodyID) (mgl64.Quat, bool) {
	body, ok := w.Body(id)
	if !ok {
		return mgl64.QuatIdent(), false
	}
	return body.Rotation(), true
}

func (w *World) Transform(id BodyID) (Transform, bool) {
	body, ok := w.Body(id)
	if !ok {
		return Transform{}, false
	}
	return body.Transform(), true
}

func (w *World) Step(delta time.Duration) {
	w.Simulate(delta.Seconds())
}

func (w *World) Simulate(dt float64) {
	if dt <= 0 {
		return
	}

	steps := w.substepCount(dt)
	substep := dt / float64(steps)
	for i := 0; i < steps; i++ {
		w.simulateSubstep(substep)
	}
	w.clearForces()
}

func (w *World) substepCount(dt float64) int {
	if w.MaxSubstep <= 0 || dt <= w.MaxSubstep {
		return 1
	}

	steps := int(math.Ceil(dt / w.MaxSubstep))
	if w.MaxSubsteps > 0 && steps > w.MaxSubsteps {
		return w.MaxSubsteps
	}
	return steps
}

func (w *World) simulateSubstep(dt float64) {
	w.integrate(dt)
	contacts := w.detectContacts()

	for i := 0; i < w.VelocityIterations; i++ {
		for j := range contacts {
			resolveContactVelocity(&contacts[j])
		}
	}

	for i := 0; i < w.PositionIterations; i++ {
		contacts = w.detectContacts()
		if len(contacts) == 0 {
			break
		}
		for j := range contacts {
			correctContactPosition(&contacts[j])
		}
	}

	w.stabilizeRestingContacts(contacts)
}

func (w *World) clearForces() {
	for _, id := range w.bodyOrder {
		if body, ok := w.bodies[id]; ok {
			body.ClearForces()
		}
	}
}
