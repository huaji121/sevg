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
	// tile overrides the block tile when non-zero (used by item/food crumbs
	// that have no source block). cross renders the particle as a crossed pair
	// of quads instead of a cube.
	tile  uint8
	cross bool
	sky   float32
	light float32
}

const maxBreakParticles = 320

// addParticle appends a particle and keeps the pool bounded.
func (g *Game) addParticle(p breakParticle) {
	g.particles = append(g.particles, p)
	if len(g.particles) > maxBreakParticles {
		g.particles = g.particles[len(g.particles)-maxBreakParticles:]
	}
}

func particleTile(p breakParticle, face int) uint8 {
	if p.tile != 0 {
		return p.tile
	}
	return p.block.Tile(face)
}

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
		g.addParticle(breakParticle{
			pos:   base,
			vel:   vel,
			life:  0.35 + rand.Float32()*0.55,
			size:  0.07 + rand.Float32()*0.08,
			block: block,
			cross: block.Cross(),
			sky:   sky,
			light: blk,
		})
	}
}

// spawnEatParticles spits small coloured crumbs from near the mouth while a
// food item is being used, echoing the block-break debris.
func (g *Game) spawnEatParticles(item core.ItemID, count int) {
	if g.world == nil || g.player == nil || count <= 0 {
		return
	}
	def := item.Def()
	tile := tileForColor([4]float32{def.Color[0], def.Color[1], def.Color[2], 1})
	eye := g.player.Eye()
	dir := g.player.LookDir()
	base := eye.Add(dir.Mul(0.30)).Add(mathx.V3(0, -0.06, 0))
	for i := 0; i < count; i++ {
		pos := base.Add(mathx.V3(rand.Float32()-0.5, rand.Float32()-0.5, rand.Float32()-0.5).Mul(0.14))
		vel := dir.Mul(0.5 + rand.Float32()*0.9)
		vel = vel.Add(mathx.V3((rand.Float32()-0.5)*0.8, -0.5+rand.Float32()*0.5, (rand.Float32()-0.5)*0.8))
		x := int(math.Floor(float64(pos.X)))
		y := int(math.Floor(float64(pos.Y)))
		z := int(math.Floor(float64(pos.Z)))
		sky := float32(g.world.SkyLightAt(x, y, z)) / 15
		blk := float32(g.world.BlockLightAt(x, y, z)) / 15
		g.addParticle(breakParticle{
			pos:   pos,
			vel:   vel,
			life:  0.35 + rand.Float32()*0.35,
			size:  0.045 + rand.Float32()*0.045,
			tile:  tile,
			sky:   sky,
			light: blk,
		})
	}
}

// spawnHitParticles bursts fragments of a tile at an impact point.
func (g *Game) spawnHitParticles(pos mathx.Vec3, tile uint8, count int) {
	if g.world == nil || count <= 0 {
		return
	}
	x := int(math.Floor(float64(pos.X)))
	y := int(math.Floor(float64(pos.Y)))
	z := int(math.Floor(float64(pos.Z)))
	sky := float32(g.world.SkyLightAt(x, y, z)) / 15
	blk := float32(g.world.BlockLightAt(x, y, z)) / 15
	for i := 0; i < count; i++ {
		vel := mathx.V3((rand.Float32()-0.5)*3, rand.Float32()*2.2, (rand.Float32()-0.5)*3)
		g.addParticle(breakParticle{
			pos:   pos,
			vel:   vel,
			life:  0.3 + rand.Float32()*0.4,
			size:  0.05 + rand.Float32()*0.05,
			tile:  tile,
			sky:   sky,
			light: blk,
		})
	}
}

// flushProjectileImpacts turns recorded projectile hits into snow particles.
func (g *Game) flushProjectileImpacts() {
	if g.mobs == nil || len(g.mobs.ProjectileImpacts) == 0 {
		return
	}
	for _, pos := range g.mobs.ProjectileImpacts {
		g.spawnHitParticles(pos, core.TileSnow, 8)
	}
	g.mobs.ProjectileImpacts = g.mobs.ProjectileImpacts[:0]
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
		if p.cross {
			appendParticleCross(target, p, min, max)
		} else {
			appendParticleCube(target, p, min, max)
		}
	}
}

func appendParticleCube(m *world.MeshData, p breakParticle, min, max mathx.Vec3) {
	for fi, f := range entityFaces {
		u0, v0, u1, v1 := world.TileUV(particleTile(p, fi))
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
	u0, v0, u1, v1 := world.TileUV(particleTile(p, 2))
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
