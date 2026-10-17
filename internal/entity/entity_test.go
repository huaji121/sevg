package entity

import (
	"testing"

	"sevg/internal/mathx"
	"sevg/internal/world"
)

func TestPickMob(t *testing.T) {
	m := NewManager(1)
	mob := m.SpawnAt(Pig, mathx.Vec3{X: 0.5, Y: 64, Z: -5.5})
	mob.Body.Width = 0.7
	hit, dist, ok := m.PickMob(mathx.Vec3{X: 0.5, Y: 64.5, Z: 0.5}, mathx.Vec3{Z: -1}, 10)
	if !ok || hit != mob {
		t.Fatalf("mob not picked: %v %f %v", hit, dist, ok)
	}
	if dist < 4 || dist > 6 {
		t.Fatalf("distance %f", dist)
	}
}

func TestRayAABBMiss(t *testing.T) {
	if _, ok := world.RayAABB(mathx.Vec3{X: 0, Y: 0, Z: 0}, mathx.Vec3{Z: -1}, mathx.Vec3{X: 2, Y: -1, Z: -1}, mathx.Vec3{X: 3, Y: 1, Z: -2}); ok {
		t.Fatal("expected miss")
	}
}
