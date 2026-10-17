package game

import (
	"math"
	"math/rand"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

type breakParticle struct {
	pos, vel mathx.Vec3
	age      float32
	life     float32
	size     float32
	block    core.BlockID
	sky      float32
	light    float32
}

const maxBreakParticles = 320

// spawnBlockParticles creates small whole-block fragments. The source block is
// not modified; this is purely a visual effect.
func (g *Game) spawnBlockParticles(block core.BlockID, pos [3]int, normal mathx.Vec3, count int) {
	if g.world == nil || block == core.Air || block == core.Water || count <= 0 {
		return
	}
	if normal.Length() < 0.01 {
		normal = mathx.V3(0, 1, 0)
	}
	normal = normal.Normalize()
	for i := 0; i < count; i++ {
		base := mathx.V3(float32(pos[0])+0.5, float32(pos[1])+0.5, float32(pos[2])+0.5)
		base = base.Add(normal.Mul(0.51))
		// Scatter across the face tangent plane.
		tangent := mathx.V3(1, 0, 0)
		if math.Abs(float64(normal.X)) > 0.5 {
			tangent = mathx.V3(0, 1, 0)
		}
		bitangent := normal.Cross(tangent).Normalize()
		base = base.Add(tangent.Mul((rand.Float32() - 0.5) * 0.7))
		base = base.Add(bitangent.Mul((rand.Float32() - 0.5) * 0.7))
		vel := normal.Mul(1.0 + rand.Float32()*2.6)
		vel = vel.Add(mathx.V3(rand.Float32()-0.5, rand.Float32()*1.8, rand.Float32()-0.5).Mul(1.4))
		sky := float32(g.world.SkyLightAt(int(math.Floor(float64(base.X))), int(math.Floor(float64(base.Y))), int(math.Floor(float64(base.Z))))) / 15
		blk := float32(g.world.BlockLightAt(int(math.Floor(float64(base.X))), int(math.Floor(float64(base.Y))), int(math.Floor(float64(base.Z))))) / 15
		g.particles = append(g.particles, breakParticle{
			pos:   base,
			vel:   vel,
			life:  0.35 + rand.Float32()*0.55,
			size:  0.07 + rand.Float32()*0.08,
			block: block,
			sky:   sky,
			light: blk,
		})
	}
	if len(g.particles) > maxBreakParticles {
		g.particles = append([]breakParticle(nil), g.particles[len(g.particles)-maxBreakParticles:]...)
	}
}

func (g *Game) updateParticles(dt float32) {
	if len(g.particles) == 0 {
		return
	}
	out := g.particles[:0]
	for _, p := range g.particles {
		p.age += dt
		if p.age >= p.life {
			continue
		}
		p.vel.Y -= 18 * dt
		oldY := p.pos.Y
		p.pos = p.pos.Add(p.vel.Mul(dt))
		if g.world != nil {
			x := int(math.Floor(float64(p.pos.X)))
			y := int(math.Floor(float64(p.pos.Y)))
			z := int(math.Floor(float64(p.pos.Z)))
			if g.world.GetBlock(x, y, z).Solid() {
				p.pos.Y = oldY
				p.vel.Y = -p.vel.Y * 0.22
				p.vel.X *= 0.55
				p.vel.Z *= 0.55
			}
		}
		out = append(out, p)
	}
	g.particles = out
}

// appendParticleMeshes renders each fragment as a miniature version of the
// source block. Transparent blocks go into the blended transparent pass.
func (g *Game) appendParticleMeshes(opaque, transparent *world.MeshData) {
	for _, p := range g.particles {
		target := opaque
		if p.block == core.Water || p.block == core.Glass || p.block == core.Ice {
			target = transparent
		}
		shr := 1 - 0.45*(p.age/p.life)
		half := p.size * shr * 0.5
		min := p.pos.Sub(mathx.V3(half, half, half))
		max := p.pos.Add(mathx.V3(half, half, half))
		if p.block.Cross() {
			appendParticleCross(target, p, min, max)
		} else {
			appendParticleCube(target, p, min, max)
		}
	}
}

func appendParticleCube(m *world.MeshData, p breakParticle, min, max mathx.Vec3) {
	for fi, f := range entityFaces {
		u0, v0, u1, v1 := world.TileUV(p.block.Tile(fi))
		base := uint32(len(m.Vertices))
		for i := 0; i < 4; i++ {
			c := f.corners[i]
			m.Vertices = append(m.Vertices, world.Vertex{
				X:     min.X + (max.X-min.X)*c[0],
				Y:     min.Y + (max.Y-min.Y)*c[1],
				Z:     min.Z + (max.Z-min.Z)*c[2],
				U:     u0 + (u1-u0)*f.uv[i][0],
				V:     v0 + (v1-v0)*f.uv[i][1],
				Light: f.light, Sky: p.sky, Block: p.light,
			})
		}
		m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}

func appendParticleCross(m *world.MeshData, p breakParticle, min, max mathx.Vec3) {
	u0, v0, u1, v1 := world.TileUV(p.block.Tile(2))
	quads := [2][4]mathx.Vec3{
		{
			{X: min.X, Y: min.Y, Z: min.Z},
			{X: max.X, Y: min.Y, Z: max.Z},
			{X: max.X, Y: max.Y, Z: max.Z},
			{X: min.X, Y: max.Y, Z: min.Z},
		},
		{
			{X: min.X, Y: min.Y, Z: max.Z},
			{X: max.X, Y: min.Y, Z: min.Z},
			{X: max.X, Y: max.Y, Z: min.Z},
			{X: min.X, Y: max.Y, Z: max.Z},
		},
	}
	uvs := [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	for _, q := range quads {
		base := uint32(len(m.Vertices))
		for i := 0; i < 4; i++ {
			m.Vertices = append(m.Vertices, world.Vertex{
				X: q[i].X, Y: q[i].Y, Z: q[i].Z,
				U: u0 + (u1-u0)*uvs[i][0], V: v0 + (v1-v0)*uvs[i][1],
				Light: 0.92, Sky: p.sky, Block: p.light,
			})
		}
		m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}
