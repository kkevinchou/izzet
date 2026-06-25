package physics

import (
	"errors"
	"math"

	"github.com/go-gl/mathgl/mgl64"
)

const (
	DefaultRestitution = 0.2
	DefaultFriction    = 0.5
)

var (
	ErrInvalidRadius      = errors.New("physics: sphere radius must be finite and greater than zero")
	ErrInvalidSize        = errors.New("physics: cube size components must be finite and greater than zero")
	ErrInvalidMass        = errors.New("physics: mass must be finite and non-negative")
	ErrInvalidBodyOptions = errors.New("physics: body options must contain only finite values")
)

type BodyID int

type ShapeType int

const (
	ShapeSphere ShapeType = iota + 1
	ShapeCube
)

func (s ShapeType) String() string {
	switch s {
	case ShapeSphere:
		return "sphere"
	case ShapeCube:
		return "cube"
	default:
		return "unknown"
	}
}

type Transform struct {
	Position mgl64.Vec3
	Rotation mgl64.Quat
}

type BodyOptions struct {
	Position        mgl64.Vec3
	Rotation        mgl64.Quat
	LinearVelocity  mgl64.Vec3
	AngularVelocity mgl64.Vec3

	// A body is immovable when Static is true or Mass is zero.
	Mass   float64
	Static bool

	Restitution    float64
	Friction       float64
	LinearDamping  float64
	AngularDamping float64
}

type SphereOptions struct {
	BodyOptions
	Radius float64
}

type CubeOptions struct {
	BodyOptions

	// Size is the full cube/cuboid extent along each local axis.
	Size mgl64.Vec3
}

type Body struct {
	id    BodyID
	shape ShapeType

	radius      float64
	halfExtents mgl64.Vec3

	position        mgl64.Vec3
	rotation        mgl64.Quat
	linearVelocity  mgl64.Vec3
	angularVelocity mgl64.Vec3
	force           mgl64.Vec3
	torque          mgl64.Vec3

	inverseMass         float64
	inverseInertiaLocal mgl64.Vec3

	restitution    float64
	friction       float64
	linearDamping  float64
	angularDamping float64
}

func DefaultBodyOptions(mass float64) BodyOptions {
	return BodyOptions{
		Rotation:    mgl64.QuatIdent(),
		Mass:        mass,
		Restitution: DefaultRestitution,
		Friction:    DefaultFriction,
	}
}

func newSphere(id BodyID, options SphereOptions) (*Body, error) {
	if !finiteFloat(options.Radius) || options.Radius <= 0 {
		return nil, ErrInvalidRadius
	}
	return newBody(id, ShapeSphere, options.Radius, mgl64.Vec3{}, options.BodyOptions)
}

func newCube(id BodyID, options CubeOptions) (*Body, error) {
	if !finiteVec3(options.Size) || options.Size.X() <= 0 || options.Size.Y() <= 0 || options.Size.Z() <= 0 {
		return nil, ErrInvalidSize
	}
	return newBody(id, ShapeCube, 0, options.Size.Mul(0.5), options.BodyOptions)
}

func newBody(id BodyID, shape ShapeType, radius float64, halfExtents mgl64.Vec3, options BodyOptions) (*Body, error) {
	if !finiteFloat(options.Mass) || options.Mass < 0 {
		return nil, ErrInvalidMass
	}
	if !finiteBodyOptions(options) {
		return nil, ErrInvalidBodyOptions
	}

	inverseMass := 0.0
	if !options.Static && options.Mass > 0 {
		inverseMass = 1 / options.Mass
	}

	body := &Body{
		id:                  id,
		shape:               shape,
		radius:              radius,
		halfExtents:         halfExtents,
		position:            options.Position,
		rotation:            normalizeQuat(options.Rotation),
		linearVelocity:      options.LinearVelocity,
		angularVelocity:     options.AngularVelocity,
		inverseMass:         inverseMass,
		restitution:         clamp01(options.Restitution),
		friction:            math.Max(0, options.Friction),
		linearDamping:       clamp01(options.LinearDamping),
		angularDamping:      clamp01(options.AngularDamping),
		inverseInertiaLocal: mgl64.Vec3{},
	}
	body.recomputeInertia(options.Mass)
	return body, nil
}

func finiteBodyOptions(options BodyOptions) bool {
	return finiteVec3(options.Position) &&
		finiteQuat(options.Rotation) &&
		finiteVec3(options.LinearVelocity) &&
		finiteVec3(options.AngularVelocity) &&
		finiteFloat(options.Restitution) &&
		finiteFloat(options.Friction) &&
		finiteFloat(options.LinearDamping) &&
		finiteFloat(options.AngularDamping)
}

func (b *Body) recomputeInertia(mass float64) {
	if b.inverseMass == 0 {
		b.inverseInertiaLocal = mgl64.Vec3{}
		return
	}

	switch b.shape {
	case ShapeSphere:
		inertia := (2.0 / 5.0) * mass * b.radius * b.radius
		b.inverseInertiaLocal = mgl64.Vec3{1 / inertia, 1 / inertia, 1 / inertia}
	case ShapeCube:
		size := b.halfExtents.Mul(2)
		x2 := size.X() * size.X()
		y2 := size.Y() * size.Y()
		z2 := size.Z() * size.Z()
		ix := (1.0 / 12.0) * mass * (y2 + z2)
		iy := (1.0 / 12.0) * mass * (x2 + z2)
		iz := (1.0 / 12.0) * mass * (x2 + y2)
		b.inverseInertiaLocal = mgl64.Vec3{inverseOrZero(ix), inverseOrZero(iy), inverseOrZero(iz)}
	}
}

