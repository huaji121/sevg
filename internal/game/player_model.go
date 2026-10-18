package game

import (
	"math"

	"sevg/internal/mathx"
	"sevg/internal/world"
)

var (
	playerSkin  = [4]float32{0.76, 0.55, 0.36, 1}
	playerShirt = [4]float32{0.18, 0.63, 0.61, 1}
	playerPants = [4]float32{0.20, 0.28, 0.59, 1}
	playerHair  = [4]float32{0.18, 0.12, 0.08, 1}
	playerShoe  = [4]float32{0.20, 0.20, 0.22, 1}
)

func rotateX(p mathx.Vec3, a float32) mathx.Vec3 {
	c, s := float32(math.Cos(float64(a))), float32(math.Sin(float64(a)))
	return mathx.Vec3{X: p.X, Y: p.Y*c - p.Z*s, Z: p.Y*s + p.Z*c}
}

func rotateY(p mathx.Vec3, a float32) mathx.Vec3 {
	c, s := float32(math.Cos(float64(a))), float32(math.Sin(float64(a)))
	return mathx.Vec3{X: p.X*c - p.Z*s, Y: p.Y, Z: p.X*s + p.Z*c}
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

// appendPlayerModel builds a Steve-like blocky player model at the player's
// feet. It is only visible in third-person view.
func (g *Game) appendPlayerModel(dst, icons, transparent *world.MeshData) {
	if g.player == nil {
		return
	}
	p := g.player
	speed := float32(math.Hypot(float64(p.Body.Vel.X), float64(p.Body.Vel.Z)))
	amp := clampF(speed/4.6, 0, 1) * 0.65
	if p.Flying || !p.Body.OnGround {
		amp *= 0.25
	}
	swing := float32(math.Sin(float64(p.WalkPhase))) * amp

	var model world.MeshData
	// Torso.
	appendPlayerPart(&model, mathx.V3(-0.25, 0.65, -0.15), mathx.V3(0.25, 1.30, 0.15), playerShirt, mathx.V3(0, 0.65, 0), 0)

	// Head, hair and face all follow head pitch around the neck.
	neck := mathx.V3(0, 1.30, 0)
	pitch := p.Pitch
	appendPlayerPart(&model, mathx.V3(-0.25, 1.30, -0.25), mathx.V3(0.25, 1.80, 0.25), playerSkin, neck, pitch)
	appendPlayerPart(&model, mathx.V3(-0.26, 1.78, -0.26), mathx.V3(0.26, 1.84, 0.26), playerHair, neck, pitch)
	appendPlayerPart(&model, mathx.V3(-0.26, 1.40, 0.24), mathx.V3(0.26, 1.80, 0.27), playerHair, neck, pitch)
	// Eyes.
	appendPlayerPart(&model, mathx.V3(-0.17, 1.55, -0.256), mathx.V3(-0.06, 1.61, -0.25), [4]float32{0.93, 0.93, 0.90, 1}, neck, pitch)
	appendPlayerPart(&model, mathx.V3(0.06, 1.55, -0.256), mathx.V3(0.17, 1.61, -0.25), [4]float32{0.93, 0.93, 0.90, 1}, neck, pitch)
	appendPlayerPart(&model, mathx.V3(-0.13, 1.56, -0.259), mathx.V3(-0.09, 1.60, -0.256), [4]float32{0.07, 0.07, 0.09, 1}, neck, pitch)
	appendPlayerPart(&model, mathx.V3(0.09, 1.56, -0.259), mathx.V3(0.13, 1.60, -0.256), [4]float32{0.07, 0.07, 0.09, 1}, neck, pitch)
	appendPlayerPart(&model, mathx.V3(-0.08, 1.40, -0.257), mathx.V3(0.08, 1.43, -0.25), [4]float32{0.28, 0.15, 0.12, 1}, neck, pitch)

	// Arms. They hang from the shoulders and swing opposite to the legs.
	appendPlayerPart(&model, mathx.V3(0.27, 0.62, -0.10), mathx.V3(0.45, 1.26, 0.10), playerSkin, mathx.V3(0.36, 1.25, 0), -swing)
	appendPlayerPart(&model, mathx.V3(0.26, 1.05, -0.11), mathx.V3(0.46, 1.27, 0.11), playerShirt, mathx.V3(0.36, 1.25, 0), -swing)
	appendPlayerPart(&model, mathx.V3(-0.45, 0.62, -0.10), mathx.V3(-0.27, 1.26, 0.10), playerSkin, mathx.V3(-0.36, 1.25, 0), swing)
	appendPlayerPart(&model, mathx.V3(-0.46, 1.05, -0.11), mathx.V3(-0.26, 1.27, 0.11), playerShirt, mathx.V3(-0.36, 1.25, 0), swing)

	// Legs with darker shoes.
	appendPlayerPart(&model, mathx.V3(0.01, 0.08, -0.12), mathx.V3(0.24, 0.66, 0.12), playerPants, mathx.V3(0.125, 0.65, 0), swing)
	appendPlayerPart(&model, mathx.V3(0.01, 0.0, -0.13), mathx.V3(0.24, 0.08, 0.13), playerShoe, mathx.V3(0.125, 0.65, 0), swing)
	appendPlayerPart(&model, mathx.V3(-0.24, 0.08, -0.12), mathx.V3(-0.01, 0.66, 0.12), playerPants, mathx.V3(-0.125, 0.65, 0), -swing)
	appendPlayerPart(&model, mathx.V3(-0.24, 0.0, -0.13), mathx.V3(-0.01, 0.08, 0.13), playerShoe, mathx.V3(-0.125, 0.65, 0), -swing)

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
	g.appendPlayerHeldIconCard(icons, swing)
	g.appendFirstPersonHeld(dst, icons, transparent)
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

// appendFirstPersonHeld reuses the exact first-person held-item meshes and
// recenters them on the player's right hand for second/third-person rendering.
func (g *Game) appendFirstPersonHeld(opaque, icons, transparent *world.MeshData) {
	if g.player == nil {
		return
	}
	oldCamera, oldState := g.camera, g.state
	g.camera, g.state = cameraFirst, StatePlaying
	held := g.buildHeldMesh()
	heldTrans := g.buildHeldTransparentMesh()
	g.camera, g.state = oldCamera, oldState

	p := g.player
	pos := p.Body.Pos
	right := mathx.Vec3{X: float32(math.Cos(float64(p.Yaw))), Z: float32(math.Sin(float64(p.Yaw)))}
	fwd := mathx.Vec3{X: float32(math.Sin(float64(p.Yaw))), Z: -float32(math.Cos(float64(p.Yaw)))}
	target := pos.Add(right.Mul(0.52)).Add(mathx.Vec3{Y: 0.58}).Add(fwd.Mul(0.06))
	speed := float32(math.Hypot(float64(p.Body.Vel.X), float64(p.Body.Vel.Z)))
	amp := clampF(speed/4.6, 0, 1) * 0.65
	if p.Flying || !p.Body.OnGround {
		amp *= 0.25
	}
	swing := float32(math.Sin(float64(p.WalkPhase))) * amp
	pivot := mathx.V3(0.36, 1.25, 0)
	local := rotateY(target.Sub(pos), -p.Yaw)
	local = rotateX(local.Sub(pivot), -swing).Add(pivot)
	target = rotateY(local, p.Yaw).Add(pos)

	appendHeldAt := func(dst *world.MeshData, src world.MeshData, flip bool) {
		if len(src.Vertices) == 0 {
			return
		}
		c := mathx.Vec3{}
		for _, v := range src.Vertices {
			c = c.Add(mathx.Vec3{X: v.X, Y: v.Y, Z: v.Z})
		}
		c = c.Mul(1 / float32(len(src.Vertices)))
		delta := target.Sub(c)
		base := uint32(len(dst.Vertices))
		for _, v := range src.Vertices {
			v.X += delta.X
			v.Y += delta.Y
			v.Z += delta.Z
			dst.Vertices = append(dst.Vertices, v)
		}
		for i := 0; i < len(src.Indices); i += 3 {
			a, b, cidx := base+src.Indices[i], base+src.Indices[i+1], base+src.Indices[i+2]
			if flip {
				// The first-person view model uses a mirrored camera basis.
				// Reverse its triangle winding when reusing it in world space.
				b, cidx = cidx, b
			}
			dst.Indices = append(dst.Indices, a, b, cidx)
		}
	}
	appendHeldAt(opaque, held, true)
	appendHeldAt(transparent, heldTrans, false)
}
