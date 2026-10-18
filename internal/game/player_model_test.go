package game

import (
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/core"
)

func TestThirdPersonPlayerModel(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.camera = cameraThird
	g.player.Yaw = 0
	opaque, _, _ := g.buildEntityMeshes()
	if len(opaque.Vertices) == 0 || len(opaque.Indices) == 0 {
		t.Fatal("third-person player model produced an empty mesh")
	}
}

func TestThirdPersonHidesFirstPersonHeldItem(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 8}
	g.camera = cameraFirst
	if len(g.buildHeldMesh().Vertices) == 0 {
		t.Fatal("first-person held item is missing")
	}
	g.camera = cameraThird
	if len(g.buildHeldMesh().Vertices) != 0 || len(g.buildHeldIconMesh().Vertices) != 0 {
		t.Fatal("first-person held item should be hidden in third person")
	}
}
func TestCameraModeCycle(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	if g.camera != cameraFirst {
		t.Fatalf("default camera = %d", g.camera)
	}
	g.onKey(sdl.K_F5)
	if g.camera != cameraThird {
		t.Fatalf("after first F5 camera = %d", g.camera)
	}
	g.onKey(sdl.K_F5)
	if g.camera != cameraSecond {
		t.Fatalf("after second F5 camera = %d", g.camera)
	}
	g.onKey(sdl.K_F5)
	if g.camera != cameraFirst {
		t.Fatalf("after third F5 camera = %d", g.camera)
	}
}

func TestSecondPersonPlayerModel(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.camera = cameraSecond
	opaque, _, _ := g.buildEntityMeshes()
	if len(opaque.Vertices) == 0 {
		t.Fatal("second-person player model produced an empty mesh")
	}
}
