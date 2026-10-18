package game

import (
	"sevg/internal/core"
	"sevg/internal/mathx"
)

func (g *Game) throwItem(item core.ItemID, count int) {
	if item == core.ItemNone || count <= 0 || g.mobs == nil || g.player == nil {
		return
	}
	dir := g.player.LookDir()
	pos := g.player.Eye().Add(dir.Mul(0.8))
	g.mobs.Drop(item, count, pos, dir.Mul(3).Add(mathx.Vec3{Y: 1.5}))
}

func (g *Game) dropFromStack(st *core.Stack, count int) {
	if st == nil || st.Empty() {
		return
	}
	if count > st.Count {
		count = st.Count
	}
	g.throwItem(st.Item, count)
	st.Count -= count
	if st.Count <= 0 {
		*st = core.Stack{}
	}
}

func (g *Game) dropHeld(count int) {
	if g.heldStack.Empty() {
		return
	}
	if count > g.heldStack.Count {
		count = g.heldStack.Count
	}
	g.throwItem(g.heldStack.Item, count)
	g.heldStack.Count -= count
	if g.heldStack.Count <= 0 {
		g.heldStack = core.Stack{}
	}
}

// dropFromInventory handles Q while the inventory is open. It drops from the
// cursor first, then from the hovered storage/craft/hotbar slot.
func (g *Game) dropFromInventory() {
	count := 1
	if g.shiftDown {
		count = 1 << 30
	}
	if !g.heldStack.Empty() {
		g.dropHeld(count)
		return
	}
	if g.mode == ModeCreative && g.inventoryTab == 1 {
		for i := 0; i < 9; i++ {
			if g.creativeHotbarRect(i).Contains(g.mouseX, g.mouseY) {
				g.dropFromStack(&g.inv.Slots[i], count)
				return
			}
		}
		// Q over the infinite creative catalog drops a fresh item without
		// consuming or creating anything in the inventory.
		if item := g.creativeItemAt(g.mouseX, g.mouseY); item != core.ItemNone {
			g.throwItem(item, creativeTakeCount(item, g.shiftDown))
		}
		return
	}
	for vis := 0; vis < core.InventorySlots; vis++ {
		if g.invSlotRect(vis).Contains(g.mouseX, g.mouseY) {
			g.dropFromStack(&g.inv.Slots[visualToSlot(vis)], count)
			return
		}
	}
	if g.mode != ModeCreative {
		for i := 0; i < g.craftGridN()*g.craftGridN(); i++ {
			if g.craftSlotRect(i).Contains(g.mouseX, g.mouseY) {
				g.dropFromStack(&g.craftGrid[i], count)
				return
			}
		}
	}
}

// dropHeldOutside is used when the player clicks outside the inventory panel
// while carrying an item on the cursor.
func (g *Game) dropHeldOutside() {
	if g.heldStack.Empty() {
		return
	}
	g.throwItem(g.heldStack.Item, g.heldStack.Count)
	g.heldStack = core.Stack{}
}
