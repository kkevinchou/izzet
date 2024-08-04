package entities

import (
	"time"

	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
)

const (
	particleSpeed      = 150
	particleSpawnTimer = 2 * time.Millisecond
)

type ParticleGenerator struct {
	ElapsedTime time.Duration
	Accumulator time.Duration

	ParticleList []Particle
	Position     mgl32.Vec3

	MaxParticles int
	InsertCursor int
}

type Particle struct {
	Position mgl32.Vec3
	Velocity mgl32.Vec3
	Active   bool
}

func NewParticleGenerator(maxParticles int) *ParticleGenerator {
	return &ParticleGenerator{
		MaxParticles: maxParticles,
		ParticleList: make([]Particle, maxParticles),
	}
}

func (p *ParticleGenerator) SetPosition(position mgl32.Vec3) {
	p.Position = position
}

func (p *ParticleGenerator) Update(delta time.Duration) {
	p.ElapsedTime = p.ElapsedTime + delta
	p.Accumulator += delta

	for i, particle := range p.ParticleList {
		if !particle.Active {
			continue
		}
		p.ParticleList[i].Position = particle.Position.Add(particle.Velocity.Mul(float32(delta.Seconds())))
	}

	for p.Accumulator > particleSpawnTimer {
		p.Accumulator -= particleSpawnTimer

		x := rand.Float32()*2 - 1
		y := rand.Float32()*2 - 1
		z := rand.Float32()*2 - 1
		p.ParticleList[p.InsertCursor%p.MaxParticles] = Particle{
			Position: p.Position,
			Velocity: mgl32.Vec3{x, y, z}.Mul(particleSpeed),
			Active:   true,
		}
		p.InsertCursor++
	}
}

func (p *ParticleGenerator) GetActiveParticles() []Particle {
	activeParticles := []Particle{}
	for _, particle := range p.ParticleList {
		if particle.Active {
			activeParticles = append(activeParticles, particle)
		}
	}
	return activeParticles
}
