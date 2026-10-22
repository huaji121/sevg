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
			w.SetBlock(x, 100, z, core.Stone)
		}
	}
	return w
}

func TestCrouchLowersEyeAndSlowsMovement(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
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
	w.SetBlock(0, 100, 0, core.Stone)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Yaw = 1.5707964 // +X
	for i := 0; i < 120; i++ {
		p.Update(0.016, w, Input{Forward: 1, Crouch: true})
	}
	if p.Body.Pos.Y < 100.9 {
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
	if p.Body.Pos.Z <= z0+0.15 || p.Body.Pos.Y < 100.9 {
		t.Fatalf("stuck or fell while moving along edge: z=%f y=%f", p.Body.Pos.Z, p.Body.Pos.Y)
	}

	// Moving back towards the platform must not be blocked at the edge.
	x0 := p.Body.Pos.X
	for i := 0; i < 60; i++ {
		p.Update(0.016, w, Input{Forward: -1, Crouch: true})
	}
	if p.Body.Pos.X >= x0-0.1 || p.Body.Pos.Y < 100.9 {
		t.Fatalf("stuck or fell while retreating: x=%f y=%f", p.Body.Pos.X, p.Body.Pos.Y)
	}
}
func TestJumpClearsOnGroundImmediately(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Update(0.016, w, Input{})
	p.Update(0.016, w, Input{Jump: true})
	if p.Body.OnGround || p.Body.Vel.Y <= 0 {
		t.Fatalf("jump state: onGround=%v vy=%f", p.Body.OnGround, p.Body.Vel.Y)
	}
}

func TestEatRestoresFoodAndSaturation(t *testing.T) {
	p := New(mathx.Vec3{})
	p.Food = 10
	p.Saturation = 0
	p.Eat(4, 2.4)
	if p.Food != 14 {
		t.Fatalf("food = %f, want 14", p.Food)
	}
	if diff := p.Saturation - 2.4; diff > 0.001 || diff < -0.001 {
		t.Fatalf("saturation = %f, want 2.4", p.Saturation)
	}
	p.Eat(100, 100)
	if p.Food != p.MaxFood || p.Saturation != p.Food {
		t.Fatalf("eating clamped wrong: food=%f sat=%f", p.Food, p.Saturation)
	}
}

func TestSprintRequiresFood(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Food = 6
	p.Update(0.05, w, Input{Forward: 1, Sprint: true})
	if p.Sprinting {
		t.Fatal("sprinted while out of food")
	}
	p.Food = 20
	p.Update(0.05, w, Input{Forward: 1, Sprint: true})
	if !p.Sprinting {
		t.Fatal("could not sprint while fed")
	}
}

func TestExhaustionDrainsFood(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Saturation = 0
	p.AddExhaustion(4)
	p.Update(0.016, w, Input{})
	if p.Food != p.MaxFood-1 {
		t.Fatalf("food = %f, want %f", p.Food, p.MaxFood-1)
	}
}

func TestFoodRegeneratesHealth(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Health = 10
	p.Saturation = 0
	for i := 0; i < 300; i++ { // 15 seconds
		p.Update(0.05, w, Input{})
	}
	if p.Health <= 10 {
		t.Fatalf("health did not regenerate: %f", p.Health)
	}
	if p.Food >= p.MaxFood {
		t.Fatalf("regeneration did not consume food: %f", p.Food)
	}
}

func TestStarvingPlayerTakesDamage(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Food = 0
	p.Saturation = 0
	for i := 0; i < 100; i++ { // 5 seconds
		p.Update(0.05, w, Input{})
	}
	if p.Health >= p.MaxHP {
		t.Fatalf("starvation did not damage: %f", p.Health)
	}
}

func TestCreativeHungerStaysFull(t *testing.T) {
	w := flatWorld(t, 3)
	p := New(mathx.Vec3{X: 0.5, Y: 101, Z: 0.5})
	p.Creative = true
	p.Food = 3
	p.Saturation = 0
	p.AddExhaustion(20)
	p.Update(0.05, w, Input{Forward: 1, Sprint: true})
	if p.Food != p.MaxFood {
		t.Fatalf("creative hunger drained: %f", p.Food)
	}
}
