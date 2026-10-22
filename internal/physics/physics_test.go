package physics

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func TestBodyLandsOnBlock(t *testing.T) {
	w := world.New(1)
	w.EnsureChunk(world.ChunkPos{})
	w.SetBlock(0, 100, 0, core.Stone)
	b := Body{Pos: mathx.Vec3{X: 0.5, Y: 106, Z: 0.5}, Width: 0.6, Height: 1.8}
	for i := 0; i < 120; i++ {
		b.Vel.Y -= 0.2
		b.Move(w, mathx.Vec3{Y: b.Vel.Y})
	}
	if b.Pos.Y < 100.9 || b.Pos.Y > 101.1 {
		t.Fatalf("landed at %f, want 101", b.Pos.Y)
	}
	if !b.Grounded(w) {
		t.Fatal("body should be grounded")
	}
}
