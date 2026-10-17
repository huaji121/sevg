// Package physics implements AABB collision against the voxel world.
package physics

import (
	"math"

	"sevg/internal/mathx"
	"sevg/internal/world"
)

type Body struct {
	Pos      mathx.Vec3 // feet centre
	Vel      mathx.Vec3
	Width    float32
	Height   float32
	OnGround bool
}

func (b *Body) Min() mathx.Vec3 {
	return mathx.Vec3{X: b.Pos.X - b.Width/2, Y: b.Pos.Y, Z: b.Pos.Z - b.Width/2}
}
func (b *Body) Max() mathx.Vec3 {
	return mathx.Vec3{X: b.Pos.X + b.Width/2, Y: b.Pos.Y + b.Height, Z: b.Pos.Z + b.Width/2}
}
func (b *Body) Eye() mathx.Vec3 {
	return mathx.Vec3{X: b.Pos.X, Y: b.Pos.Y + b.Height*0.9, Z: b.Pos.Z}
}

func Collides(w *world.World, pos mathx.Vec3, width, height float32) bool {
	minX := int(math.Floor(float64(pos.X - width/2)))
	maxX := int(math.Floor(float64(pos.X + width/2 - 0.0001)))
	minY := int(math.Floor(float64(pos.Y)))
	maxY := int(math.Floor(float64(pos.Y + height - 0.0001)))
	minZ := int(math.Floor(float64(pos.Z - width/2)))
	maxZ := int(math.Floor(float64(pos.Z + width/2 - 0.0001)))
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			for z := minZ; z <= maxZ; z++ {
				if w.GetBlock(x, y, z).Solid() {
					return true
				}
			}
		}
	}
	return false
}

// Move applies delta one axis at a time and resolves collisions. It returns
// which axes were blocked.
func (b *Body) Move(w *world.World, delta mathx.Vec3) (hitX, hitY, hitZ bool) {
	orig := b.Pos
	// Y
	b.Pos.Y += delta.Y
	if Collides(w, b.Pos, b.Width, b.Height) {
		b.Pos.Y = orig.Y
		if delta.Y < 0 {
			b.OnGround = true
		}
		b.Vel.Y = 0
		hitY = true
	} else if delta.Y < 0 {
		b.OnGround = false
	}
	// X
	prevX := b.Pos.X
	b.Pos.X += delta.X
	if Collides(w, b.Pos, b.Width, b.Height) {
		b.Pos.X = prevX
		b.Vel.X = 0
		hitX = true
	}
	// Z
	prevZ := b.Pos.Z
	b.Pos.Z += delta.Z
	if Collides(w, b.Pos, b.Width, b.Height) {
		b.Pos.Z = prevZ
		b.Vel.Z = 0
		hitZ = true
	}
	return
}

func (b *Body) Grounded(w *world.World) bool {
	p := b.Pos
	p.Y -= 0.04
	return Collides(w, p, b.Width, b.Height)
}

func (b *Body) StepUp(w *world.World, delta mathx.Vec3) {
	// Reserved for future auto-jump / stair support.
	_, _, _ = w, delta, b
}
