package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

// shadowStrength is the opacity of a shadow patch resting directly on the
// ground.
const shadowStrength = float32(0.5)

// shadowFadeHeight is how far an entity may rise above a surface before the
// shadow patch on that surface has faded out completely (Minecraft uses ~2).
const shadowFadeHeight = float32(2.0)

// shadowScale widens the shadow footprint relative to the entity hitbox so it
// reads like Minecraft's soft ellipse and spans the block cells below.
const shadowScale = float32(1.3)

// groundTop returns the top surface Y of the highest solid block in the
// (bx, bz) column at or below startY. Non-solid blocks such as plants and
// water are ignored so shadows land on real ground.
func groundTop(w *world.World, bx, bz, startY int) (float32, bool) {
	for by := startY; by >= startY-8; by-- {
		if w.GetBlock(bx, by, bz).Solid() {
			return float32(by + 1), true
		}
	}
	return 0, false
}

// appendShadow projects a Minecraft-style shadow onto the terrain beneath an
// entity. The footprint is split into block cells; every cell samples its own
// ground surface, so the shadow drapes down over steps and ledges instead of
// hovering on a single plane, and each patch fades with its own height.
func appendShadow(dst *world.MeshData, w *world.World, x, y, z, size float32) {
	if w == nil || size <= 0 {
		return
	}
	diameter := size * shadowScale
	half := diameter * 0.5
	minX, maxX := x-half, x+half
	minZ, maxZ := z-half, z+half
	startY := int(math.Floor(float64(y)))
	u0, v0, u1, v1 := world.TileUV(core.TileShadow)

	bx0 := int(math.Floor(float64(minX)))
	bx1 := int(math.Floor(float64(maxX)))
	bz0 := int(math.Floor(float64(minZ)))
	bz1 := int(math.Floor(float64(maxZ)))
	for bx := bx0; bx <= bx1; bx++ {
		for bz := bz0; bz <= bz1; bz++ {
			// Clip this block cell to the shadow footprint.
			cx0 := max(minX, float32(bx))
			cx1 := min(maxX, float32(bx+1))
			cz0 := max(minZ, float32(bz))
			cz1 := min(maxZ, float32(bz+1))
			if cx1 <= cx0 || cz1 <= cz0 {
				continue
			}
			groundY, ok := groundTop(w, bx, bz, startY)
			if !ok {
				continue
			}
			height := y - groundY
			if height < 0 {
				height = 0
			}
			fade := clampF(1-height/shadowFadeHeight, 0, 1)
			if fade <= 0.01 {
				continue
			}
			alpha := shadowStrength * fade
			top := groundY + 0.02

			// Sample the radial gradient at this cell's place in the footprint.
			vu0 := u0 + (u1-u0)*((cx0-minX)/diameter)
			vu1 := u0 + (u1-u0)*((cx1-minX)/diameter)
			vv0 := v0 + (v1-v0)*((cz0-minZ)/diameter)
			vv1 := v0 + (v1-v0)*((cz1-minZ)/diameter)

			base := uint32(len(dst.Vertices))
			pts := [4]mathx.Vec3{
				{X: cx0, Y: top, Z: cz0},
				{X: cx1, Y: top, Z: cz0},
				{X: cx1, Y: top, Z: cz1},
				{X: cx0, Y: top, Z: cz1},
			}
			uvs := [4][2]float32{{vu0, vv0}, {vu1, vv0}, {vu1, vv1}, {vu0, vv1}}
			for i := 0; i < 4; i++ {
				dst.Vertices = append(dst.Vertices, world.Vertex{
					X: pts[i].X, Y: pts[i].Y, Z: pts[i].Z,
					U: uvs[i][0], V: uvs[i][1],
					Light: 1, Sky: 1, Block: 0, Alpha: alpha,
				})
			}
			dst.Indices = append(dst.Indices, base, base+1, base+2, base, base+2, base+3)
		}
	}
}

// appendEntityShadows adds the ground shadow for every visible entity. It is
// drawn in the blended transparent pass so it darkens terrain beneath the
// entity without occluding anything.
func (g *Game) appendEntityShadows(dst *world.MeshData) {
	if g.world == nil || g.mobs == nil {
		return
	}
	for _, mob := range g.mobs.Mobs {
		if mob.Dead {
			continue
		}
		p := mob.Body.Pos
		appendShadow(dst, g.world, p.X, p.Y, p.Z, mob.Body.Width)
	}
	for _, d := range g.mobs.Items {
		if d.Dead {
			continue
		}
		p := d.Body.Pos
		appendShadow(dst, g.world, p.X, p.Y, p.Z, 0.45)
	}
	for _, f := range g.mobs.Falling {
		p := f.Body.Pos
		appendShadow(dst, g.world, p.X, p.Y, p.Z, 0.9)
	}
	if g.player != nil && g.camera != cameraFirst {
		p := g.player.Body.Pos
		appendShadow(dst, g.world, p.X, p.Y, p.Z, g.player.Body.Width)
	}
}
