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

func TestDropsStackOnSpawn(t *testing.T) {
	m := NewManager(1)
	pos := mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}
	for i := 0; i < 5; i++ {
		m.Drop(core.ItemCobble, 1, pos, mathx.Vec3{})
	}
	if m.ItemCount() != 1 {
		t.Fatalf("same-type drops did not stack: %d entities", m.ItemCount())
	}
	if m.Items[0].Count != 5 {
		t.Fatalf("stack count = %d, want 5", m.Items[0].Count)
	}
}

func TestDropsRespectStackLimit(t *testing.T) {
	m := NewManager(1)
	pos := mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}
	m.Drop(core.ItemCobble, 40, pos, mathx.Vec3{})
	m.Drop(core.ItemCobble, 40, pos, mathx.Vec3{})
	if m.ItemCount() != 2 {
		t.Fatalf("overflow should create a second stack, got %d", m.ItemCount())
	}
	if got := m.Items[0].Count; got != core.StackLimit {
		t.Fatalf("first stack = %d, want %d", got, core.StackLimit)
	}
	if total := m.Items[0].Count + m.Items[1].Count; total != 80 {
		t.Fatalf("total count = %d, want 80", total)
	}
}

func TestUnstackableToolsDoNotMerge(t *testing.T) {
	m := NewManager(1)
	pos := mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}
	m.Drop(core.ItemIronSword, 1, pos, mathx.Vec3{})
	m.Drop(core.ItemIronSword, 1, pos, mathx.Vec3{})
	if m.ItemCount() != 2 {
		t.Fatalf("unstackable tools merged: %d entities", m.ItemCount())
	}
}

func TestProjectileHitsGround(t *testing.T) {
	m := NewManager(1)
	w := world.New(1)
	w.EnsureChunk(world.ChunkPos{})
	w.SetBlock(8, 100, 8, core.Stone)
	m.SpawnProjectile(core.ItemSnowball, mathx.Vec3{X: 8.5, Y: 106, Z: 8.5}, mathx.Vec3{Y: -5})
	for i := 0; i < 300 && len(m.Projectiles) > 0; i++ {
		m.UpdateProjectiles(w, 0.05)
	}
	if len(m.Projectiles) != 0 {
		t.Fatal("projectile never landed")
	}
	if len(m.ProjectileImpacts) == 0 {
		t.Fatal("landing did not record an impact")
	}
	for _, pos := range m.ProjectileImpacts {
		if pos.Y < 100 || pos.Y > 102 {
			t.Fatalf("impact at unexpected height: %f", pos.Y)
		}
	}
}

func TestDropsMergeWhenTheySettleTogether(t *testing.T) {
	m := NewManager(1)
	w := world.New(1)
	w.EnsureChunk(world.ChunkPos{})
	var inv core.Inventory
	player := mathx.Vec3{X: 100, Y: 64, Z: 100}
	m.Drop(core.ItemCobble, 3, mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}, mathx.Vec3{})
	m.Drop(core.ItemCobble, 4, mathx.Vec3{X: 5.5, Y: 64, Z: 0.5}, mathx.Vec3{})
	if m.ItemCount() != 2 {
		t.Fatalf("far-apart drops merged on spawn: %d", m.ItemCount())
	}
	// Nudge the second drop next to the first; the per-tick merge should fold
	// them into one seven-item stack.
	m.Items[1].Body.Pos = mathx.Vec3{X: 0.7, Y: 64, Z: 0.5}
	m.Update(0.05, w, player, &inv)
	if m.ItemCount() != 1 {
		t.Fatalf("nearby drops did not merge: %d", m.ItemCount())
	}
	if m.Items[0].Count != 7 {
		t.Fatalf("merged count = %d, want 7", m.Items[0].Count)
	}
}
