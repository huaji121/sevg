package entity

import (
	"math"

	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/world"
)

// pushBodies separates two overlapping bodies horizontally along the axis of
// least penetration, moving each one half the overlap. Movement goes through
// physics.Body.Move so the push still respects terrain collisions.
func pushBodies(w *world.World, a, b *physics.Body) {
	aMin, aMax := a.Min(), a.Max()
	bMin, bMax := b.Min(), b.Max()
	if aMax.Y <= bMin.Y+0.001 || aMin.Y >= bMax.Y-0.001 {
		return
	}
	dx := (aMin.X+aMax.X)*0.5 - (bMin.X+bMax.X)*0.5
	dz := (aMin.Z+aMax.Z)*0.5 - (bMin.Z+bMax.Z)*0.5
	overlapX := (aMax.X-aMin.X+bMax.X-bMin.X)*0.5 - float32(math.Abs(float64(dx)))
	overlapZ := (aMax.Z-aMin.Z+bMax.Z-bMin.Z)*0.5 - float32(math.Abs(float64(dz)))
	if overlapX <= 0 || overlapZ <= 0 {
		return
	}
	var pushA, pushB mathx.Vec3
	if overlapX < overlapZ {
		if dx < 0 {
			pushA.X, pushB.X = -overlapX, overlapX
		} else {
			pushA.X, pushB.X = overlapX, -overlapX
		}
	} else {
		if dz < 0 {
			pushA.Z, pushB.Z = -overlapZ, overlapZ
		} else {
			pushA.Z, pushB.Z = overlapZ, -overlapZ
		}
	}
	a.Move(w, pushA.Mul(0.5))
	b.Move(w, pushB.Mul(0.5))
}

// separateMobs resolves overlaps between every pair of nearby mobs so they
// cannot stack on top of one another.
func (m *Manager) separateMobs(w *world.World) {
	for i := 0; i < len(m.Mobs); i++ {
		a := m.Mobs[i]
		if a.Dead {
			continue
		}
		for j := i + 1; j < len(m.Mobs); j++ {
			b := m.Mobs[j]
			if b.Dead {
				continue
			}
			pushBodies(w, &a.Body, &b.Body)
		}
	}
}

// PushPlayer separates the player from every mob. The game calls this after
// both the player and the mobs have moved for the frame.
func (m *Manager) PushPlayer(w *world.World, player *physics.Body) {
	if player == nil {
		return
	}
	for _, mob := range m.Mobs {
		if mob.Dead {
			continue
		}
		pushBodies(w, player, &mob.Body)
	}
}
