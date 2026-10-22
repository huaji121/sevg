package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func TestMobModelsRender(t *testing.T) {
	g := newTestGame(t)
	pig := g.mobs.SpawnAt(entity.Pig, mathx.Vec3{X: 8.5, Y: 70, Z: 8.5})
	sheep := g.mobs.SpawnAt(entity.Sheep, mathx.Vec3{X: 10.5, Y: 70, Z: 8.5})
	pig.WalkPhase = 0.4
	sheep.WalkPhase = 1.1
	opaque, _, _ := g.buildEntityMeshes()
	if len(opaque.Vertices) == 0 || len(opaque.Indices) == 0 {
		t.Fatal("mob models produced no geometry")
	}
}

func TestMobSkinUVsStayInTiles(t *testing.T) {
	g := newTestGame(t)
	inRect := func(r [4]float32, u, v float32) bool {
		return u >= r[0]-1e-4 && u <= r[2]+1e-4 && v >= r[1]-1e-4 && v <= r[3]+1e-4
	}
	pigTile := tileRect(core.TilePig)
	sheepSkin := tileRect(core.TileSheepSkin)
	sheepWool := tileRect(core.TileSheepWool)

	for _, tc := range []struct {
		kind  entity.Kind
		rects [][4]float32
	}{
		{entity.Pig, [][4]float32{pigTile}},
		{entity.Sheep, [][4]float32{sheepSkin, sheepWool}},
	} {
		mob := g.mobs.SpawnAt(tc.kind, mathx.Vec3{X: 8.5, Y: 70, Z: 8.5})
		var m world.MeshData
		appendMobModel(&m, mob)
		if len(m.Vertices) == 0 {
			t.Fatalf("%s model is empty", tc.kind.Name())
		}
		for _, v := range m.Vertices {
			ok := false
			for _, r := range tc.rects {
				if inRect(r, v.U, v.V) {
					ok = true
					break
				}
			}
			if !ok {
				t.Fatalf("%s vertex UV %f,%f outside its skin tiles", tc.kind.Name(), v.U, v.V)
			}
		}
	}
}

func tileRect(tile uint8) [4]float32 {
	u0, v0, u1, v1 := world.TileUV(tile)
	return [4]float32{u0, v0, u1, v1}
}

func TestMobLegsAnimateWithWalkPhase(t *testing.T) {
	g := newTestGame(t)
	mob := g.mobs.SpawnAt(entity.Pig, mathx.Vec3{X: 8.5, Y: 70, Z: 8.5})
	mob.Body.Vel = mathx.Vec3{Z: -1.6} // full speed so the gait is active

	mob.WalkPhase = 0
	var idle world.MeshData
	appendMobModel(&idle, mob)

	mob.WalkPhase = 1.5707963 // sin -> 1
	var step world.MeshData
	appendMobModel(&step, mob)

	if len(idle.Vertices) != len(step.Vertices) {
		t.Fatalf("vertex count changed: %d vs %d", len(idle.Vertices), len(step.Vertices))
	}
	moved := false
	for i := range idle.Vertices {
		if idle.Vertices[i] != step.Vertices[i] {
			moved = true
			break
		}
	}
	if !moved {
		t.Fatal("legs did not move between walk phases")
	}
}

func TestMobModelFollowsYaw(t *testing.T) {
	g := newTestGame(t)
	mob := g.mobs.SpawnAt(entity.Pig, mathx.Vec3{X: 8.5, Y: 70, Z: 8.5})
	mob.Yaw = 0
	var a world.MeshData
	appendMobModel(&a, mob)
	mob.Yaw = 1.2
	var b world.MeshData
	appendMobModel(&b, mob)
	if a.Vertices[0] == b.Vertices[0] {
		t.Fatal("model did not rotate with yaw")
	}
}