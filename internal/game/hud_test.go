package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func TestDroppedBlockUsesPlacedBlockTiles(t *testing.T) {
	var m world.MeshData
	appendBlockBox(&m, mathx.Vec3{}, mathx.Vec3{X: 0.3, Y: 0.3, Z: 0.3}, core.Grass)
	if len(m.Vertices) != 24 || len(m.Indices) != 36 {
		t.Fatalf("unexpected cube size: %d vertices, %d indices", len(m.Vertices), len(m.Indices))
	}
	// Face order is +X, -X, +Y, -Y, +Z, -Z; each face has four vertices.
	sideU, _, _, sideV := world.TileUV(core.TileGrassSide)
	topU, _, _, topV := world.TileUV(core.TileGrassTop)
	bottomU, bottomV, _, _ := world.TileUV(core.TileDirt)
	if m.Vertices[0].U != sideU || m.Vertices[0].V != sideV {
		t.Fatalf("side face UV does not match grass side")
	}
	if m.Vertices[8].U != topU || m.Vertices[8].V != topV {
		t.Fatalf("top face UV does not match grass top")
	}
	if m.Vertices[12].U != bottomU || m.Vertices[12].V != bottomV {
		t.Fatalf("bottom face UV does not match dirt")
	}
}
func TestPlantDropGeometry(t *testing.T) {
	var m world.MeshData
	appendPlant(&m, mathx.Vec3{X: 0.5, Y: 64.2, Z: 0.5}, 0.34, 0.42, core.TallGrass)
	if len(m.Vertices) != 8 || len(m.Indices) != 12 {
		t.Fatalf("plant vertices=%d indices=%d", len(m.Vertices), len(m.Indices))
	}
	u0, _, _, v1 := world.TileUV(core.TallGrass.Tile(2))
	if m.Vertices[0].U != u0 || m.Vertices[0].V != v1 {
		t.Fatalf("plant does not use the tall grass tile: %+v", m.Vertices[0])
	}
}

func TestToolDropGeometry(t *testing.T) {
	var sword world.MeshData
	appendSword(&sword, mathx.Vec3{X: 0.5, Y: 64.2, Z: 0.5}, core.TileWood)
	if len(sword.Vertices) == 0 || len(sword.Indices) == 0 {
		t.Fatal("sword has no geometry")
	}
	var pick world.MeshData
	appendPickaxe(&pick, mathx.Vec3{X: 0.5, Y: 64.2, Z: 0.5}, core.TileGray)
	if len(pick.Vertices) == 0 || len(pick.Indices) == 0 {
		t.Fatal("pickaxe has no geometry")
	}
}

func TestGlassDropUsesTransparentMesh(t *testing.T) {
	g := newTestGame(t)
	g.mobs.Drop(core.ItemGlass, 1, mathx.Vec3{X: 0.5, Y: 64.2, Z: 0.5}, mathx.Vec3{})
	opaque, transparent := g.buildEntityMeshes()
	if len(transparent.Vertices) != 24 || len(transparent.Indices) != 36 {
		t.Fatalf("glass drop transparent mesh: %d/%d", len(transparent.Vertices), len(transparent.Indices))
	}
	if len(opaque.Vertices) != 0 {
		t.Fatalf("glass drop leaked into opaque mesh: %d vertices", len(opaque.Vertices))
	}
}
func TestHeldItemMesh(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 4}
	m := g.buildHeldMesh()
	if len(m.Vertices) == 0 || len(m.Indices) == 0 {
		t.Fatal("held planks produced no mesh")
	}
	g.inv.Slots[0] = core.Stack{}
	if m = g.buildHeldMesh(); len(m.Vertices) != 0 {
		t.Fatal("empty hand should produce no mesh")
	}
	g.inv.Slots[0] = core.Stack{Item: core.ItemStonePickaxe, Count: 1}
	if m = g.buildHeldMesh(); len(m.Vertices) != 0 {
		t.Fatal("tool should use the inventory icon, not the 3D held mesh")
	}
}
