package game

import (
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/core"
	"sevg/internal/render"
)

func TestInventoryQDropHeldStack(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.heldStack = core.Stack{Item: core.ItemPlanks, Count: 3}
	before := len(g.mobs.Items)
	g.onKey(sdl.K_Q)
	if g.heldStack.Count != 2 {
		t.Fatalf("held count = %d", g.heldStack.Count)
	}
	if len(g.mobs.Items) != before+1 || g.mobs.Items[before].Count != 1 {
		t.Fatalf("dropped items = %+v", g.mobs.Items[before:])
	}
}

func TestInventoryShiftQDropHeldStack(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.heldStack = core.Stack{Item: core.ItemPlanks, Count: 3}
	g.shiftDown = true
	before := len(g.mobs.Items)
	g.onKey(sdl.K_Q)
	if !g.heldStack.Empty() {
		t.Fatalf("held stack = %+v", g.heldStack)
	}
	if len(g.mobs.Items) != before+1 || g.mobs.Items[before].Count != 3 {
		t.Fatalf("dropped items = %+v", g.mobs.Items[before:])
	}
}

func TestInventoryQDropHoveredSlot(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.inv.Slots[9] = core.Stack{Item: core.ItemCobble, Count: 5}
	r := g.invSlotRect(0)
	g.mouseX = r.X + r.W/2
	g.mouseY = r.Y + r.H/2
	before := len(g.mobs.Items)
	g.dropFromInventory()
	if g.inv.Slots[9].Count != 4 {
		t.Fatalf("slot count = %d", g.inv.Slots[9].Count)
	}
	if len(g.mobs.Items) != before+1 || g.mobs.Items[before].Item != core.ItemCobble {
		t.Fatalf("dropped items = %+v", g.mobs.Items[before:])
	}
}

func TestInventoryClickOutsideDropsHeldStack(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.heldStack = core.Stack{Item: core.ItemPlanks, Count: 4}
	p := g.invPanelRect()
	g.mouseX = p.X - 10
	g.mouseY = p.Y + 10
	before := len(g.mobs.Items)
	g.handleUIClick(false)
	if !g.heldStack.Empty() {
		t.Fatalf("held stack = %+v", g.heldStack)
	}
	if len(g.mobs.Items) != before+1 || g.mobs.Items[before].Count != 4 {
		t.Fatalf("dropped items = %+v", g.mobs.Items[before:])
	}
}
func TestGameplayShiftQDropWholeStack(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 5}
	g.shiftDown = true
	before := len(g.mobs.Items)
	g.onKey(sdl.K_Q)
	if !g.inv.Slots[0].Empty() {
		t.Fatalf("selected slot = %+v", g.inv.Slots[0])
	}
	if len(g.mobs.Items) != before+1 || g.mobs.Items[before].Count != 5 {
		t.Fatalf("dropped items = %+v", g.mobs.Items[before:])
	}
}

func TestCreativeCatalogQDrop(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.mode = ModeCreative
	g.inventoryTab = 1
	r := g.creativeItemRect(int(core.ItemPlanks) - 1)
	g.mouseX = r.X + r.W/2
	g.mouseY = r.Y + r.H/2
	before := len(g.mobs.Items)
	g.onKey(sdl.K_Q)
	if len(g.mobs.Items) != before+1 || g.mobs.Items[before].Item != core.ItemPlanks || g.mobs.Items[before].Count != 1 {
		t.Fatalf("catalog Q drop = %+v", g.mobs.Items[before:])
	}
	if !g.heldStack.Empty() {
		t.Fatalf("catalog Q should not fill cursor: %+v", g.heldStack)
	}
	g.shiftDown = true
	g.onKey(sdl.K_Q)
	// The full stack drops next to the single plank and the two merge into a
	// 64-item stack plus a one-item remainder.
	total, maxStack := 0, 0
	for _, d := range g.mobs.Items[before:] {
		if d.Item != core.ItemPlanks {
			continue
		}
		total += d.Count
		if d.Count > maxStack {
			maxStack = d.Count
		}
	}
	if total != 65 || maxStack != 64 {
		t.Fatalf("catalog shift Q stacks: total=%d max=%d", total, maxStack)
	}
}
