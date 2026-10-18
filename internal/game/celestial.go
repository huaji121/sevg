package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/render"
	"sevg/internal/world"
)

// sunDirection returns the world-space direction of the sun for a day time in
// [0,1). Sunrise is at 0.25 and sunset at 0.75, matching the sky colour cycle.
func sunDirection(t float64) mathx.Vec3 {
	angle := (t - 0.25) * 2 * math.Pi
	return mathx.Vec3{
		X: float32(math.Cos(angle)),
		Y: float32(math.Sin(angle)),
		Z: 0.22,
	}.Normalize()
}

func moonDirection(t float64) mathx.Vec3 {
	return sunDirection(t).Mul(-1)
}

// buildSkyMesh places the sun and moon as camera-facing textured quads far
// enough away to sit behind all terrain. They are regular world geometry, so
// terrain depth occludes them naturally, including partial occlusion.
func (g *Game) buildSkyMesh(cam render.Camera) world.MeshData {
	var out world.MeshData
	const dist = float32(420)
	if cam.Fov <= 0 {
		cam.Fov = mathx.Radians(72)
	}
	baseSize := 2 * dist * float32(math.Tan(float64(cam.Fov*0.08)))
	appendBody := func(dir mathx.Vec3, tile uint8, scale float32) {
		if dir.Y < -0.04 {
			return
		}
		right := dir.Cross(mathx.V3(0, 1, 0)).Normalize()
		if right.LengthSq() < 0.01 {
			right = cam.Dir.Cross(mathx.V3(0, 1, 0)).Normalize()
		}
		up := right.Cross(dir).Normalize()
		center := cam.Pos.Add(dir.Mul(dist))
		half := baseSize * scale / 2
		u0, v0, u1, v1 := world.TileUV(tile)
		base := uint32(len(out.Vertices))
		for _, c := range [4][2]float32{{-1, -1}, {1, -1}, {1, 1}, {-1, 1}} {
			p := center.Add(right.Mul(c[0] * half)).Add(up.Mul(c[1] * half))
			u := u0
			if c[0] > 0 {
				u = u1
			}
			v := v0
			if c[1] > 0 {
				v = v1
			}
			out.Vertices = append(out.Vertices, world.Vertex{
				X: p.X, Y: p.Y, Z: p.Z, U: u, V: v, Light: 1, Sky: 1,
			})
		}
		out.Indices = append(out.Indices, base, base+1, base+2, base, base+2, base+3)
	}
	appendBody(moonDirection(g.time), core.TileMoon, 0.82)
	appendBody(sunDirection(g.time), core.TileSun, 1)
	return out
}
