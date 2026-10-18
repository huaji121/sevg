package game

import (
	"testing"

	"sevg/internal/mathx"
)

func TestInventoryDoesNotPauseWorld(t *testing.T) {
	g := newTestGame(t)
	g.state = StateInventory
	g.player.Body.Pos = mathx.Vec3{X: 8.5, Y: 90, Z: 8.5}
	g.player.Body.Vel = mathx.Vec3{}
	y0 := g.player.Body.Pos.Y
	t0 := g.time
	for i := 0; i < 10; i++ {
		g.updateWorld(0.05, false)
	}
	if g.player.Body.Pos.Y >= y0 {
		t.Fatalf("player did not fall while inventory was open: %f -> %f", y0, g.player.Body.Pos.Y)
	}
	if g.time == t0 {
		t.Fatal("world time did not advance while inventory was open")
	}
}
