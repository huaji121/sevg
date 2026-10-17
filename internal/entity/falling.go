package entity

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/physics"
	"sevg/internal/world"
)

// FallingBlock is a sand/gravel block that has lost its support.
type FallingBlock struct {
	Block core.BlockID
	Body  physics.Body
	Age   float32
}

func fallableBlock(b core.BlockID) bool {
	return b == core.Sand || b == core.Gravel
}

// StartFalling converts unsupported fallable blocks at and above the given
// position into falling entities. It is called after a block edit.
func (m *Manager) StartFalling(w *world.World, x, y, z int) {
	for yy := y; yy < world.ChunkHeight; yy++ {
		b := w.GetBlock(x, yy, z)
		if !fallableBlock(b) {
			break
		}
		if w.GetBlock(x, yy-1, z).Solid() {
			break
		}
		if !w.SetBlock(x, yy, z, core.Air) {
			break
		}
		m.Falling = append(m.Falling, &FallingBlock{
			Block: b,
			Body: physics.Body{
				Pos:    mathx.Vec3{X: float32(x) + 0.5, Y: float32(yy), Z: float32(z) + 0.5},
				Width:  0.98,
				Height: 0.98,
			},
		})
	}
}

func (m *Manager) UpdateFalling(w *world.World, dt float32) {
	if len(m.Falling) == 0 {
		return
	}
	out := m.Falling[:0]
	for _, f := range m.Falling {
		f.Age += dt
		f.Body.Vel.Y -= 18 * dt
		if f.Body.Vel.Y < -32 {
			f.Body.Vel.Y = -32
		}
		f.Body.Move(w, f.Body.Vel.Mul(dt))
		if f.Body.OnGround {
			x := int(math.Floor(float64(f.Body.Pos.X)))
			y := int(math.Floor(float64(f.Body.Pos.Y) + 0.001))
			z := int(math.Floor(float64(f.Body.Pos.Z)))
			cur := w.GetBlock(x, y, z)
			if cur == core.Air || cur == core.Water || cur.Cross() {
				if w.SetBlock(x, y, z, f.Block) {
					continue
				}
			}
			f.Body.Vel.Y = 0
		}
		if f.Body.Pos.Y < 0 {
			continue
		}
		out = append(out, f)
	}
	m.Falling = out
}
