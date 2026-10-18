package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
	"sevg/internal/ui"
)

const chestSlots = 27

type chestLayout struct {
	panel  ui.Rect
	chest  [chestSlots]ui.Rect
	inv    [27]ui.Rect
	hotbar [9]ui.Rect
}

func (g *Game) chestLayout() chestLayout {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	const slot, gap = float32(36), float32(4)
	const cols = 9
	gridW := cols*slot + (cols-1)*gap
	pw := gridW + 64
	ph := float32(430)
	p := ui.Rect{X: (w - pw) / 2, Y: (h - ph) / 2, W: pw, H: ph}
	l := chestLayout{panel: p}
	x0 := p.X + (p.W-gridW)/2
	chestY := p.Y + 62
	invY := p.Y + 232
	hotY := p.Y + 364
	for i := 0; i < chestSlots; i++ {
		col, row := i%9, i/9
		l.chest[i] = ui.Rect{X: x0 + float32(col)*(slot+gap), Y: chestY + float32(row)*(slot+gap), W: slot, H: slot}
	}
	for i := 0; i < 27; i++ {
		col, row := i%9, i/9
		l.inv[i] = ui.Rect{X: x0 + float32(col)*(slot+gap), Y: invY + float32(row)*(slot+gap), W: slot, H: slot}
	}
	for i := 0; i < 9; i++ {
		l.hotbar[i] = ui.Rect{X: x0 + float32(i)*(slot+gap), Y: hotY, W: slot, H: slot}
	}
	return l
}

func (g *Game) openChest(pos [3]int) {
	if g.state != StatePlaying {
		return
	}
	if g.chests == nil {
		g.chests = make(map[[3]int]*[27]core.Stack)
	}
	if _, ok := g.chests[pos]; !ok {
		slots := [27]core.Stack{}
		g.chests[pos] = &slots
	}
	g.chestPos = pos
	g.placeCool = 0.25
	g.setState(StateChest)
}

func (g *Game) chestAt(pos [3]int) *[27]core.Stack {
	if g.chests == nil {
		g.chests = make(map[[3]int]*[27]core.Stack)
	}
	slots := g.chests[pos]
	if slots == nil {
		s := [27]core.Stack{}
		slots = &s
		g.chests[pos] = slots
	}
	return slots
}

func (g *Game) dropChest(pos [3]int) {
	slots := g.chests[pos]
	if slots == nil {
		return
	}
	delete(g.chests, pos)
	for i, st := range *slots {
		if st.Empty() || g.mobs == nil {
			continue
		}
		ang := float32((g.frame+i*37)%360) * 0.0174
		p := mathx.Vec3{X: float32(pos[0]) + 0.5, Y: float32(pos[1]) + 0.6, Z: float32(pos[2]) + 0.5}
		vel := mathx.Vec3{X: float32(math.Cos(float64(ang))) * 1.3, Y: 2.4, Z: float32(math.Sin(float64(ang))) * 1.3}
		g.mobs.Drop(st.Item, st.Count, p, vel)
	}
}

func (g *Game) chestSaveData() map[[3]int][27]core.Stack {
	if len(g.chests) == 0 {
		return nil
	}
	out := make(map[[3]int][27]core.Stack, len(g.chests))
	for pos, slots := range g.chests {
		out[pos] = *slots
	}
	return out
}

func (g *Game) drawChest(w, h float32) {
	c := g.canv
	l := g.chestLayout()
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.48})
	p := l.panel
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(24, 28, 32, 245), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+20, p.Y+14, i18n.T("chest.title"), 1.4, ui.TextColor)
	c.Text(p.X+20, p.Y+42, i18n.T("chest.storage"), 0.9, ui.MutedColor)
	c.Text(p.X+20, p.Y+212, i18n.T("chest.inventory"), 0.9, ui.MutedColor)

	slots := g.chestAt(g.chestPos)
	for i, r := range l.chest {
		g.drawInventorySlot(r, slots[i], ui.RGBA(16, 20, 24, 220), ui.RGBA(70, 82, 90, 200), 0)
	}
	for i, r := range l.inv {
		g.drawInventorySlot(r, g.inv.Slots[9+i], ui.RGBA(16, 20, 24, 220), ui.RGBA(70, 82, 90, 200), 0)
	}
	for i, r := range l.hotbar {
		bd := ui.RGBA(70, 82, 90, 200)
		if i == g.inv.Selected {
			bd = ui.Accent
		}
		g.drawInventorySlot(r, g.inv.Slots[i], ui.RGBA(16, 20, 24, 220), bd, 0)
	}
	c.TextCentered(p.X+p.W/2, p.Y+404, i18n.T("chest.hint"), 0.8, ui.MutedColor)

	hover := func(r ui.Rect, st core.Stack) {
		if !r.Contains(g.mouseX, g.mouseY) || st.Empty() {
			return
		}
		g.drawItemTooltip(i18n.ItemName(st.Item))
	}
	for i, r := range l.chest {
		hover(r, slots[i])
	}
	for i, r := range l.inv {
		hover(r, g.inv.Slots[9+i])
	}
	for i, r := range l.hotbar {
		hover(r, g.inv.Slots[i])
	}
	g.drawCursorStack()
}

func (g *Game) clickChest(right bool) {
	l := g.chestLayout()
	if !l.panel.Contains(g.mouseX, g.mouseY) {
		g.dropHeldOutside()
		return
	}
	slots := g.chestAt(g.chestPos)
	for i, r := range l.chest {
		if r.Contains(g.mouseX, g.mouseY) {
			g.clickStack(&slots[i], right)
			return
		}
	}
	for i, r := range l.inv {
		if r.Contains(g.mouseX, g.mouseY) {
			g.clickStack(&g.inv.Slots[9+i], right)
			return
		}
	}
	for i, r := range l.hotbar {
		if r.Contains(g.mouseX, g.mouseY) {
			g.clickStack(&g.inv.Slots[i], right)
			return
		}
	}
}
