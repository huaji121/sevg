package world

import (
	"testing"

	"sevg/internal/core"
)

// The opaque render pipeline uses back-face culling. Every block face must be
// wound so its geometric front normal points out of the block.
func TestBlockFacesWinding(t *testing.T) {
	for fi, f := range faces {
		a := [3]float32{f.corners[1][0] - f.corners[0][0], f.corners[1][1] - f.corners[0][1], f.corners[1][2] - f.corners[0][2]}
		b := [3]float32{f.corners[2][0] - f.corners[0][0], f.corners[2][1] - f.corners[0][1], f.corners[2][2] - f.corners[0][2]}
		n := [3]float32{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
		dot := n[0]*float32(f.normal[0]) + n[1]*float32(f.normal[1]) + n[2]*float32(f.normal[2])
		if dot <= 0 {
			t.Fatalf("face %d winding is inward: cross=%v normal=%v", fi, n, f.normal)
		}
	}
}
func TestCrossPlantDoubleSided(t *testing.T) {
	var m MeshData
	addCross(&m, 0, 0, 0, 0, 0, core.Flower, func(int, int, int) uint8 { return 15 }, func(int, int, int) uint8 { return 0 })
	if len(m.Vertices) != 8 || len(m.Indices) != 24 {
		t.Fatalf("cross plant vertices=%d indices=%d", len(m.Vertices), len(m.Indices))
	}
}