func inverseOrZero(value float64) float64 {
	if math.Abs(value) <= epsilon {
		return 0
	}
	return 1 / value
}

func (b *Body) ID() BodyID {
	return b.id
}

func (b *Body) ShapeType() ShapeType {
	return b.shape
}

func (b *Body) Static() bool {
	return b.inverseMass == 0
}

func (b *Body) Mass() float64 {
	if b.inverseMass == 0 {
		return math.Inf(1)
	}
	return 1 / b.inverseMass
}

func (b *Body) InverseMass() float64 {
	return b.inverseMass
}

func (b *Body) Radius() (float64, bool) {
	if b.shape != ShapeSphere {
		return 0, false
	}
	return b.radius, true
}

func (b *Body) Size() (mgl64.Vec3, bool) {
	if b.shape != ShapeCube {
		return mgl64.Vec3{}, false
	}
	return b.halfExtents.Mul(2), true
}

func (b *Body) Transform() Transform {
	return Transform{Position: b.position, Rotation: b.rotation}
}

func (b *Body) Position() mgl64.Vec3 {
	return b.position
}

func (b *Body) SetPosition(position mgl64.Vec3) {
	if !finiteVec3(position) {
		return
	}
	b.position = position
}

func (b *Body) Rotation() mgl64.Quat {
	return b.rotation
}

func (b *Body) SetRotation(rotation mgl64.Quat) {
	if !finiteQuat(rotation) {
		return
	}
	b.rotation = normalizeQuat(rotation)
}

func (b *Body) SetTransform(transform Transform) {
	if !finiteVec3(transform.Position) || !finiteQuat(transform.Rotation) {
		return
	}
	b.position = transform.Position
	b.rotation = normalizeQuat(transform.Rotation)
}

func (b *Body) LinearVelocity() mgl64.Vec3 {
	return b.linearVelocity
}

func (b *Body) SetLinearVelocity(velocity mgl64.Vec3) {
	if !finiteVec3(velocity) {
		return
	}
	b.linearVelocity = velocity
}

func (b *Body) AngularVelocity() mgl64.Vec3 {
	return b.angularVelocity
}

func (b *Body) SetAngularVelocity(velocity mgl64.Vec3) {
	if !finiteVec3(velocity) {
		return
	}
	b.angularVelocity = velocity
}

func (b *Body) ApplyForce(force mgl64.Vec3) {
	if b.Static() || !finiteVec3(force) {
		return
	}
	b.force = b.force.Add(force)
}

func (b *Body) ApplyForceAtPoint(force, worldPoint mgl64.Vec3) {
	if b.Static() || !finiteVec3(force) || !finiteVec3(worldPoint) {
		return
	}
	b.force = b.force.Add(force)
	b.torque = b.torque.Add(worldPoint.Sub(b.position).Cross(force))
}

func (b *Body) ApplyImpulse(impulse, worldPoint mgl64.Vec3) {
	if !finiteVec3(impulse) || !finiteVec3(worldPoint) {
		return
	}
	b.applyImpulse(impulse, worldPoint)
}

func (b *Body) ClearForces() {
	b.force = mgl64.Vec3{}
	b.torque = mgl64.Vec3{}
}

func (b *Body) BoundingRadius() float64 {
	switch b.shape {
	case ShapeSphere:
		return b.radius
	case ShapeCube:
		return b.halfExtents.Len()
	default:
		return 0
	}
}

func (b *Body) AABB() (mgl64.Vec3, mgl64.Vec3) {
	switch b.shape {
	case ShapeSphere:
		extent := mgl64.Vec3{b.radius, b.radius, b.radius}
		return b.position.Sub(extent), b.position.Add(extent)
	case ShapeCube:
		axes := cubeAxes(b)
		worldExtent := componentAbs(axes[0]).Mul(b.halfExtents.X()).
			Add(componentAbs(axes[1]).Mul(b.halfExtents.Y())).
			Add(componentAbs(axes[2]).Mul(b.halfExtents.Z()))
		return b.position.Sub(worldExtent), b.position.Add(worldExtent)
	default:
		return b.position, b.position
	}
}

func (b *Body) applyImpulse(impulse, worldPoint mgl64.Vec3) {
	if b.Static() {
		return
	}

	b.linearVelocity = b.linearVelocity.Add(impulse.Mul(b.inverseMass))
	r := worldPoint.Sub(b.position)
	b.angularVelocity = b.angularVelocity.Add(b.applyInverseInertia(r.Cross(impulse)))
}

func (b *Body) applyInverseInertia(worldVector mgl64.Vec3) mgl64.Vec3 {
	if b.Static() {
		return mgl64.Vec3{}
	}

	localVector := b.rotation.Conjugate().Rotate(worldVector)
	localResult := componentMul(localVector, b.inverseInertiaLocal)
	return b.rotation.Rotate(localResult)
}
