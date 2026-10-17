package game

import (
	"path/filepath"
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	g := newTestGame(t)
	g.saveAt = filepath.Join(t.TempDir(), "save.gob")
	g.world.SetBlock(8, 90, 8, core.Brick)
	g.player.Body.Pos = mathx.Vec3{X: 3.25, Y: 70.5, Z: 4.75}
	g.player.Yaw = 1.25
	g.player.Pitch = -0.3
	g.inv.Slots[0] = core.Stack{Item: core.ItemObsidian, Count: 7}
	if err := g.saveGame(); err != nil {
		t.Fatal(err)
	}
	sv, err := loadSave(g.saveAt)
	if err != nil {
		t.Fatal(err)
	}
	if sv.Seed != 123 {
		t.Fatalf("seed %d", sv.Seed)
	}
	if len(sv.Edits) != 1 {
		t.Fatalf("edits %d", len(sv.Edits))
	}
	if sv.Player.X != 3.25 || sv.Player.Z != 4.75 {
		t.Fatalf("player %+v", sv.Player)
	}
	if sv.Inventory.Count(core.ItemObsidian) != 7 {
		t.Fatalf("inventory not saved")
	}
}

func TestCraftPlanks(t *testing.T) {
	g := newTestGame(t)
	g.inv = core.Inventory{}
	g.inv.Add(core.ItemLog, 2)
	g.craft(0)
	if got := g.inv.Count(core.ItemPlanks); got != 4 {
		t.Fatalf("planks %d", got)
	}
	if got := g.inv.Count(core.ItemLog); got != 1 {
		t.Fatalf("logs %d", got)
	}
}
