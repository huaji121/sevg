package game

import (
	"testing"

	"sevg/internal/core"
)

func TestSandFallsAndLands(t *testing.T) {
	g := newTestGame(t)
	g.world.SetBlock(8, 68, 8, core.Stone)
	g.world.SetBlock(8, 69, 8, core.Air)
	g.world.SetBlock(8, 70, 8, core.Sand)
	g.mobs.StartFalling(g.world, 8, 70, 8)
	if b := g.world.GetBlock(8, 70, 8); b != core.Air {
		t.Fatalf("sand block was not removed: %v", b)
	}
	if len(g.mobs.Falling) != 1 {
		t.Fatalf("falling entities = %d", len(g.mobs.Falling))
	}
	for i := 0; i < 100 && len(g.mobs.Falling) > 0; i++ {
		g.mobs.UpdateFalling(g.world, 0.05)
	}
	if len(g.mobs.Falling) != 0 {
		t.Fatal("falling sand did not land")
	}
	if b := g.world.GetBlock(8, 69, 8); b != core.Sand {
		t.Fatalf("landed block = %v", b)
	}
}
