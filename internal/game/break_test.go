package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/world"
)

func TestBreakOverlayGeometry(t *testing.T) {
	g := newTestGame(t)
	g.world.SetBlock(8, 100, 8, core.Stone)
	g.breakPos = [3]int{8, 100, 8}
	g.targetValid = true
	g.breakActive = true
	g.breakProgress = 0.45
	var m world.MeshData
	g.appendBreakOverlay(&m)
	if len(m.Vertices) == 0 || len(m.Indices) == 0 {
		t.Fatal("break overlay produced no geometry")
	}
	u0, _, _, _ := world.TileUV(core.TileCrack4)
	if m.Vertices[0].U != u0 {
		t.Fatalf("break overlay has wrong crack stage UV: %v", m.Vertices[0].U)
	}
}
