package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/entity"
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

// appendPlayerHeld rotates a carried block around the right shoulder.
func appendPlayerHeld(dst *world.MeshData, center mathx.Vec3, half float32, block core.BlockID, pivot mathx.Vec3, rotX float32) {
	var part world.MeshData
	appendBlockBox(&part, center.Sub(mathx.V3(half, half, half)), center.Add(mathx.V3(half, half, half)), block)
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
func (g *Game) appendPlayerModel(dst *world.MeshData) {
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

	// Carried block in the right hand. Tools and plants keep the hand empty;
	// blocks use their real per-face atlas tiles.
	if st := g.inv.SelectedStack(); !st.Empty() {
		if def := st.Item.Def(); def.Kind == core.KindBlock {
			switch entity.VisualForItem(st.Item) {
			case entity.DropVisualBlock:
				appendPlayerHeld(&model, mathx.V3(0.49, 0.58, -0.20), 0.13, def.Block, mathx.V3(0.36, 1.25, 0), -swing)
			}
		}
	}

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
}
