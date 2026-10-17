package player

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/world"
)

type Input struct {
	Forward, Strafe float32
	Jump, Descend   bool
	Sprint          bool
}

type Player struct {
	Body     physics.Body
	Yaw      float32
	Pitch    float32
	Health   float32
	MaxHP    float32
	Flying   bool
	InWater  bool
	Spawning mathx.Vec3

	fallStart    float32
	falling      bool
	hurtCooldown float32
	stepTimer    float32
	Bob          float32
}

func New(spawn mathx.Vec3) *Player {
	p := &Player{Health: 20, MaxHP: 20, Spawning: spawn}
	p.Body = physics.Body{
		Pos: spawn, Width: 0.6, Height: 1.8,
	}
	p.fallStart = spawn.Y
	return p
}

func (p *Player) Eye() mathx.Vec3 {
	e := p.Body.Eye()
	e.Y += p.Bob
	return e
}

func (p *Player) LookDir() mathx.Vec3 {
	cp := float32(math.Cos(float64(p.Pitch)))
	return mathx.Vec3{
		X: float32(math.Sin(float64(p.Yaw))) * cp,
		Y: float32(math.Sin(float64(p.Pitch))),
		Z: -float32(math.Cos(float64(p.Yaw))) * cp,
	}
}

func (p *Player) Rotate(dx, dy, sensitivity float32) {
	p.Yaw += dx * sensitivity
	p.Pitch -= dy * sensitivity
	limit := mathx.Radians(89)
	if p.Pitch > limit {
		p.Pitch = limit
	} else if p.Pitch < -limit {
		p.Pitch = -limit
	}
}

func (p *Player) Damage(amount float32) bool {
	if p.hurtCooldown > 0 || amount <= 0 {
		return false
	}
	p.Health -= amount
	p.hurtCooldown = 0.4
	if p.Health < 0 {
		p.Health = 0
	}
	return true
}

func (p *Player) Heal(amount float32) {
	p.Health += amount
	if p.Health > p.MaxHP {
		p.Health = p.MaxHP
	}
}

func (p *Player) Respawn() {
	p.Body.Pos = p.Spawning
	p.Body.Vel = mathx.Vec3{}
	p.Health = p.MaxHP
	p.FallingReset()
}

func (p *Player) FallingReset() {
	p.fallStart = p.Body.Pos.Y
	p.falling = false
}

func (p *Player) Update(dt float32, w *world.World, in Input) {
	if dt > 0.05 {
		dt = 0.05
	}
	if p.hurtCooldown > 0 {
		p.hurtCooldown -= dt
	}
	feet := w.GetBlock(int(math.Floor(float64(p.Body.Pos.X))), int(math.Floor(float64(p.Body.Pos.Y))), int(math.Floor(float64(p.Body.Pos.Z))))
	p.InWater = feet == core.Water

	yawSin, yawCos := float32(math.Sin(float64(p.Yaw))), float32(math.Cos(float64(p.Yaw)))
	forward := mathx.Vec3{X: yawSin, Z: -yawCos}
	right := mathx.Vec3{X: yawCos, Z: yawSin}
	move := forward.Mul(in.Forward).Add(right.Mul(in.Strafe))
	if move.LengthSq() > 1 {
		move = move.Normalize()
	}
	speed := float32(4.6)
	if in.Sprint {
		speed = 7.0
	}
	if p.Flying {
		speed = 11.0
	}
	if p.InWater && !p.Flying {
		speed *= 0.65
	}
	targetX, targetZ := move.X*speed, move.Z*speed
	accel := float32(1 - math.Exp(float64(-dt*14)))
	if !p.Body.Grounded(w) && !p.Flying {
		accel *= 0.45
	}
	p.Body.Vel.X += (targetX - p.Body.Vel.X) * accel
	p.Body.Vel.Z += (targetZ - p.Body.Vel.Z) * accel

	grounded := p.Body.Grounded(w)
	if p.Flying {
		v := float32(0)
		if in.Jump {
			v += speed
		}
		if in.Descend {
			v -= speed
		}
		p.Body.Vel.Y = v
	} else if p.InWater {
		p.Body.Vel.Y -= 7 * dt
		p.Body.Vel.Y *= 0.94
		if in.Jump {
			p.Body.Vel.Y = 4.2
		}
	} else {
		p.Body.Vel.Y -= 25 * dt
		if p.Body.Vel.Y < -55 {
			p.Body.Vel.Y = -55
		}
		if in.Jump && grounded {
			p.Body.Vel.Y = 8.4
			grounded = false
		}
	}

	delta := p.Body.Vel.Mul(dt)
	p.Body.Move(w, delta)

	if !p.Flying && !p.InWater {
		if !grounded && p.Body.Vel.Y < -0.2 {
			if !p.falling {
				p.falling = true
				p.fallStart = p.Body.Pos.Y
			}
		}
		if p.Body.OnGround || p.Body.Grounded(w) {
			if p.falling {
				fall := p.fallStart - p.Body.Pos.Y
				if fall > 3.6 {
					p.Damage((fall - 3.6) * 2.2)
				}
				p.falling = false
			}
		}
	} else {
		p.falling = false
	}

	// Head bob while walking on the ground.
	hs := float32(math.Sqrt(float64(p.Body.Vel.X*p.Body.Vel.X + p.Body.Vel.Z*p.Body.Vel.Z)))
	if grounded && hs > 0.5 && !p.Flying {
		p.stepTimer += dt * hs * 1.7
		p.Bob = float32(math.Sin(float64(p.stepTimer*2))) * 0.035
	} else {
		p.Bob *= 0.85
	}

	if p.Body.Pos.Y < -10 {
		p.Respawn()
	}
}

func (p *Player) IntersectsBlock(x, y, z int) bool {
	min, max := p.Body.Min(), p.Body.Max()
	return max.X > float32(x) && min.X < float32(x+1) &&
		max.Y > float32(y) && min.Y < float32(y+1) &&
		max.Z > float32(z) && min.Z < float32(z+1)
}
