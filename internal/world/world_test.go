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
