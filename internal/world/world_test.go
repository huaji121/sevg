package world

import (
	"math"
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
)

func TestNoiseDeterministicAndBounded(t *testing.T) {
	a := NewNoise(123)
	b := NewNoise(123)
	for i := 0; i < 100; i++ {
		x, y, z := float64(i)*0.37, float64(i)*0.11, float64(i)*0.73
		if a.Noise3(x, y, z) != b.Noise3(x, y, z) {
			t.Fatal("noise is not deterministic")
		}
		v := a.Fbm2(x, y, 5, 2, 0.5)
		if v < -1 || v > 1 {
			t.Fatalf("fbm out of range: %f", v)
		}
	}
}

func TestGenerationDeterministic(t *testing.T) {
	g1, g2 := NewGenerator(99), NewGenerator(99)
	a := g1.GenerateChunk(ChunkPos{X: 2, Z: -3})
	b := g2.GenerateChunk(ChunkPos{X: 2, Z: -3})
	if len(a) != ChunkArea || len(b) != ChunkArea {
		t.Fatal("bad chunk size")
	}
	solid := 0
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("chunk differs at %d", i)
		}
		if a[i] != core.Air {
			solid++
		}
	}
	if solid < 1000 {
		t.Fatalf("terrain too empty: %d solid blocks", solid)
	}
	if a[0] != core.Bedrock {
		t.Fatalf("bottom block should be bedrock, got %v", a[0])
	}
}

func TestMeshWorldCoordinates(t *testing.T) {
	w := New(7)
	cp := ChunkPos{X: 1, Z: 1}
	for dz := -1; dz <= 1; dz++ {
		for dx := -1; dx <= 1; dx++ {
			w.EnsureChunk(ChunkPos{X: cp.X + dx, Z: cp.Z + dz})
		}
	}
	mesh, ok := w.Mesh(cp)
	if !ok {
		t.Fatal("mesh failed")
	}
	if len(mesh.Opaque.Vertices) == 0 || len(mesh.Opaque.Indices) == 0 {
		t.Fatal("opaque mesh empty")
	}
	baseX := float32(cp.X * ChunkSize)
	baseZ := float32(cp.Z * ChunkSize)
	for _, v := range mesh.Opaque.Vertices {
		if v.X < baseX-0.01 || v.X > baseX+ChunkSize+0.01 || v.Z < baseZ-0.01 || v.Z > baseZ+ChunkSize+0.01 {
			t.Fatalf("vertex outside chunk: %v", v)
		}
	}
}

func TestRaycastSurface(t *testing.T) {
	w := New(42)
	cp := ChunkPos{X: 0, Z: 0}
	w.EnsureChunk(cp)
	x, z := 8, 8
	h := -1
	for y := ChunkHeight - 1; y >= 0; y-- {
		b := w.GetBlock(x, y, z)
		if b != core.Air && b != core.Water {
			h = y
			break
		}
	}
	if h < 0 {
		t.Fatal("no surface")
	}
	hit := w.Raycast(mathx.Vec3{X: 8.5, Y: float32(h) + 8, Z: 8.5}, mathx.Vec3{Y: -1}, 20)
	if !hit.Hit || hit.Pos[1] != h {
		t.Fatalf("ray missed surface: %+v height=%d", hit, h)
	}
	if hit.Normal.Y <= 0 {
		t.Fatalf("expected upward normal, got %+v", hit.Normal)
	}
}

func TestEditsPersistInChunk(t *testing.T) {
	w := New(5)
	w.EnsureChunk(ChunkPos{})
	w.SetBlock(3, 80, 4, core.Brick)
	if got := w.GetBlock(3, 80, 4); got != core.Brick {
		t.Fatalf("edit not applied: %v", got)
	}
	recs := w.ExportEdits()
	if len(recs) != 1 {
		t.Fatalf("expected one edit, got %d", len(recs))
	}
}

