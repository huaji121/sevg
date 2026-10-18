package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/ui"
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

func TestGlassDropUsesTransparentMesh(t *testing.T) {
	g := newTestGame(t)
	g.mobs.Drop(core.ItemGlass, 1, mathx.Vec3{X: 0.5, Y: 64.2, Z: 0.5}, mathx.Vec3{})
	opaque, transparent, _ := g.buildEntityMeshes()
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
func TestIconCardGeometry(t *testing.T) {
	var m world.MeshData
	icon := ui.Glyph{U0: 0.1, V0: 0.2, U1: 0.3, V1: 0.4}
	appendIconCard(&m, mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}, mathx.V3(1, 0, 0), mathx.V3(0, 1, 0), mathx.V3(0, 0, 1), 0.3, 0.3, 0.012, icon, []ui.IconSide{{AX: -0.5, AY: -0.5, BX: -0.5, BY: 0.5, U: 0.2, V: 0.3}})
	if len(m.Vertices) != 12 || len(m.Indices) != 18 {
		t.Fatalf("icon card vertices=%d indices=%d", len(m.Vertices), len(m.Indices))
	}
	if m.Vertices[0].U != icon.U0 || m.Vertices[0].V != icon.V1 {
		t.Fatalf("icon card front UV = %v,%v", m.Vertices[0].U, m.Vertices[0].V)
	}
}
func TestHeldTransparentGlassMesh(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.inv.Slots[0] = core.Stack{Item: core.ItemGlass, Count: 4}
	m := g.buildHeldTransparentMesh()
	if len(m.Vertices) == 0 || len(m.Indices) == 0 {
		t.Fatal("held glass produced no transparent mesh")
	}
}
