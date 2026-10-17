package game

import (
	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/ui"
)

const (
	creativeCols = 9
	creativeRows = 4
)

func (g *Game) creativePanelRect() ui.Rect {
	return g.invPanelRect()
}

func (g *Game) creativeCloseRect() ui.Rect {
	p := g.creativePanelRect()
	return ui.Rect{X: p.X + p.W - 38, Y: p.Y + 12, W: 26, H: 26}
}

func (g *Game) creativeItemRect(idx int) ui.Rect {
	p := g.creativePanelRect()
	const slot, gap = float32(42), float32(6)
	total := slot*9 + gap*8
	col, row := idx%creativeCols, idx/creativeCols
	return ui.Rect{
		X: p.X + (p.W-total)/2 + float32(col)*(slot+gap),
		Y: p.Y + 70 + float32(row)*(slot+gap),
		W: slot, H: slot,
	}
}

func (g *Game) creativeHotbarRect(idx int) ui.Rect {
	p := g.creativePanelRect()
	const slot, gap = float32(42), float32(6)
	total := slot*9 + gap*8
	x0 := p.X + (p.W-total)/2
	return ui.Rect{X: x0 + float32(idx)*(slot+gap), Y: p.Y + 300, W: slot, H: slot}
}

func (g *Game) creativeItemAt(x, y float32) core.ItemID {
	for idx := 0; idx < creativeCols*creativeRows; idx++ {
		if idx >= int(core.ItemCount)-1 {
			continue
		}
		if g.creativeItemRect(idx).Contains(x, y) {
			return core.ItemID(idx + 1)
		}
	}
	return core.ItemNone
}

func (g *Game) giveCreativeItem(item core.ItemID) {
	if item == core.ItemNone || item >= core.ItemCount {
		return
	}
	count := item.Def().MaxStack
	if count <= 0 {
		count = 64
	}
	// The creative catalog is an infinite source: the picked stack stays on
	// the cursor until it is placed into a slot.
	g.heldStack = core.Stack{Item: item, Count: count}
}

func (g *Game) drawCreativeInventory(w, h float32) {
	c := g.canv
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.48})
	p := g.creativePanelRect()
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(24, 28, 32, 248), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+24, p.Y+14, i18n.T("inventory.creative_title"), 2, ui.TextColor)
	c.Text(p.X+24, p.Y+50, i18n.T("inventory.creative_hint"), 0.85, ui.MutedColor)

	closeR := g.creativeCloseRect()
	c.Rect(closeR.X, closeR.Y, closeR.W, closeR.H, ui.RGBA(180, 60, 60, 230))
	c.TextCentered(closeR.X+closeR.W/2, closeR.Y+3, "X", 1.5, ui.TextColor)

	for idx := 0; idx < creativeCols*creativeRows; idx++ {
		r := g.creativeItemRect(idx)
		item := core.ItemNone
		if idx < int(core.ItemCount)-1 {
			item = core.ItemID(idx + 1)
		}
		hover := item != core.ItemNone && r.Contains(g.mouseX, g.mouseY)
		bg := ui.RGBA(16, 20, 24, 220)
		bd := ui.RGBA(70, 82, 90, 200)
		if hover {
			bg = ui.RGBA(44, 51, 58, 240)
			bd = ui.Accent
		}
		st := core.Stack{}
		if item != core.ItemNone {
			st = core.Stack{Item: item, Count: 1}
		}
		g.drawInventorySlot(r, st, bg, bd, 0)
	}
	if item := g.creativeItemAt(g.mouseX, g.mouseY); item != core.ItemNone {
		name := i18n.ItemName(item)
		tw := c.TextWidth(name, 1.1) + 16
		c.Panel(g.mouseX+14, g.mouseY+12, tw, 24, ui.RGBA(8, 12, 16, 240), ui.RGBA(90, 105, 115, 220))
		c.Text(g.mouseX+22, g.mouseY+17, name, 1.1, ui.TextColor)
	}

	const slot = float32(42)
	for i := 0; i < 9; i++ {
		r := g.creativeHotbarRect(i)
		selected := i == g.inv.Selected
		bg := ui.RGBA(18, 22, 26, 220)
		bd := ui.RGBA(88, 98, 106, 220)
		if selected {
			bg = ui.RGBA(38, 43, 48, 235)
			bd = ui.Accent
		} else if r.Contains(g.mouseX, g.mouseY) {
			bg = ui.RGBA(40, 47, 53, 235)
		}
		g.drawInventorySlot(r, g.inv.Slots[i], bg, bd, i+1)
	}
	hintY := g.creativeHotbarRect(0).Y + slot + 8
	c.TextCentered(p.X+p.W/2, hintY, i18n.T("inventory.creative_get"), 0.9, ui.MutedColor)
	g.drawCursorStack()
}

func (g *Game) inventoryTabRects() (ui.Rect, ui.Rect) {
	p := g.invPanelRect()
	y := p.Y - 36
	if y < 4 {
		y = 4
	}
	return ui.Rect{X: p.X + 16, Y: y, W: 150, H: 30},
		ui.Rect{X: p.X + 174, Y: y, W: 150, H: 30}
}

func (g *Game) drawInventoryTabs() {
	c := g.canv
	inv, creative := g.inventoryTabRects()
	labels := []string{i18n.T("inventory.tab_survival"), i18n.T("inventory.tab_creative")}
	for i, r := range []ui.Rect{inv, creative} {
		active := (i == 0 && g.inventoryTab == 0) || (i == 1 && g.inventoryTab == 1)
		bg := ui.RGBA(18, 22, 26, 240)
		bd := ui.RGBA(80, 95, 105, 220)
		col := ui.MutedColor
		if active {
			bg = ui.RGBA(40, 46, 52, 250)
			bd = ui.Accent
			col = ui.TextColor
		} else if r.Contains(g.mouseX, g.mouseY) {
			bg = ui.RGBA(32, 38, 44, 245)
		}
		c.Rect(r.X, r.Y, r.W, r.H, bg)
		c.Border(r.X, r.Y, r.W, r.H, 2, bd)
		c.TextCentered(r.X+r.W/2, r.Y+6, labels[i], 1.0, col)
	}
}
