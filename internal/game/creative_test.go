package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/render"
)

func TestCreativeInventoryClick(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.mode = ModeCreative
	g.inventoryTab = 1
	g.inv.Selected = 2
	before := g.inv.Slots[2]

	first := g.creativeItemRect(0)
	g.mouseX = first.X + first.W/2
	g.mouseY = first.Y + first.H/2
	g.handleUIClick(false)
	if g.heldStack.Item != core.ItemID(1) || g.heldStack.Count != 1 {
		t.Fatalf("creative cursor stack = %+v", g.heldStack)
	}
	if g.inv.Slots[2] != before {
		t.Fatalf("creative pick modified the slot before placement")
	}

	hotbar := g.creativeHotbarRect(7)
	g.mouseX = hotbar.X + hotbar.W/2
	g.mouseY = hotbar.Y + hotbar.H/2
	g.handleUIClick(false)
	if g.inv.Slots[7].Item != core.ItemID(1) || !g.heldStack.Empty() {
		t.Fatalf("creative placement: slot=%+v held=%+v", g.inv.Slots[7], g.heldStack)
	}

	// Picking another item and clicking an occupied slot must use the same
	// swap semantics as the survival inventory.
	second := g.creativeItemRect(1)
	g.mouseX = second.X + second.W/2
	g.mouseY = second.Y + second.H/2
	g.handleUIClick(false)
	occupied := g.creativeHotbarRect(4)
	g.mouseX = occupied.X + occupied.W/2
	g.mouseY = occupied.Y + occupied.H/2
	g.handleUIClick(false)
	if g.inv.Slots[4].Item != core.ItemID(2) || g.heldStack.Item == core.ItemID(2) || g.heldStack.Empty() {
		t.Fatalf("creative swap: slot=%+v held=%+v", g.inv.Slots[4], g.heldStack)
	}
}
func TestCreativeInventoryTabs(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.mode = ModeCreative
	g.inventoryTab = 1
	inv, creative := g.inventoryTabRects()
	g.mouseX = inv.X + inv.W/2
	g.mouseY = inv.Y + inv.H/2
	g.handleUIClick(false)
	if g.inventoryTab != 0 {
		t.Fatalf("survival tab did not activate: %d", g.inventoryTab)
	}
	g.mouseX = creative.X + creative.W/2
	g.mouseY = creative.Y + creative.H/2
	g.handleUIClick(false)
	if g.inventoryTab != 1 {
		t.Fatalf("creative tab did not activate: %d", g.inventoryTab)
	}
}
func TestCreativeSlotPickupFloats(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.mode = ModeCreative
	g.inventoryTab = 1
	before := g.inv.Slots[0]
	r := g.creativeHotbarRect(0)
	g.mouseX = r.X + r.W/2
	g.mouseY = r.Y + r.H/2
	g.handleUIClick(false)
	if g.heldStack != before || !g.inv.Slots[0].Empty() {
		t.Fatalf("creative slot pickup: held=%+v slot=%+v", g.heldStack, g.inv.Slots[0])
	}
}
func TestClosingCreativeReturnsHeldStack(t *testing.T) {
	g := newTestGame(t)
	g.inv = core.Inventory{}
	g.mode = ModeCreative
	g.heldStack = core.Stack{Item: core.ItemPlanks, Count: 7}
	g.returnHeldStack()
	if !g.heldStack.Empty() {
		t.Fatalf("held stack not cleared: %+v", g.heldStack)
	}
	if got := g.inv.Count(core.ItemPlanks); got != 7 {
		t.Fatalf("returned planks = %d", got)
	}
}
func TestCreativeClickCountsAndShiftStack(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.mode = ModeCreative
	g.inventoryTab = 1
	planks := g.creativeItemRect(int(core.ItemPlanks) - 1)
	g.mouseX = planks.X + planks.W/2
	g.mouseY = planks.Y + planks.H/2

	g.handleUIClick(false)
	if g.heldStack.Count != 1 {
		t.Fatalf("single click count = %d", g.heldStack.Count)
	}
	g.shiftDown = true
	g.handleUIClick(false)
	g.shiftDown = false
	if g.heldStack.Item != core.ItemPlanks || g.heldStack.Count != 64 {
		t.Fatalf("shift click stack = %+v", g.heldStack)
	}
}

func TestCreativeCatalogDifferentItemDestroysHeld(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.state = StateInventory
	g.mode = ModeCreative
	g.inventoryTab = 1
	first := g.creativeItemRect(0)
	g.mouseX = first.X + first.W/2
	g.mouseY = first.Y + first.H/2
	g.handleUIClick(false)
	second := g.creativeItemRect(1)
	g.mouseX = second.X + second.W/2
	g.mouseY = second.Y + second.H/2
	g.handleUIClick(false)
	if !g.heldStack.Empty() {
		t.Fatalf("different catalog item should destroy held stack: %+v", g.heldStack)
	}
	// The clicked item is not taken on the destroying click.
	g.handleUIClick(false)
	if g.heldStack.Item != core.ItemID(2) || g.heldStack.Count != 1 {
		t.Fatalf("clicked item should be taken on the next click: %+v", g.heldStack)
	}
}
