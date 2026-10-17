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
