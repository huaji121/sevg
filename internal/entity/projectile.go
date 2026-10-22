package entity

import (
	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/world"
)

// Projectile is a thrown item (currently snowballs). It is pulled by gravity
// and disappears on impact with terrain or a mob.
type Projectile struct {
	Item   core.ItemID
	Body   physics.Body
	Age    float32
	Damage float32
}

const (
	projectileGravity = float32(20)
	projectileMaxAge  = float32(8)
)

// SpawnProjectile launches a thrown item.
func (m *Manager) SpawnProjectile(item core.ItemID, pos, vel mathx.Vec3) *Projectile {
	p := &Projectile{
		Item:   item,
		Damage: item.Def().Damage,
		Body:   physics.Body{Pos: pos, Vel: vel, Width: 0.25, Height: 0.25},
	}
	m.Projectiles = append(m.Projectiles, p)
	return p
}

// UpdateProjectiles advances every in-flight projectile.
func (m *Manager) UpdateProjectiles(w *world.World, dt float32) {
	if len(m.Projectiles) == 0 {
		return
	}
	out := m.Projectiles[:0]
	for _, p := range m.Projectiles {
		p.Age += dt
		p.Body.Vel.Y -= projectileGravity * dt
		if p.Body.Vel.Y < -40 {
			p.Body.Vel.Y = -40
		}
		hx, hy, hz := p.Body.Move(w, p.Body.Vel.Mul(dt))
		hit := hx || hy || hz || p.Body.OnGround
		if !hit {
			if mob := m.projectileHitMob(p); mob != nil {
				m.knockMob(mob, p)
				hit = true
			}
		}
		if hit || p.Age > projectileMaxAge || p.Body.Pos.Y < -5 {
			if hit {
				m.ProjectileImpacts = append(m.ProjectileImpacts, p.Body.Pos)
			}
			continue
		}
		out = append(out, p)
	}
	m.Projectiles = out
}

func (m *Manager) projectileHitMob(p *Projectile) *Mob {
	min, max := p.Body.Min(), p.Body.Max()
	for _, mob := range m.Mobs {
		if mob.Dead {
			continue
		}
		bmin, bmax := mob.Bounds()
		if max.X < bmin.X || min.X > bmax.X ||
			max.Y < bmin.Y || min.Y > bmax.Y ||
			max.Z < bmin.Z || min.Z > bmax.Z {
			continue
		}
		return mob
	}
	return nil
}

// knockMob applies the push a projectile delivers, plus damage when the item
// deals any.
func (m *Manager) knockMob(mob *Mob, p *Projectile) {
	dir := p.Body.Vel
	dir.Y = 0
	if dir.LengthSq() < 1e-5 {
		dir = mob.Body.Pos.Sub(p.Body.Pos)
		dir.Y = 0
	}
	if dir.LengthSq() > 1e-5 {
		dir = dir.Normalize()
	}
	mob.Body.Vel.X = dir.X * 4
	mob.Body.Vel.Z = dir.Z * 4
	mob.Body.Vel.Y = 2.2
	mob.HitFlash = 0.2
	if p.Damage > 0 {
		m.DamageMob(mob, p.Damage, p.Body.Pos.Sub(dir))
	}
}
