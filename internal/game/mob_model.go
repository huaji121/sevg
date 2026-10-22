package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

// Mobs are modelled with the standard Minecraft scale of 1 model pixel = 1/16
// block, so every box dimension below reads directly off the 64x64 skin layout
// (same box-unwrap convention as the player skin). The feet sit at y = 0 and
// the front of the animal faces -Z, matching entity.Mob.Forward.
const mobUnit = float32(1.0 / 16.0)

const mobSkinSize = float32(64)

type mobSkinFaces [6][4]float32

func mobSkinFace(x0, y0, x1, y1 float32) [4]float32 {
	return [4]float32{x0 / mobSkinSize, y0 / mobSkinSize, x1 / mobSkinSize, y1 / mobSkinSize}
}

// mobBoxUV lays out the six faces of a w x h x d pixel box at texture origin
// (tx, ty) using the standard Minecraft box unwrap. The face order matches
// entityFaces: +X, -X, +Y, -Y, +Z, -Z.
func mobBoxUV(tx, ty, w, h, d float32) mobSkinFaces {
	return mobSkinFaces{
		mobSkinFace(tx, ty+d, tx+d, ty+d+h),             // +X
		mobSkinFace(tx+d+w, ty+d, tx+d+w+d, ty+d+h),     // -X
		mobSkinFace(tx+d, ty, tx+d+w, ty+d),             // +Y
		mobSkinFace(tx+d+w, ty, tx+d+w+w, ty+d),         // -Y
		mobSkinFace(tx+d+w+d, ty+d, tx+d+w+d+w, ty+d+h), // +Z
		mobSkinFace(tx+d, ty+d, tx+d+w, ty+d+h),         // -Z
	}
}

// mobPart is one textured box of a mob model. Coordinates are model pixels;
// min/max/pivot are converted with mobUnit before being emitted.
type mobPart struct {
	min, max, pivot mathx.Vec3
	rotX            float32
	tile            uint8
	faces           mobSkinFaces
}

func mobPx(x, y, z float32) mathx.Vec3 {
	return mathx.V3(x*mobUnit, y*mobUnit, z*mobUnit)
}

func mobBox(minX, minY, minZ, maxX, maxY, maxZ float32) (mathx.Vec3, mathx.Vec3) {
	return mobPx(minX, minY, minZ), mobPx(maxX, maxY, maxZ)
}

