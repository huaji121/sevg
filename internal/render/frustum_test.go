package render

import (
	"testing"

	"sevg/internal/mathx"
	"sevg/internal/world"
)

func TestFrustumChunkAABB(t *testing.T) {
	cam := Camera{
		Pos: mathx.V3(0, 64, 0),
		Dir: mathx.V3(0, 0, -1),
		Up:  mathx.V3(0, 1, 0),
		Fov: mathx.Radians(70), Aspect: 1, Near: 0.1, Far: 500,
	}
	f := newFrustum(cam.ViewProj())
	front := [2]mathx.Vec3{
		mathx.V3(-8, 0, -40),
		mathx.V3(8, world.ChunkHeight, -24),
	}
	back := [2]mathx.Vec3{
		mathx.V3(-8, 0, 24),
		mathx.V3(8, world.ChunkHeight, 40),
	}
	side := [2]mathx.Vec3{
		mathx.V3(92, 0, -40),
		mathx.V3(108, world.ChunkHeight, -24),
	}
	if !f.containsAABB(front[0], front[1]) {
		t.Fatal("chunk in front was culled")
	}
	if f.containsAABB(back[0], back[1]) {
		t.Fatal("chunk behind camera was not culled")
	}
	if f.containsAABB(side[0], side[1]) {
		t.Fatal("chunk outside side plane was not culled")
	}
}
func TestChunkDistanceUsesFullVerticalBounds(t *testing.T) {
	gc := &gpuChunk{pos: world.ChunkPos{0, 0}}
	for _, y := range []float32{190, 191, 220, 256} {
		cam := Camera{Pos: mathx.V3(8, y, 8), Dir: mathx.V3(0, -1, 0), Up: mathx.V3(0, 0, -1), Fov: mathx.Radians(70), Aspect: 1, Near: 0.1, Far: 500}
		f := newFrustum(cam.ViewProj())
		if !(&Renderer{}).chunkVisible(gc, cam, 88, &f) {
			t.Fatalf("chunk below camera at y=%f was culled", y)
		}
	}
	min := mathx.V3(0, 0, 0)
	max := mathx.V3(16, world.ChunkHeight, 16)
	if d := chunkAABBDistSq(min, max, mathx.V3(8, 300, 8)); d <= 0 {
		t.Fatalf("distance above chunk = %f", d)
	}
}
