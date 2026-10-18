package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/render"
)

func TestChestClickTransfer(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StatePlaying
	pos := [3]int{8, 100, 8}
	g.world.SetBlock(8, 100, 8, core.Chest)
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemDirt, Count: 5}
	g.openChest(pos)
	if g.state != StateChest {
		t.Fatalf("state = %v", g.state)
	}
	l := g.chestLayout()
	g.mouseX = l.hotbar[0].X + l.hotbar[0].W/2
	g.mouseY = l.hotbar[0].Y + l.hotbar[0].H/2
	g.clickChest(false)
	if g.heldStack.Item != core.ItemDirt || g.heldStack.Count != 5 {
		t.Fatalf("cursor stack = %+v", g.heldStack)
	}
	g.mouseX = l.chest[0].X + l.chest[0].W/2
	g.mouseY = l.chest[0].Y + l.chest[0].H/2
	g.clickChest(false)
	if got := (*g.chestAt(pos))[0]; got.Item != core.ItemDirt || got.Count != 5 {
		t.Fatalf("chest slot = %+v", got)
	}
}

func TestChestContentsDropWhenRemoved(t *testing.T) {
	g := newTestGame(t)
	pos := [3]int{8, 100, 8}
	slots := g.chestAt(pos)
	slots[0] = core.Stack{Item: core.ItemDirt, Count: 5}
	g.dropChest(pos)
	if g.mobs.ItemCount() != 1 {
		t.Fatalf("drops = %d", g.mobs.ItemCount())
	}
	if d := g.mobs.Items[0]; d.Item != core.ItemDirt || d.Count != 5 {
		t.Fatalf("drop = %+v", d)
	}
	if _, ok := g.chests[pos]; ok {
		t.Fatal("chest data remained after removal")
	}
}

func TestChestCloseReturnsCursorStack(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StatePlaying
	g.inv = core.Inventory{}
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemStone, Count: 3}
	g.openChest([3]int{1, 2, 3})
	l := g.chestLayout()
	g.mouseX = l.hotbar[0].X + 2
	g.mouseY = l.hotbar[0].Y + 2
	g.clickChest(false)
	if g.heldStack.Empty() {
		t.Fatal("cursor did not pick up stack")
	}
	g.setState(StatePlaying)
	if !g.heldStack.Empty() {
		t.Fatal("cursor stack was not returned on close")
	}
	if got := g.inv.Count(core.ItemStone); got != 3 {
		t.Fatalf("stone returned = %d", got)
	}
}

func TestChestSaveDataCopiesSlots(t *testing.T) {
	g := newTestGame(t)
	pos := [3]int{4, 5, 6}
	(*g.chestAt(pos))[3] = core.Stack{Item: core.ItemApple, Count: 2}
	data := g.chestSaveData()
	if got := data[pos][3]; got.Item != core.ItemApple || got.Count != 2 {
		t.Fatalf("saved chest slot = %+v", got)
	}
}
func TestChestContentsPersistInSave(t *testing.T) {
	g := newTestGame(t)
	g.saveAt = t.TempDir() + "/save.gob"
	pos := [3]int{2, 70, 3}
	(*g.chestAt(pos))[5] = core.Stack{Item: core.ItemIronIngot, Count: 7}
	if err := g.saveGame(); err != nil {
		t.Fatal(err)
	}
	sv, err := loadSave(g.saveAt)
	if err != nil {
		t.Fatal(err)
	}
	if got := sv.Chests[pos][5]; got.Item != core.ItemIronIngot || got.Count != 7 {
		t.Fatalf("saved chest = %+v", got)
	}
}
