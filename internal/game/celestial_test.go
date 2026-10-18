package game

import (
	"math"
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/render"
	"sevg/internal/world"
)

func TestSunAndMoonCycle(t *testing.T) {
	if y := sunDirection(0.25).Y; math.Abs(float64(y)) > 0.01 {
		t.Fatalf("sunrise sun height = %f", y)
	}
	if y := sunDirection(0.5).Y; y < 0.95 {
		t.Fatalf("noon sun height = %f", y)
	}
	if y := sunDirection(0.0).Y; y > -0.95 {
		t.Fatalf("midnight sun height = %f", y)
	}
	sun := sunDirection(0.6)
	moon := moonDirection(0.6)
	if math.Abs(float64(sun.X+moon.X)) > 0.001 || math.Abs(float64(sun.Y+moon.Y)) > 0.001 {
		t.Fatalf("moon is not opposite sun: sun=%+v moon=%+v", sun, moon)
	}
}

func TestSkyMeshContainsBodies(t *testing.T) {
	g := newTestGame(t)
	cam := render.Camera{Pos: mathx.V3(0, 70, 0), Dir: mathx.V3(0, 0, -1), Up: mathx.V3(0, 1, 0), Fov: mathx.Radians(72)}
	g.time = 0.5
	noon := g.buildSkyMesh(cam)
	if len(noon.Vertices) != 4 || len(noon.Indices) != 6 {
		t.Fatalf("noon sky mesh = %d/%d, want sun only", len(noon.Vertices), len(noon.Indices))
	}
	u0, v0, u1, v1 := world.TileUV(core.TileSun)
	for _, v := range noon.Vertices {
		if v.U < u0 || v.U > u1 || v.V < v0 || v.V > v1 {
			t.Fatalf("sun UV outside tile: %f,%f", v.U, v.V)
		}
	}
	g.time = 0.0
	midnight := g.buildSkyMesh(cam)
	if len(midnight.Vertices) != 4 || len(midnight.Indices) != 6 {
		t.Fatalf("midnight sky mesh = %d/%d, want moon only", len(midnight.Vertices), len(midnight.Indices))
	}
}
