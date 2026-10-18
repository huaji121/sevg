package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func rotateX(p mathx.Vec3, a float32) mathx.Vec3 {
	c, s := float32(math.Cos(float64(a))), float32(math.Sin(float64(a)))
	return mathx.Vec3{X: p.X, Y: p.Y*c - p.Z*s, Z: p.Y*s + p.Z*c}
}

func rotateY(p mathx.Vec3, a float32) mathx.Vec3 {
	c, s := float32(math.Cos(float64(a))), float32(math.Sin(float64(a)))
	return mathx.Vec3{X: p.X*c - p.Z*s, Y: p.Y, Z: p.X*s + p.Z*c}
}

var (
	skinSkinUV  = [4]float32{0.00, 0.00, 0.25, 0.25}
	skinFaceUV  = [4]float32{0.25, 0.00, 0.50, 0.25}
	skinShirtUV = [4]float32{0.50, 0.00, 0.75, 0.25}
	skinPantsUV = [4]float32{0.75, 0.00, 1.00, 0.25}
	skinHairUV  = [4]float32{0.00, 0.25, 0.25, 0.50}
	skinShoeUV  = [4]float32{0.25, 0.25, 0.50, 0.50}
)

// appendPlayerSkinBox maps one texture region to every side of a player part,
// with an optional different region for the front (-Z) face.
func appendPlayerSkinBox(dst *world.MeshData, min, max mathx.Vec3, pivot mathx.Vec3, rotX float32, side, front [4]float32) {
	tu0, tv0, tu1, tv1 := world.TileUV(core.TilePlayerSkinMap)
	for fi, f := range entityFaces {
		r := side
		if fi == 5 {
			r = front
		}
		u0 := tu0 + (tu1-tu0)*r[0]
		v0 := tv0 + (tv1-tv0)*r[1]
		u1 := tu0 + (tu1-tu0)*r[2]
		v1 := tv0 + (tv1-tv0)*r[3]
		base := uint32(len(dst.Vertices))
		for i := 0; i < 4; i++ {
			c := f.corners[i]
			x := min.X + (max.X-min.X)*c[0]
			y := min.Y + (max.Y-min.Y)*c[1]
			z := min.Z + (max.Z-min.Z)*c[2]
			p := mathx.Vec3{X: x, Y: y, Z: z}.Sub(pivot)
			p = rotateX(p, rotX).Add(pivot)
			dst.Vertices = append(dst.Vertices, world.Vertex{
				X: p.X, Y: p.Y, Z: p.Z,
				U:     u0 + (u1-u0)*f.uv[i][0],
				V:     v0 + (v1-v0)*f.uv[i][1],
				Light: f.light, Sky: 1,
			})
		}
		dst.Indices = append(dst.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}

// appendPlayerPart rotates one box around a local pivot before adding it to
// the model mesh. This keeps the blocky Minecraft-style limbs articulated.
func appendPlayerPart(dst *world.MeshData, min, max mathx.Vec3, col [4]float32, pivot mathx.Vec3, rotX float32) {
	var part world.MeshData
	appendEntityBox(&part, min, max, col)
	base := uint32(len(dst.Vertices))
	for _, v := range part.Vertices {
		p := mathx.Vec3{X: v.X, Y: v.Y, Z: v.Z}.Sub(pivot)
		p = rotateX(p, rotX).Add(pivot)
		v.X, v.Y, v.Z = p.X, p.Y, p.Z
		dst.Vertices = append(dst.Vertices, v)
	}
	for _, idx := range part.Indices {
		dst.Indices = append(dst.Indices, base+idx)
	}
}

// appendPlayerHeldBlock places a fixed-orientation block in the right hand.
// It follows body yaw and arm swing, never head pitch.
func (g *Game) appendPlayerHeldBlock(opaque, transparent *world.MeshData, swing float32) {
	st := g.inv.SelectedStack()
	if st.Empty() || !st.Item.IsBlock() {
		return
	}
	def := st.Item.Def()
	dst := opaque
	switch entity.VisualForItem(st.Item) {
	case entity.DropVisualBlock:
	case entity.DropVisualGlass:
		dst = transparent
	default:
		return
	}
	center := mathx.V3(0.48, 0.56, -0.16)
	const half = float32(0.13)
	var part world.MeshData
	appendBlockBox(&part, center.Sub(mathx.V3(half, half, half)), center.Add(mathx.V3(half, half, half)), def.Block)
	p := g.player
	pivot := mathx.V3(0.36, 1.25, 0)
	base := uint32(len(dst.Vertices))
	for _, v := range part.Vertices {
		q := mathx.Vec3{X: v.X, Y: v.Y, Z: v.Z}.Sub(pivot)
		q = rotateX(q, -swing).Add(pivot)
		q = rotateY(q, p.Yaw).Add(p.Body.Pos)
		v.X, v.Y, v.Z = q.X, q.Y, q.Z
		dst.Vertices = append(dst.Vertices, v)
	}
	for _, idx := range part.Indices {
		dst.Indices = append(dst.Indices, base+idx)
	}
}

// appendPlayerRightArm builds the same right arm geometry used by the
// third-person player model, in player-local coordinates.
func appendPlayerRightArm(dst *world.MeshData, walkSwing, actionSwing float32) {
	arm := -walkSwing - actionSwing
	appendPlayerSkinBox(dst, mathx.V3(0.27, 0.62, -0.10), mathx.V3(0.45, 1.26, 0.10), mathx.V3(0.36, 1.25, 0), arm, skinSkinUV, skinSkinUV)
	appendPlayerSkinBox(dst, mathx.V3(0.26, 1.05, -0.11), mathx.V3(0.46, 1.27, 0.11), mathx.V3(0.36, 1.25, 0), arm, skinShirtUV, skinShirtUV)
}

// appendPlayerModel builds a Steve-like blocky player model at the player's
// feet. It is only visible in third-person view.
func (g *Game) appendPlayerModel(dst, icons, transparent *world.MeshData) {
	if g.player == nil {
		return
	}
	p := g.player
	speed := float32(math.Hypot(float64(p.Body.Vel.X), float64(p.Body.Vel.Z)))
	move := clampF(speed/4.6, 0, 1)
	airScale := float32(1)
	if p.Flying || !p.Body.OnGround {
		airScale = 0.25
	}
	phase := float32(math.Sin(float64(p.WalkPhase)))
	legSwing := phase * move * 0.65 * airScale
	armSwing := phase * move * 0.22 * airScale

	var model world.MeshData
	// Torso.
	appendPlayerSkinBox(&model, mathx.V3(-0.25, 0.65, -0.15), mathx.V3(0.25, 1.30, 0.15), mathx.V3(0, 0.65, 0), 0, skinShirtUV, skinShirtUV)

	// Head, hair and face all follow head pitch around the neck.
	neck := mathx.V3(0, 1.30, 0)
	pitch := p.Pitch
	appendPlayerSkinBox(&model, mathx.V3(-0.25, 1.30, -0.25), mathx.V3(0.25, 1.80, 0.25), neck, pitch, skinSkinUV, skinFaceUV)
	appendPlayerSkinBox(&model, mathx.V3(-0.26, 1.78, -0.26), mathx.V3(0.26, 1.84, 0.26), neck, pitch, skinHairUV, skinHairUV)
	appendPlayerSkinBox(&model, mathx.V3(-0.26, 1.40, 0.24), mathx.V3(0.26, 1.80, 0.27), neck, pitch, skinHairUV, skinHairUV)
	// Eyes.

	// Arms. They hang from the shoulders and swing opposite to the legs.
	appendPlayerRightArm(&model, armSwing, g.swingAngle())
	appendPlayerSkinBox(&model, mathx.V3(-0.45, 0.62, -0.10), mathx.V3(-0.27, 1.26, 0.10), mathx.V3(-0.36, 1.25, 0), armSwing, skinSkinUV, skinSkinUV)
	appendPlayerSkinBox(&model, mathx.V3(-0.46, 1.05, -0.11), mathx.V3(-0.26, 1.27, 0.11), mathx.V3(-0.36, 1.25, 0), armSwing, skinShirtUV, skinShirtUV)

	// Legs with darker shoes.
	appendPlayerSkinBox(&model, mathx.V3(0.01, 0.08, -0.12), mathx.V3(0.24, 0.66, 0.12), mathx.V3(0.125, 0.65, 0), legSwing, skinPantsUV, skinPantsUV)
	appendPlayerSkinBox(&model, mathx.V3(0.01, 0.0, -0.13), mathx.V3(0.24, 0.08, 0.13), mathx.V3(0.125, 0.65, 0), legSwing, skinShoeUV, skinShoeUV)
	appendPlayerSkinBox(&model, mathx.V3(-0.24, 0.08, -0.12), mathx.V3(-0.01, 0.66, 0.12), mathx.V3(-0.125, 0.65, 0), -legSwing, skinPantsUV, skinPantsUV)
	appendPlayerSkinBox(&model, mathx.V3(-0.24, 0.0, -0.13), mathx.V3(-0.01, 0.08, 0.13), mathx.V3(-0.125, 0.65, 0), -legSwing, skinShoeUV, skinShoeUV)

	// Rotate the whole model to the player's yaw and move it to the feet.
	base := uint32(len(dst.Vertices))
	pos := p.Body.Pos
	for _, v := range model.Vertices {
		q := rotateY(mathx.Vec3{X: v.X, Y: v.Y, Z: v.Z}, p.Yaw).Add(pos)
		v.X, v.Y, v.Z = q.X, q.Y, q.Z
		dst.Vertices = append(dst.Vertices, v)
	}
	for _, idx := range model.Indices {
		dst.Indices = append(dst.Indices, base+idx)
	}
	g.appendPlayerHeldIconCard(icons, armSwing+g.swingAngle())
	g.appendPlayerHeldBlock(dst, transparent, armSwing+g.swingAngle())
}

// appendPlayerHeldIconCard places a thick icon card in front of the right
// hand, angled between the player-facing plane and the perpendicular plane.
func (g *Game) appendPlayerHeldIconCard(dst *world.MeshData, swing float32) {
	st := g.inv.SelectedStack()
	if st.Empty() || st.Item.IsBlock() || g.canv == nil {
		return
	}
	glyph, ok := g.canv.Icons[st.Item]
	if !ok {
		return
	}
	center := mathx.V3(0.55, 0.65, -0.20)
	// Requested local-space composition:
	// 1. yaw 90 degrees around vertical Y back to the front-facing card,
	// 2. pitch 90 degrees forward around horizontal X, making it horizontal.
	right := mathx.V3(1, 0, 0)
	up := mathx.V3(0, 0, -1)
	normal := right.Cross(up)
	cardRight := right
	cardUp := up
	var card world.MeshData
	appendIconCard(&card, center, cardRight, cardUp, normal, 0.42, 0.42, 0.012, glyph, g.canv.IconSides[st.Item])
	p := g.player
	pivot := mathx.V3(0.36, 1.25, 0)
	base := uint32(len(dst.Vertices))
	for _, v := range card.Vertices {
		q := mathx.Vec3{X: v.X, Y: v.Y, Z: v.Z}.Sub(pivot)
		q = rotateX(q, -swing).Add(pivot)
		q = rotateY(q, p.Yaw).Add(p.Body.Pos)
		v.X, v.Y, v.Z = q.X, q.Y, q.Z
		dst.Vertices = append(dst.Vertices, v)
	}
	for _, idx := range card.Indices {
		dst.Indices = append(dst.Indices, base+idx)
	}
}
