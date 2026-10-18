package game

import (
	"testing"

	"sevg/internal/core"
)

func TestThirdPersonPlayerModel(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.thirdPerson = true
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
	g.thirdPerson = false
	if len(g.buildHeldMesh().Vertices) == 0 {
		t.Fatal("first-person held item is missing")
	}
	g.thirdPerson = true
	if len(g.buildHeldMesh().Vertices) != 0 || len(g.buildHeldIconMesh().Vertices) != 0 {
		t.Fatal("first-person held item should be hidden in third person")
	}
}
