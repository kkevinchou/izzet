package entities

import (
	"time"

	"github.com/go-gl/mathgl/mgl64"
)

type Particles struct {
	ElapsedTime time.Duration
	Accumulator time.Duration

	ParticleList []Particle
	Position     mgl64.Vec3
}

type Particle struct {
	Position mgl64.Vec3
	Velocity mgl64.Vec3
}

func (p *Particles) SetPosition(position mgl64.Vec3) {
	p.Position = position
}

func (p *Particles) Update(delta time.Duration) {
	p.ElapsedTime = p.ElapsedTime + delta
	p.Accumulator += delta

	for _, particle := range p.ParticleList {
		particle.Position = particle.Position.Add(particle.Velocity.Mul(delta.Seconds()))
	}

	if p.Accumulator > time.Duration(1)*time.Second {
		p.Accumulator -= time.Duration(1) * time.Second
		p.ParticleList = append(
			p.ParticleList,
			Particle{
				Position: p.Position,
				Velocity: mgl64.Vec3{1, 0, 0},
			},
		)
	}
}

func (p *Particles) GetCurrentParticles() []Particle {
	return p.ParticleList
}
