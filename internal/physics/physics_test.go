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
	w.SetBlock(0, 64, 0, core.Stone)
	b := Body{Pos: mathx.Vec3{X: 0.5, Y: 70, Z: 0.5}, Width: 0.6, Height: 1.8}
	for i := 0; i < 120; i++ {
		b.Vel.Y -= 0.2
		b.Move(w, mathx.Vec3{Y: b.Vel.Y})
	}
	if b.Pos.Y < 64.9 || b.Pos.Y > 65.1 {
		t.Fatalf("landed at %f, want 65", b.Pos.Y)
	}
	if !b.Grounded(w) {
		t.Fatal("body should be grounded")
	}
}
