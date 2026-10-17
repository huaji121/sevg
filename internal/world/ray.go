package world

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/mathx"
)

type RayHit struct {
	Hit      bool
	Block    core.BlockID
	Pos      [3]int
	Normal   mathx.Vec3
	Distance float32
}

func targetable(b core.BlockID) bool {
	return b != core.Air && b != core.Water
}

// Raycast walks the voxel grid using the Amanatides-Woo DDA algorithm.
func (w *World) Raycast(origin, dir mathx.Vec3, maxDist float32) RayHit {
	dir = dir.Normalize()
	x, y, z := int(math.Floor(float64(origin.X))), int(math.Floor(float64(origin.Y))), int(math.Floor(float64(origin.Z)))
	stepX, stepY, stepZ := 0, 0, 0
	if dir.X > 0 {
		stepX = 1
	} else if dir.X < 0 {
		stepX = -1
	}
	if dir.Y > 0 {
		stepY = 1
	} else if dir.Y < 0 {
		stepY = -1
	}
	if dir.Z > 0 {
		stepZ = 1
	} else if dir.Z < 0 {
		stepZ = -1
	}
	inf := float32(math.Inf(1))
	tMaxX, tMaxY, tMaxZ := inf, inf, inf
	tDeltaX, tDeltaY, tDeltaZ := inf, inf, inf
	if dir.X != 0 {
		tDeltaX = float32(math.Abs(1 / float64(dir.X)))
		if dir.X > 0 {
			tMaxX = (float32(x+1) - origin.X) / dir.X
		} else {
			tMaxX = (origin.X - float32(x)) / -dir.X
		}
	}
	if dir.Y != 0 {
		tDeltaY = float32(math.Abs(1 / float64(dir.Y)))
		if dir.Y > 0 {
			tMaxY = (float32(y+1) - origin.Y) / dir.Y
		} else {
			tMaxY = (origin.Y - float32(y)) / -dir.Y
		}
	}
	if dir.Z != 0 {
		tDeltaZ = float32(math.Abs(1 / float64(dir.Z)))
		if dir.Z > 0 {
			tMaxZ = (float32(z+1) - origin.Z) / dir.Z
		} else {
			tMaxZ = (origin.Z - float32(z)) / -dir.Z
		}
	}
	normal := mathx.Vec3{}
	var dist float32
	for i := 0; i < 512; i++ {
		b := w.GetBlock(x, y, z)
		if targetable(b) {
			return RayHit{Hit: true, Block: b, Pos: [3]int{x, y, z}, Normal: normal, Distance: dist}
		}
		if tMaxX < tMaxY && tMaxX < tMaxZ {
			x += stepX
			dist = tMaxX
			tMaxX += tDeltaX
			normal = mathx.Vec3{X: float32(-stepX)}
		} else if tMaxY < tMaxZ {
			y += stepY
			dist = tMaxY
			tMaxY += tDeltaY
			normal = mathx.Vec3{Y: float32(-stepY)}
		} else {
			z += stepZ
			dist = tMaxZ
			tMaxZ += tDeltaZ
			normal = mathx.Vec3{Z: float32(-stepZ)}
		}
		if dist > maxDist {
			break
		}
	}
	return RayHit{}
}

// RayAABB returns the nearest positive hit distance for a ray and AABB.
func RayAABB(origin, dir, min, max mathx.Vec3) (float32, bool) {
	tmin := float32(-math.MaxFloat32)
	tmax := float32(math.MaxFloat32)
	axes := [3][2]float32{{origin.X, dir.X}, {origin.Y, dir.Y}, {origin.Z, dir.Z}}
	mins := [3]float32{min.X, min.Y, min.Z}
	maxs := [3]float32{max.X, max.Y, max.Z}
	for i := 0; i < 3; i++ {
		o, d := axes[i][0], axes[i][1]
		if math.Abs(float64(d)) < 1e-7 {
			if o < mins[i] || o > maxs[i] {
				return 0, false
			}
			continue
		}
		inv := 1 / d
		t1 := (mins[i] - o) * inv
		t2 := (maxs[i] - o) * inv
		if t1 > t2 {
			t1, t2 = t2, t1
		}
		if t1 > tmin {
			tmin = t1
		}
		if t2 < tmax {
			tmax = t2
		}
		if tmin > tmax {
			return 0, false
		}
	}
	if tmax < 0 {
		return 0, false
	}
	if tmin < 0 {
		tmin = 0
	}
	return tmin, true
}
