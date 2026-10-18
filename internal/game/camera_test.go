package game

import (
	"math"
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
)

func TestCameraCollisionClamp(t *testing.T) {
	g := newTestGame(t)
	g.world.SetBlock(8, 90, 10, core.Stone)
	origin := mathx.Vec3{X: 8.5, Y: 90.5, Z: 8.5}
	pos := g.clampCamera(origin, mathx.Vec3{Z: 3.6})
	if pos.Z >= 10 {
		t.Fatalf("camera entered wall: %+v", pos)
	}
	if pos.Z <= origin.Z {
		t.Fatalf("camera was not moved along ray: %+v", pos)
	}
	free := g.clampCamera(origin, mathx.Vec3{Z: -3.6})
	if math.Abs(float64(free.Z-(origin.Z-3.6))) > 0.01 {
		t.Fatalf("unobstructed camera was clamped: %+v", free)
	}
}