func TestSafeSpawnClear(t *testing.T) {
	w := New(12345)
	p := w.SafeSpawn(mathx.Vec3{X: 0.5, Y: 0, Z: 0.5}, 12)
	x, y, z := int(math.Floor(float64(p.X))), int(math.Floor(float64(p.Y))), int(math.Floor(float64(p.Z)))
	if !w.GetBlock(x, y, z).IsAir() || !w.GetBlock(x, y+1, z).IsAir() {
		t.Fatalf("spawn not clear at %d,%d,%d", x, y, z)
	}
	if !w.GetBlock(x, y-1, z).Solid() {
		t.Fatalf("spawn has no ground at %d,%d,%d", x, y-1, z)
	}
}
func TestLeavesDoNotOccludeNeighborFaces(t *testing.T) {
	// Leaf textures contain transparent holes, so a solid block face next
	// to leaves must still be emitted.
	if !faceVisible(core.Stone, core.Leaves) {
		t.Fatal("stone face next to leaves was culled")
	}
	if faceVisible(core.Leaves, core.Stone) {
		t.Fatal("leaf face against solid stone should stay hidden")
	}
	if faceVisible(core.Leaves, core.Leaves) {
		t.Fatal("interior leaf faces should stay hidden")
	}
	if !faceVisible(core.Leaves, core.Air) {
		t.Fatal("outer leaf face should be visible")
	}
}
func TestMeshEmitsSolidFaceAgainstLeaves(t *testing.T) {
	w := New(11)
	cp := ChunkPos{}
	w.EnsureChunk(cp)
	for x := 7; x <= 9; x++ {
		for y := 69; y <= 72; y++ {
			for z := 7; z <= 10; z++ {
				w.SetBlock(x, y, z, core.Air)
			}
		}
	}
	w.SetBlock(8, 70, 8, core.Stone)
	w.SetBlock(8, 70, 9, core.Leaves)
	mesh, ok := w.Mesh(cp)
	if !ok {
		t.Fatal("mesh failed")
	}
	count := 0
	for i := 0; i+3 < len(mesh.Opaque.Vertices); i += 4 {
		all := true
		for j := 0; j < 4; j++ {
			v := mesh.Opaque.Vertices[i+j]
			if v.Z < 8.999 || v.Z > 9.001 || v.X < 8 || v.X > 9 || v.Y < 70 || v.Y > 71 {
				all = false
				break
			}
		}
		if all {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected the stone +Z face adjacent to leaves, got %d matching quads", count)
	}
}
func TestSkyAndTorchLight(t *testing.T) {
	w := New(123)
	if _, ok := w.EnsureChunk(ChunkPos{0, 0}); !ok {
		t.Fatal("chunk was not created")
	}
	w.SetBlock(8, 120, 8, core.Stone)
	if got := w.SkyLightAt(8, 119, 8); got != 0 {
		t.Fatalf("sky light under roof = %d, want 0", got)
	}
	if got := w.SkyLightAt(8, 121, 8); got != 15 {
		t.Fatalf("sky light above roof = %d, want 15", got)
	}
	if core.Torch.Def().Light != 14 || core.ItemForBlock(core.Torch) != core.ItemTorch {
		t.Fatal("torch is not registered as a 14-level light source item")
	}
	w.SetBlock(8, 110, 8, core.Torch)
	if got := w.BlockLightAt(8, 110, 8); got != 14 {
		t.Fatalf("torch light = %d, want 14", got)
	}
	if got := w.BlockLightAt(9, 110, 8); got != 13 {
		t.Fatalf("torch neighbor light = %d, want 13", got)
	}
	if got := w.BlockLightAt(10, 110, 8); got != 12 {
		t.Fatalf("torch distance light = %d, want 12", got)
	}
	mesh, ok := w.Mesh(ChunkPos{0, 0})
	if !ok {
		t.Fatal("missing mesh")
	}
	lit := false
	for _, v := range mesh.Opaque.Vertices {
		if v.Block > 0 {
			lit = true
			break
		}
	}
	if !lit {
		t.Fatal("torch light was not baked into the chunk mesh")
	}
}
func TestWaterFlowsAndRetracts(t *testing.T) {
	w := New(42)
	w.EnsureChunk(ChunkPos{0, 0})
	for x := 6; x <= 18; x++ {
		for z := 6; z <= 10; z++ {
			w.SetBlock(x, 69, z, core.Stone)
			for y := 70; y <= 73; y++ {
				w.SetBlock(x, y, z, core.Air)
			}
		}
	}
	w.SetBlock(8, 70, 8, core.Water)
	for i := 0; i < 40; i++ {
		w.TickFluids()
	}
	if got := w.GetBlock(9, 70, 8); got != core.Water || w.WaterLevelAt(9, 70, 8) != 1 {
		t.Fatalf("first flow cell = %v level %d", got, w.WaterLevelAt(9, 70, 8))
	}
	if got := w.WaterLevelAt(15, 70, 8); got != 7 {
		t.Fatalf("last flow level = %d, want 7", got)
	}
	if got := w.GetBlock(16, 70, 8); got != core.Air {
		t.Fatalf("water exceeded max range: %v", got)
	}

	w.SetBlock(8, 70, 8, core.Air)
	for i := 0; i < 60; i++ {
		w.TickFluids()
	}
	for x := 9; x <= 15; x++ {
		if got := w.GetBlock(x, 70, 8); got != core.Air {
			t.Fatalf("flow did not retract at %d: %v", x, got)
		}
	}
}
func TestSmoothLightingDefaultsOnAndMarksChunksDirty(t *testing.T) {
	w := New(7)
	c, _ := w.EnsureChunk(ChunkPos{0, 0})
	if !w.SmoothLighting {
		t.Fatal("smooth lighting should default on")
	}
	c.Dirty = false
	w.SetSmoothLighting(false)
	if w.SmoothLighting || !c.Dirty {
		t.Fatal("toggling smooth lighting did not mark the chunk dirty")
	}
}
func TestBlockLightCrossesChunkBoundaryAndDirtiesNeighbor(t *testing.T) {
	w := New(123)
	w.EnsureChunk(ChunkPos{0, 0})
	n, _ := w.EnsureChunk(ChunkPos{1, 0})
	w.SetBlock(14, 110, 8, core.Torch)
	if got := w.BlockLightAt(16, 110, 8); got != 12 {
		t.Fatalf("light across chunk boundary = %d, want 12", got)
	}
	w.ClearDirty(ChunkPos{1, 0})
	w.SetBlock(14, 110, 8, core.Air)
	if got := w.BlockLightAt(14, 110, 8); got != 0 {
		t.Fatalf("removed torch light = %d", got)
	}
	w.SetBlock(14, 110, 8, core.Torch)
	if got := w.BlockLightAt(14, 110, 8); got != 14 {
		t.Fatalf("torch light = %d", got)
	}
	if !n.Dirty {
		t.Fatal("neighbor chunk was not marked dirty after light changed")
	}
}
func TestDynamicLightFollowsAndCrossesChunks(t *testing.T) {
	w := New(123)
	w.EnsureChunk(ChunkPos{0, 0})
	n, _ := w.EnsureChunk(ChunkPos{1, 0})
	w.SetDynamicLight(14, 110, 8, 14)
	if got := w.BlockLightAt(14, 110, 8); got != 14 {
		t.Fatalf("dynamic source light = %d", got)
	}
	if got := w.BlockLightAt(16, 110, 8); got != 12 {
		t.Fatalf("dynamic cross-chunk light = %d", got)
	}
	if !n.Dirty {
		t.Fatal("neighbor chunk was not marked dirty by dynamic light")
	}
	w.SetBlock(2, 110, 30, core.Air)
	w.SetDynamicLight(2, 110, 30, 14)
	if got := w.BlockLightAt(14, 110, 8); got != 0 {
		t.Fatalf("old dynamic light = %d", got)
	}
	if got := w.BlockLightAt(2, 110, 30); got != 14 {
		t.Fatalf("new dynamic light = %d", got)
	}
	w.SetDynamicLight(2, 110, 30, 0)
	if got := w.BlockLightAt(2, 110, 30); got != 0 {
		t.Fatalf("cleared dynamic light = %d", got)
	}
}
func BenchmarkDynamicLightMove(b *testing.B) {
	w := New(123)
	for x := -2; x <= 2; x++ {
		for z := -2; z <= 2; z++ {
			w.EnsureChunk(ChunkPos{x, z})
		}
	}
	w.SetDynamicLight(8, 100, 8, 14)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.SetDynamicLight(8+i%8, 100, 8, 14)
	}
}
func TestDynamicLightBakedIntoMesh(t *testing.T) {
	w := New(123)
	w.EnsureChunk(ChunkPos{0, 0})
	w.SetBlock(8, 110, 8, core.Air)
	w.SetBlock(9, 110, 8, core.Stone)
	w.SetDynamicLight(8, 110, 8, 14)
	mesh, ok := w.Mesh(ChunkPos{0, 0})
	if !ok {
		t.Fatal("missing mesh")
	}
	for _, v := range mesh.Opaque.Vertices {
		if v.Block > 0 {
			return
		}
	}
	t.Fatal("dynamic light was not baked into the chunk mesh")
}
func TestDeferredDynamicLightForNewChunk(t *testing.T) {
	w := New(123)
	w.EnsureChunk(ChunkPos{0, 0})
	w.SetDynamicLight(14, 110, 8, 14)
	w.EnsureChunk(ChunkPos{1, 0})
	w.RefreshDynamicLightIfNeeded()
	if got := w.BlockLightAt(16, 110, 8); got != 12 {
		t.Fatalf("deferred dynamic light across chunk = %d", got)
	}
}
