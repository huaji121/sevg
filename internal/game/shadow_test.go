package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func shadowTileRect() [4]float32 {
	u0, v0, u1, v1 := world.TileUV(core.TileShadow)
	return [4]float32{u0, v0, u1, v1}
}

func maxAlpha(m world.MeshData) float32 {
	best := float32(-1)
	for _, v := range m.Vertices {
		if v.Alpha > best {
			best = v.Alpha
		}
	}
	return best
}

func TestShadowAppearsOnGround(t *testing.T) {
	g := newTestGame(t)
	for y := 101; y <= 110; y++ {
		g.world.SetBlock(8, y, 8, core.Air)
	}
	g.world.SetBlock(8, 100, 8, core.Stone)
	g.mobs.SpawnAt(entity.Pig, mathx.Vec3{X: 8.5, Y: 101, Z: 8.5})

	var m world.MeshData
	g.appendEntityShadows(&m)
	if len(m.Vertices) == 0 || len(m.Vertices)%4 != 0 || len(m.Indices) != len(m.Vertices)/2*3 {
		t.Fatalf("shadow mesh: %d verts / %d indices", len(m.Vertices), len(m.Indices))
	}
	r := shadowTileRect()
	grounded := false
	for _, v := range m.Vertices {
		if v.U < r[0]-1e-4 || v.U > r[2]+1e-4 || v.V < r[1]-1e-4 || v.V > r[3]+1e-4 {
			t.Fatalf("shadow vertex outside the shadow tile: %f,%f", v.U, v.V)
		}
		if v.Alpha <= 0 || v.Alpha > 1 {
			t.Fatalf("shadow alpha out of range: %f", v.Alpha)
		}
		if v.Y == 101.02 {
			grounded = true
		}
	}
	if !grounded {
		t.Fatal("no shadow patch resting on the ground surface")
	}
}

func TestShadowFadesWithHeight(t *testing.T) {
	g := newTestGame(t)
	for y := 99; y <= 110; y++ {
		g.world.SetBlock(8, y, 8, core.Air)
	}
	g.world.SetBlock(8, 99, 8, core.Stone)
	mob := g.mobs.SpawnAt(entity.Pig, mathx.Vec3{X: 8.5, Y: 100, Z: 8.5})

	var onGround world.MeshData
	g.appendEntityShadows(&onGround)
	base := maxAlpha(onGround)
	if base <= 0 {
		t.Fatal("no shadow on the ground")
	}

	mob.Body.Pos.Y = 101
	var lifted world.MeshData
	g.appendEntityShadows(&lifted)
	if a := maxAlpha(lifted); a >= base {
		t.Fatalf("shadow did not fade with height: %f vs %f", a, base)
	}

	// Far enough above the ground the shadow disappears entirely.
	mob.Body.Pos.Y = 104
	var high world.MeshData
	g.appendEntityShadows(&high)
	if len(high.Vertices) != 0 {
		t.Fatalf("shadow did not fade out: %d verts", len(high.Vertices))
	}
}

func TestShadowConformsToSteps(t *testing.T) {
	g := newTestGame(t)
	for y := 99; y <= 112; y++ {
		for _, bx := range []int{8, 9} {
			g.world.SetBlock(bx, y, 8, core.Air)
		}
	}
	g.world.SetBlock(8, 100, 8, core.Stone) // higher ledge
	g.world.SetBlock(9, 99, 8, core.Stone)  // lower ledge
	g.mobs.SpawnAt(entity.Pig, mathx.Vec3{X: 8.9, Y: 101, Z: 8.5})

	var m world.MeshData
	g.appendEntityShadows(&m)
	ys := map[float32]bool{}
	for _, v := range m.Vertices {
		ys[v.Y] = true
	}
	if !ys[101.02] || !ys[100.02] {
		t.Fatalf("shadow did not drop onto the lower ledge: %v", ys)
	}
}

func TestShadowSkipsWhenNoGround(t *testing.T) {
	g := newTestGame(t)
	for y := 242; y <= 250; y++ {
		g.world.SetBlock(8, y, 8, core.Air)
	}
	g.mobs.SpawnAt(entity.Sheep, mathx.Vec3{X: 8.5, Y: 250, Z: 8.5})
	var m world.MeshData
	g.appendEntityShadows(&m)
	if len(m.Vertices) != 0 {
		t.Fatalf("shadow rendered without ground below: %d verts", len(m.Vertices))
	}
}