func appendMobPart(dst *world.MeshData, part mobPart, yaw, light float32, pos mathx.Vec3) {
	tu0, tv0, tu1, tv1 := world.TileUV(part.tile)
	for fi, f := range entityFaces {
		r := part.faces[fi]
		u0 := tu0 + (tu1-tu0)*r[0]
		v0 := tv0 + (tv1-tv0)*r[1]
		u1 := tu0 + (tu1-tu0)*r[2]
		v1 := tv0 + (tv1-tv0)*r[3]
		base := uint32(len(dst.Vertices))
		for i := 0; i < 4; i++ {
			c := f.corners[i]
			x := part.min.X + (part.max.X-part.min.X)*c[0]
			y := part.min.Y + (part.max.Y-part.min.Y)*c[1]
			z := part.min.Z + (part.max.Z-part.min.Z)*c[2]
			p := mathx.Vec3{X: x, Y: y, Z: z}.Sub(part.pivot)
			p = rotateX(p, part.rotX).Add(part.pivot)
			q := rotateY(p, yaw).Add(pos)
			dst.Vertices = append(dst.Vertices, world.Vertex{
				X: q.X, Y: q.Y, Z: q.Z,
				U:     u0 + (u1-u0)*f.uv[i][0],
				V:     v0 + (v1-v0)*f.uv[i][1],
				Light: f.light * light,
			})
		}
		dst.Indices = append(dst.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}

// appendMobModel builds the textured, animated model for one mob.
func appendMobModel(dst *world.MeshData, mob *entity.Mob) {
	flash := clampF(mob.HitFlash/0.25, 0, 1)
	light := 1 + flash*1.6
	switch mob.Kind {
	case entity.Sheep:
		appendSheepModel(dst, mob, light)
	default:
		appendPigModel(dst, mob, light)
	}
}

// mobGait returns the horizontal movement factor and the phase sine used to
// swing the legs. Quadrupeds trot, so the diagonal leg pairs move together.
func mobGait(mob *entity.Mob, speedRef, swingAmt float32) (move, phase, swing float32) {
	speed := float32(math.Hypot(float64(mob.Body.Vel.X), float64(mob.Body.Vel.Z)))
	move = clampF(speed/speedRef, 0, 1)
	phase = float32(math.Sin(float64(mob.WalkPhase)))
	swing = phase * move * swingAmt
	return
}

func appendPigModel(dst *world.MeshData, mob *entity.Mob, light float32) {
	pos, yaw := mob.Body.Pos, mob.Yaw
	move, phase, legSwing := mobGait(mob, 1.6, 0.7)
	headBob := phase * move * 0.05

	bodyUV := mobBoxUV(0, 16, 8, 8, 12)
	headUV := mobBoxUV(0, 0, 8, 8, 8)
	snoutUV := mobBoxUV(32, 0, 4, 3, 1)
	legUV := mobBoxUV(42, 0, 4, 6, 4)

	// Body.
	bmin, bmax := mobBox(-4, 6, -6, 4, 14, 6)
	appendMobPart(dst, mobPart{min: bmin, max: bmax, tile: core.TilePig, faces: bodyUV}, yaw, light, pos)

	// Head and snout nod together around the neck.
	neck := mobPx(0, 10, -6)
	hmin, hmax := mobBox(-4, 6, -14, 4, 14, -6)
	appendMobPart(dst, mobPart{min: hmin, max: hmax, pivot: neck, rotX: headBob, tile: core.TilePig, faces: headUV}, yaw, light, pos)
	smin, smax := mobBox(-2, 8, -15, 2, 11, -14)
	appendMobPart(dst, mobPart{min: smin, max: smax, pivot: neck, rotX: headBob, tile: core.TilePig, faces: snoutUV}, yaw, light, pos)

	appendLegs(dst, yaw, light, pos, core.TilePig, legUV, 6, -3, 3, legSwing)
}

func appendSheepModel(dst *world.MeshData, mob *entity.Mob, light float32) {
	pos, yaw := mob.Body.Pos, mob.Yaw
	move, phase, legSwing := mobGait(mob, 1.6, 0.6)
	headBob := phase * move * 0.04

	bodyUV := mobBoxUV(0, 16, 8, 8, 12)
	headUV := mobBoxUV(0, 0, 6, 6, 6)
	legUV := mobBoxUV(28, 0, 4, 12, 4)
	woolBodyUV := mobBoxUV(0, 16, 10, 10, 14)
	woolCapUV := mobBoxUV(0, 0, 7, 4, 6)

	// Base body, head and legs.
	bmin, bmax := mobBox(-4, 12, -6, 4, 20, 6)
	appendMobPart(dst, mobPart{min: bmin, max: bmax, tile: core.TileSheepSkin, faces: bodyUV}, yaw, light, pos)

	neck := mobPx(0, 15, -6)
	hmin, hmax := mobBox(-3, 12, -12, 3, 18, -6)
	appendMobPart(dst, mobPart{min: hmin, max: hmax, pivot: neck, rotX: headBob, tile: core.TileSheepSkin, faces: headUV}, yaw, light, pos)

	appendLegs(dst, yaw, light, pos, core.TileSheepSkin, legUV, 12, -4, 4, legSwing)

	// Wool coat: an outer shell over the body plus a cap over the head.
	wmin, wmax := mobBox(-5, 11, -7, 5, 21, 7)
	appendMobPart(dst, mobPart{min: wmin, max: wmax, tile: core.TileSheepWool, faces: woolBodyUV}, yaw, light, pos)
	cmin, cmax := mobBox(-3.5, 17, -12, 3.5, 21, -6)
	appendMobPart(dst, mobPart{min: cmin, max: cmax, pivot: neck, rotX: headBob, tile: core.TileSheepWool, faces: woolCapUV}, yaw, light, pos)
}

// appendLegs emits the four 4-pixel-wide legs with a diagonal trot: the
// front-left and back-right legs swing together, the other diagonal opposes.
// legTop is the pivot height, frontZ/backZ the leg rows.
func appendLegs(dst *world.MeshData, yaw, light float32, pos mathx.Vec3, tile uint8, legUV mobSkinFaces, legTop, frontZ, backZ, swing float32) {
	leg := func(side, z, rot float32) {
		x0, x1 := float32(0), float32(4)
		if side < 0 {
			x0, x1 = -4, 0
		}
		mn, mx := mobBox(x0, 0, z-2, x1, legTop, z+2)
		appendMobPart(dst, mobPart{
			min: mn, max: mx,
			pivot: mobPx(side*2, legTop, z),
			rotX:  rot, tile: tile, faces: legUV,
		}, yaw, light, pos)
	}
	leg(1, frontZ, swing)
	leg(-1, frontZ, -swing)
	leg(1, backZ, -swing)
	leg(-1, backZ, swing)
}
