package player

import (
	"math"
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func flatWorld(t *testing.T, radius int) *world.World {
	t.Helper()
	w := world.New(1)
	w.EnsureChunk(world.ChunkPos{0, 0})
	for x := -radius; x <= radius; x++ {
		for z := -radius; z <= radius; z++ {
			w.SetBlock(x, 64, z, core.Stone)
		}
	}
	return w
}

func TestCrouchLowersEyeAndSlowsMovement(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 65, Z: 0.5})
	p.Update(0.016, w, Input{})
	standEye := p.Eye().Y
	p.Update(0.016, w, Input{Crouch: true})
	if !p.Crouching || p.Body.Height != 1.5 {
		t.Fatalf("crouch=%v height=%f", p.Crouching, p.Body.Height)
	}
	if p.Eye().Y >= standEye {
		t.Fatalf("crouched eye %f did not lower from %f", p.Eye().Y, standEye)
	}
	for i := 0; i < 30; i++ {
		p.Update(0.016, w, Input{Forward: 1, Crouch: true})
	}
	if sp := float32(math.Hypot(float64(p.Body.Vel.X), float64(p.Body.Vel.Z))); sp > 2.0 {
		t.Fatalf("crouched speed = %f", sp)
	}
	p.Update(0.016, w, Input{})
	if p.Crouching || p.Body.Height != 1.8 {
		t.Fatalf("failed to stand: crouch=%v height=%f", p.Crouching, p.Body.Height)
	}
}

func TestCrouchStopsAtBlockEdge(t *testing.T) {
	w := world.New(2)
	w.EnsureChunk(world.ChunkPos{0, 0})
	w.SetBlock(0, 64, 0, core.Stone)
	p := New(mathx.Vec3{X: 0.5, Y: 65, Z: 0.5})
	p.Yaw = 1.5707964 // +X
	for i := 0; i < 120; i++ {
		p.Update(0.016, w, Input{Forward: 1, Crouch: true})
	}
	if p.Body.Pos.Y < 64.9 {
		t.Fatalf("crouching player fell off the edge: y=%f x=%f", p.Body.Pos.Y, p.Body.Pos.X)
	}
	if p.Body.Pos.X <= 1.0 || p.Body.Pos.X > 1.31 {
		t.Fatalf("crouching player should overhang past the edge but not fall: x=%f", p.Body.Pos.X)
	}

	// Moving parallel to the edge must still work while overhanging.
	z0 := p.Body.Pos.Z
	for i := 0; i < 60; i++ {
		p.Update(0.016, w, Input{Strafe: 1, Crouch: true})
	}
	if p.Body.Pos.Z <= z0+0.15 || p.Body.Pos.Y < 64.9 {
		t.Fatalf("stuck or fell while moving along edge: z=%f y=%f", p.Body.Pos.Z, p.Body.Pos.Y)
	}

	// Moving back towards the platform must not be blocked at the edge.
	x0 := p.Body.Pos.X
	for i := 0; i < 60; i++ {
		p.Update(0.016, w, Input{Forward: -1, Crouch: true})
	}
	if p.Body.Pos.X >= x0-0.1 || p.Body.Pos.Y < 64.9 {
		t.Fatalf("stuck or fell while retreating: x=%f y=%f", p.Body.Pos.X, p.Body.Pos.Y)
	}
}
