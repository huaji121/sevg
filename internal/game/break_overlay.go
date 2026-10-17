package game

import (
	"sevg/internal/core"
	"sevg/internal/world"
)

var crackNormals = [6][3]int{
	{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1},
}

func (g *Game) appendBreakOverlay(m *world.MeshData) {
	if g.world == nil || !g.targetValid || !g.breakActive || g.breakProgress <= 0 {
		return
	}
	x, y, z := g.breakPos[0], g.breakPos[1], g.breakPos[2]
	b := g.world.GetBlock(x, y, z)
	if b == core.Air || b == core.Water || b == core.Bedrock {
		return
	}
	stage := int(clampF(g.breakProgress, 0, 0.999) * 10)
	if stage < 0 {
		stage = 0
	}
	if stage > 9 {
		stage = 9
	}
	tile := uint8(core.TileCrack0 + stage)
	u0, v0, u1, v1 := world.TileUV(tile)
	var lightSky, lightBlock uint8
	for _, n := range crackNormals {
		if v := g.world.SkyLightAt(x+n[0], y+n[1], z+n[2]); v > lightSky {
			lightSky = v
		}
		if v := g.world.BlockLightAt(x+n[0], y+n[1], z+n[2]); v > lightBlock {
			lightBlock = v
		}
	}
	const offset = 0.004

	for fi, f := range entityFaces {
		n := crackNormals[fi]
		nb := g.world.GetBlock(x+n[0], y+n[1], z+n[2])
		if nb == b || nb.Opaque() {
			continue
		}
		base := uint32(len(m.Vertices))
		for i := 0; i < 4; i++ {
			p := f.corners[i]
			m.Vertices = append(m.Vertices, world.Vertex{
				X:     float32(x) + p[0] + float32(n[0])*offset,
				Y:     float32(y) + p[1] + float32(n[1])*offset,
				Z:     float32(z) + p[2] + float32(n[2])*offset,
				U:     u0 + (u1-u0)*f.uv[i][0],
				V:     v0 + (v1-v0)*f.uv[i][1],
				Light: 1,
				Sky:   float32(lightSky) / 15, Block: float32(lightBlock) / 15,
			})
		}
		m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}
