package entity

import (
	"testing"

	"sevg/internal/core"

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
func TestBlockDropsUseBlockTextures(t *testing.T) {
	m := NewManager(1)
	m.Drop(core.ItemPlanks, 1, mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}, mathx.Vec3{})
	boxes := m.Items[0].Boxes()
	if len(boxes) != 1 || boxes[0].Block != core.Planks {
		t.Fatalf("plank drop block = %v", boxes)
	}

	m.Drop(core.ItemApple, 1, mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}, mathx.Vec3{})
	boxes = m.Items[1].Boxes()
	if len(boxes) != 1 || boxes[0].Block != core.Air {
		t.Fatalf("apple drop should stay a palette cube: %v", boxes)
	}
}
func TestDropVisuals(t *testing.T) {
	cases := []struct {
		item core.ItemID
		want DropVisual
	}{
		{core.ItemGlass, DropVisualGlass},
		{core.ItemIce, DropVisualGlass},
		{core.ItemTallGrass, DropVisualPlant},
		{core.ItemFlower, DropVisualPlant},
		{core.ItemStonePickaxe, DropVisualPickaxe},
		{core.ItemIronPickaxe, DropVisualPickaxe},
		{core.ItemWoodSword, DropVisualSword},
		{core.ItemIronSword, DropVisualSword},
		{core.ItemPlanks, DropVisualBlock},
		{core.ItemApple, DropVisualCube},
		{core.ItemCoal, DropVisualCube},
	}
	for _, tc := range cases {
		d := &ItemDrop{Item: tc.item}
		if got := d.Visual(); got != tc.want {
			t.Fatalf("%v visual = %v, want %v", tc.item, got, tc.want)
		}
	}
}
func TestPlayerDropPickupDelay(t *testing.T) {
	m := NewManager(1)
	w := world.New(1)
	w.EnsureChunk(world.ChunkPos{})
	w.SetBlock(8, 69, 8, core.Stone)
	pos := mathx.Vec3{X: 8.5, Y: 70, Z: 8.5}
	var inv core.Inventory
	m.DropPlayer(core.ItemPlanks, 1, pos, mathx.Vec3{})
	if m.Items[0].PickupDelay < 1 {
		t.Fatalf("player drop pickup delay = %f", m.Items[0].PickupDelay)
	}
	for i := 0; i < 12; i++ {
		m.Update(0.05, w, pos, &inv)
	}
	if m.ItemCount() != 1 || inv.Count(core.ItemPlanks) != 0 {
		t.Fatalf("player drop picked up too early: items=%d inv=%d", m.ItemCount(), inv.Count(core.ItemPlanks))
	}
	for i := 0; i < 30; i++ {
		m.Update(0.05, w, pos, &inv)
	}
	if m.ItemCount() != 0 || inv.Count(core.ItemPlanks) != 1 {
		t.Fatalf("player drop was not picked up: items=%d inv=%d", m.ItemCount(), inv.Count(core.ItemPlanks))
	}
}
