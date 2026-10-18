package game

import (
	"reflect"
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/core"
	"sevg/internal/ui"
	"sevg/internal/world"
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
func TestThirdPersonPlayerHeldIcon(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.camera = cameraThird
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemStonePickaxe, Count: 1}
	g.canv = &ui.Canvas{
		Icons: map[core.ItemID]ui.Glyph{
			core.ItemStonePickaxe: {U0: 0.1, V0: 0.1, U1: 0.2, V1: 0.2},
		},
	}
	opaque, _, icons := g.buildEntityMeshes()
	if len(opaque.Vertices) == 0 {
		t.Fatal("player model is missing")
	}
	if len(icons.Vertices) == 0 || len(icons.Indices) == 0 {
		t.Fatal("held texture item has no icon-card mesh")
	}
}
func TestHeldBlockIgnoresHeadPitch(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 1}
	g.player.Pitch = -0.8
	var a, ta world.MeshData
	g.appendPlayerHeldBlock(&a, &ta, 0)
	g.player.Pitch = 0.8
	var b, tb world.MeshData
	g.appendPlayerHeldBlock(&b, &tb, 0)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("held block changed with head pitch")
	}
}
func TestPlayerSkinUsesTextureTile(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.camera = cameraThird
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{}
	opaque, _, _ := g.buildEntityMeshes()
	if len(opaque.Vertices) == 0 {
		t.Fatal("player model is empty")
	}
	u0, v0, u1, v1 := world.TileUV(core.TilePlayerSkinMap)
	for _, v := range opaque.Vertices {
		if v.U < u0 || v.U > u1 || v.V < v0 || v.V > v1 {
			t.Fatalf("vertex UV outside skin tile: %f,%f", v.U, v.V)
		}
	}
}
func TestActionSwingAnimation(t *testing.T) {
	g := newTestGame(t)
	g.triggerSwing()
	g.swingT = 0.15
	if a := g.swingAngle(); a < 0.8 {
		t.Fatalf("mid-swing angle = %f", a)
	}
	g.swingT = 0
	if a := g.swingAngle(); a != 0 {
		t.Fatalf("finished swing angle = %f", a)
	}
}
func TestCrouchLowersThirdPersonModel(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.camera = cameraThird
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{}
	standing, _, _ := g.buildEntityMeshes()
	maxStand := float32(-100)
	for _, v := range standing.Vertices {
		if v.Y > maxStand {
			maxStand = v.Y
		}
	}
	g.player.Crouching = true
	g.player.Body.Height = 1.5
	crouched, _, _ := g.buildEntityMeshes()
	maxCrouch := float32(-100)
	for _, v := range crouched.Vertices {
		if v.Y > maxCrouch {
			maxCrouch = v.Y
		}
	}
	if maxCrouch >= maxStand-0.2 {
		t.Fatalf("crouched model top %f vs standing %f", maxCrouch, maxStand)
	}
}
func TestStandardSkinFaceLayout(t *testing.T) {
	if got, want := skinHead[5], skinFace(8, 8, 16, 16); got != want {
		t.Fatalf("head front UV = %v, want %v", got, want)
	}
	if got, want := skinTorso[4], skinFace(32, 20, 40, 32); got != want {
		t.Fatalf("torso back UV = %v, want %v", got, want)
	}
	if skinArmLeft[5][0] <= skinArmLeft[5][2] {
		t.Fatal("left arm front UV is not mirrored")
	}
	if skinLegLeft[5][0] <= skinLegLeft[5][2] {
		t.Fatal("left leg front UV is not mirrored")
	}
	for _, set := range []skinFaces{skinHead, skinTorso, skinArmRight, skinArmLeft, skinLegRight, skinLegLeft} {
		for _, r := range set {
			if r[0] < 0 || r[1] < 0 || r[2] > 1 || r[3] > 1 {
				t.Fatalf("skin face outside texture: %v", r)
			}
		}
	}
}
