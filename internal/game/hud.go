package game

import (
	"fmt"
	"math"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
	"sevg/internal/ui"
	"sevg/internal/world"
)

func (g *Game) drawUI() {
	g.canv.Resize(g.gfx.W, g.gfx.H)
	g.canv.Reset()
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	switch g.state {
	case StateMenu:
		g.drawMenu(w, h)
	default:
		g.drawHUD(w, h)
		switch g.state {
		case StateInventory:
			g.drawInventory(w, h)
		case StatePaused:
			g.drawPause(w, h)
		case StateDead:
			g.drawDead(w, h)
		}
	}
	if g.toastTime > 0 && g.toast != "" {
		toastY := float32(58)
		if g.state != StatePlaying {
			toastY = 18
		}
		tw := g.canv.TextWidth(g.toast, 1.5) + 32
		x := w/2 - tw/2
		g.canv.Panel(x, toastY, tw, 34, ui.RGBA(20, 25, 30, 230), ui.RGBA(255, 196, 82, 220))
		g.canv.TextCentered(w/2, toastY+9, g.toast, 1.5, ui.TextColor)
	}
}

func (g *Game) drawMenu(w, h float32) {
	// Sky gradient.
	for i := 0; i < 30; i++ {
		t := float32(i) / 29
		col := [4]float32{0.05 + 0.30*t, 0.10 + 0.38*t, 0.20 + 0.45*t, 1}
		g.canv.Rect(0, h*t/1.0, w, h/29+1, col)
	}
	// Pixel-art horizon.
	base := h * 0.80
	for x := float32(0); x < w; x += 8 {
		hh := 24 + 20*float32(math.Sin(float64(x)*0.011)) + 10*float32(math.Sin(float64(x)*0.037))
		if hh < 6 {
			hh = 6
		}
		g.canv.Rect(x, base-hh, 8, hh+2, ui.RGBA(32, 58, 42, 255))
		g.canv.Rect(x, base+2, 8, 50, ui.RGBA(46, 34, 25, 255))
	}
	// A few blocky trees.
	for i := 0; i < 11; i++ {
		x := w * float32(i) / 10.5
		th := 30 + float32((i*37)%22)
		g.canv.Rect(x-3, base-th, 7, th, ui.RGBA(54, 38, 24, 255))
		g.canv.Rect(x-20, base-th-24, 42, 28, ui.RGBA(31, 76, 42, 255))
		g.canv.Rect(x-14, base-th-34, 28, 14, ui.RGBA(39, 92, 50, 255))
	}
	g.canv.Rect(0, base+52, w, h, ui.RGBA(28, 22, 18, 255))

	g.canv.TextShadowCentered(w/2, h*0.12, "SEVG", 7.5, ui.TextColor)
	g.canv.TextShadowCentered(w/2, h*0.33, i18n.T("menu.subtitle"), 2, ui.Accent)
	g.canv.TextCentered(w/2, h*0.39, i18n.T("menu.tagline"), 1, ui.MutedColor)

	rects := g.menuRects()
	labels := []string{i18n.T("menu.new"), i18n.T("menu.continue"), i18n.T("menu.quit")}
	for i, r := range rects {
		enabled := i != 1 || g.saveExists()
		g.drawButton(r, labels[i], enabled)
	}
	g.canv.TextCentered(w/2, rects[0].Y-32, i18n.T("menu.enter_hint"), 1.2, ui.MutedColor)

	// Language selector.
	lr := g.menuLangRect()
	g.drawButtonScale(lr, i18n.T("menu.language", i18n.LocaleName(i18n.Current())), true, 1.4)
	g.canv.TextCentered(lr.X+lr.W/2, lr.Y+lr.H+10, i18n.T("menu.language_hint"), 0.9, ui.MutedColor)

	// Controls card.
	px, py := float32(26), h-238
	g.canv.Panel(px, py, 430, 208, ui.RGBA(12, 16, 20, 220), ui.RGBA(80, 100, 110, 180))
	g.canv.Text(px+18, py+14, i18n.T("menu.controls"), 1.6, ui.Accent)
	lines := []string{
		i18n.T("menu.ctrl_move"),
		i18n.T("menu.ctrl_sprint"),
		i18n.T("menu.ctrl_mine"),
		i18n.T("menu.ctrl_pick"),
		i18n.T("menu.ctrl_inv"),
		i18n.T("menu.ctrl_pause"),
	}
	for i, line := range lines {
		g.canv.Text(px+18, py+46+float32(i)*22, line, 1.0, ui.TextColor)
	}
	rd := i18n.T("menu.render_distance", g.opts.RenderDist)
	g.canv.Text(w-26-g.canv.TextWidth(rd, 1), h-28, rd, 1, ui.RGBA(120, 140, 150, 255))
}

func (g *Game) drawHUD(w, h float32) {
	if g.player == nil {
		return
	}
	c := g.canv
	// Underwater tint.
	eye := g.player.Eye()
	eb := g.world.GetBlock(int(math.Floor(float64(eye.X))), int(math.Floor(float64(eye.Y))), int(math.Floor(float64(eye.Z))))
	if eb == core.Water {
		c.Rect(0, 0, w, h, [4]float32{0.05, 0.30, 0.50, 0.22})
	}

	// Crosshair and mining progress are only shown while playing.
	if g.state == StatePlaying {
		cx, cy := w/2, h/2
		col := ui.RGBA(245, 245, 245, 230)
		shadow := ui.RGBA(0, 0, 0, 180)
		for _, r := range []ui.Rect{{X: cx - 11, Y: cy - 1, W: 22, H: 2}, {X: cx - 1, Y: cy - 11, W: 2, H: 22}} {
			c.Rect(r.X+1, r.Y+1, r.W, r.H, shadow)
			c.Rect(r.X, r.Y, r.W, r.H, col)
		}
		if g.breakActive && g.breakProgress > 0 {
			bw := float32(76)
			bx, by := cx-bw/2, cy+20
			c.Rect(bx-2, by-2, bw+4, 10, ui.RGBA(0, 0, 0, 180))
			c.Rect(bx, by, bw*clampF(g.breakProgress, 0, 1), 6, ui.Accent)
		}
	}

	// Health.
	hx, hy := float32(24), float32(22)
	c.Panel(hx, hy, 214, 48, ui.RGBA(12, 16, 20, 190), ui.RGBA(70, 90, 100, 180))
	c.Text(hx+12, hy+6, i18n.T("hud.health"), 1.1, ui.MutedColor)
	c.Rect(hx+12, hy+27, 180, 10, ui.RGBA(60, 20, 20, 255))
	c.Rect(hx+12, hy+27, 180*clampF(g.player.Health/g.player.MaxHP, 0, 1), 10, ui.RGBA(220, 55, 60, 255))
	c.Text(hx+196-c.TextWidth(fmt.Sprintf("%.0f", g.player.Health), 1), hy+6, fmt.Sprintf("%.0f", g.player.Health), 1, ui.TextColor)

	// Time and chunk progress.
	clock := g.timeString()
	c.TextShadow(w-28-c.TextWidth(clock, 1.6), 22, clock, 1.6, ui.TextColor)
	chunkText := i18n.T("hud.chunks", len(g.meshed), g.loadedTarget())
	c.Text(w-28-c.TextWidth(chunkText, 1), 48, chunkText, 1, ui.MutedColor)
	if len(g.meshed) < g.loadedTarget()/3 {
		prog := float32(len(g.meshed)) / float32(g.loadedTarget())
		c.Rect(w-228, 66, 200, 8, ui.RGBA(0, 0, 0, 150))
		c.Rect(w-228, 66, 200*prog, 8, ui.Accent)
		gen := i18n.T("hud.generating")
		c.Text(w-28-c.TextWidth(gen, 1), 78, gen, 1, ui.MutedColor)
	}

	if g.state != StatePlaying {
		return
	}
	// Hotbar.
	slot := float32(54)
	gap := float32(6)
	total := slot*9 + gap*8
	x0 := w/2 - total/2
	y0 := h - 20 - slot
	for i := 0; i < 9; i++ {
		x := x0 + float32(i)*(slot+gap)
		selected := i == g.inv.Selected
		bg := ui.RGBA(18, 22, 26, 205)
		border := ui.RGBA(88, 98, 106, 220)
		if selected {
			bg = ui.RGBA(38, 43, 48, 230)
			border = ui.Accent
		}
		c.Rect(x, y0, slot, slot, bg)
		c.Border(x, y0, slot, slot, 2, border)
		st := g.inv.Slots[i]
		if !st.Empty() {
			g.drawItemIcon(x+7, y0+7, slot-14, st, 1)
		}
		c.Text(x+4, y0+2, fmt.Sprint(i+1), 0.9, ui.MutedColor)
	}
	// Selected item name.
	st := g.inv.SelectedStack()
	if !st.Empty() {
		label := i18n.ItemName(st.Item)
		lw := c.TextWidth(label, 1.2) + 24
		c.Panel(w/2-lw/2, y0-34, lw, 26, ui.RGBA(12, 16, 20, 200), ui.RGBA(70, 90, 100, 150))
		c.TextCentered(w/2, y0-29, label, 1.2, ui.TextColor)
	}

	if g.helpTime > 0 && g.state == StatePlaying {
		px, py := float32(24), h-300
		c.Panel(px, py, 500, 192, ui.RGBA(12, 16, 20, 195), ui.RGBA(80, 100, 110, 160))
		c.Text(px+14, py+10, i18n.T("hud.quick_start"), 1.5, ui.Accent)
		lines := []string{
			i18n.T("hud.help1"),
			i18n.T("hud.help2"),
			i18n.T("hud.help3"),
			i18n.T("hud.help4"),
			i18n.T("hud.help5"),
			i18n.T("hud.help6"),
		}
		for i, line := range lines {
			c.Text(px+14, py+38+float32(i)*22, line, 1.0, ui.TextColor)
		}
	}
	if g.debug {
		g.drawDebug(w)
	}
}

func (g *Game) drawDebug(w float32) {
	c := g.canv
	px, py := float32(24), float32(86)
	c.Panel(px, py, 380, 276, ui.RGBA(8, 12, 16, 220), ui.RGBA(80, 100, 110, 180))
	biome := i18n.T("common.unknown")
	target := i18n.T("hud.target_none")
	if g.player != nil && g.world != nil {
		_, b := g.world.Gen.HeightAt(int(g.player.Body.Pos.X), int(g.player.Body.Pos.Z))
		biome = i18n.BiomeName(b)
	}
	if g.breakActive {
		target = i18n.T("hud.target_pos", g.breakPos[0], g.breakPos[1], g.breakPos[2])
	}
	lines := []string{
		i18n.T("hud.debug_title"),
		i18n.T("hud.fps", g.fps),
		i18n.T("hud.pos", g.player.Body.Pos.X, g.player.Body.Pos.Y, g.player.Body.Pos.Z),
		i18n.T("hud.chunk", worldChunkX(g.player.Body.Pos.X), worldChunkZ(g.player.Body.Pos.Z)),
		i18n.T("hud.biome", biome),
		i18n.T("hud.seed", g.seed),
		i18n.T("hud.time", g.timeString()),
		i18n.T("hud.chunk_stats", g.world.Loaded(), len(g.meshed)),
		i18n.T("hud.mob_stats", g.mobs.MobCount(), g.mobs.ItemCount()),
		i18n.T("hud.target", target),
		i18n.T("hud.fly", onOff(g.player.Flying)),
	}
	for i, line := range lines {
		col := ui.TextColor
		if i == 0 {
			col = ui.Accent
		}
		c.Text(px+12, py+10+float32(i)*21, line, 1.0, col)
	}
}

func (g *Game) drawInventory(w, h float32) {
	c := g.canv
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.48})
	p := g.invPanelRect()
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(24, 28, 32, 245), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+24, p.Y+14, i18n.T("inventory.title"), 2, ui.TextColor)
	c.Text(p.X+24, p.Y+49, i18n.T("inventory.hint"), 0.9, ui.MutedColor)
	closeR := g.invCloseRect()
	c.Rect(closeR.X, closeR.Y, closeR.W, closeR.H, ui.RGBA(180, 60, 60, 230))
	c.TextCentered(closeR.X+closeR.W/2, closeR.Y+3, "X", 1.5, ui.TextColor)

	for vis := 0; vis < core.InventorySlots; vis++ {
		r := g.invSlotRect(vis)
		idx := visualToSlot(vis)
		selected := idx == g.inv.Selected
		bg := ui.RGBA(16, 20, 24, 220)
		bd := ui.RGBA(70, 82, 90, 200)
		if selected {
			bd = ui.Accent
		}
		c.Rect(r.X, r.Y, r.W, r.H, bg)
		c.Border(r.X, r.Y, r.W, r.H, 2, bd)
		st := g.inv.Slots[idx]
		if !st.Empty() {
			g.drawItemIcon(r.X+6, r.Y+6, r.W-12, st, 1)
		}
	}

	// Crafting column.
	cx := p.X + 474
	c.Text(cx, p.Y+14, i18n.T("inventory.crafting"), 1.8, ui.Accent)
	c.Text(cx, p.Y+47, i18n.T("inventory.handmade"), 0.8, ui.MutedColor)
	for i, r := range g.recipeRects() {
		rec := recipes[i]
		ok := g.canCraft(rec)
		bg := ui.RGBA(40, 46, 52, 230)
		bd := ui.RGBA(90, 105, 115, 230)
		if ok {
			bg = ui.RGBA(44, 64, 48, 240)
			bd = ui.RGBA(115, 200, 120, 230)
		}
		if r.Contains(g.mouseX, g.mouseY) && ok {
			bg = ui.RGBA(60, 84, 60, 245)
			bd = ui.Good
		}
		c.Rect(r.X, r.Y, r.W, r.H, bg)
		c.Border(r.X, r.Y, r.W, r.H, 2, bd)
		c.Text(r.X+8, r.Y+2, i18n.ItemName(rec.Output.Item), 1.0, ui.TextColor)
		c.Text(r.X+8, r.Y+20, recipeInputs(rec), 0.72, ui.MutedColor)
		// Output count badge.
		c.Text(r.X+r.W-30, r.Y+3, fmt.Sprintf("x%d", rec.Output.Count), 1.0, ui.Accent)
	}
	c.Text(cx, p.Y+402, i18n.T("inventory.tip1"), 0.85, ui.MutedColor)
	c.Text(cx, p.Y+420, i18n.T("inventory.tip2"), 0.85, ui.MutedColor)

	// Tooltip for the hovered inventory slot.
	for vis := 0; vis < core.InventorySlots; vis++ {
		if g.invSlotRect(vis).Contains(g.mouseX, g.mouseY) {
			st := g.inv.Slots[visualToSlot(vis)]
			if !st.Empty() {
				name := i18n.ItemName(st.Item)
				tw := c.TextWidth(name, 1.1) + 16
				c.Panel(g.mouseX+14, g.mouseY+12, tw, 24, ui.RGBA(8, 12, 16, 240), ui.RGBA(90, 105, 115, 220))
				c.Text(g.mouseX+22, g.mouseY+17, name, 1.1, ui.TextColor)
			}
		}
	}
	// Cursor stack.
	if !g.heldStack.Empty() {
		g.drawItemIcon(g.mouseX-20, g.mouseY-20, 40, g.heldStack, 1)
	}
}

func (g *Game) drawPause(w, h float32) {
	c := g.canv
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.55})
	c.TextShadowCentered(w/2, h*0.13, i18n.T("pause.title"), 4.5, ui.TextColor)
	labels := []string{i18n.T("pause.resume"), i18n.T("pause.save"), i18n.T("pause.title_screen"), i18n.T("pause.quit")}
	for i, r := range g.pauseRects() {
		g.drawButton(r, labels[i], true)
	}
	c.TextCentered(w/2, h*0.72, i18n.T("pause.hint"), 1.2, ui.MutedColor)
}

func (g *Game) drawDead(w, h float32) {
	c := g.canv
	c.Rect(0, 0, w, h, [4]float32{0.25, 0, 0, 0.55})
	c.TextShadowCentered(w/2, h*0.16, i18n.T("death.title"), 4.5, ui.RGBA(240, 100, 100, 255))
	c.TextCentered(w/2, h*0.34, i18n.T("death.subtitle"), 1.2, ui.MutedColor)
	g.drawButton(g.respawnRect(), i18n.T("death.respawn"), true)
	r := g.deadTitleRect()
	hover := r.Contains(g.mouseX, g.mouseY)
	bg := ui.RGBA(30, 34, 38, 230)
	if hover {
		bg = ui.RGBA(48, 54, 60, 245)
	}
	c.Rect(r.X, r.Y, r.W, r.H, bg)
	c.Border(r.X, r.Y, r.W, r.H, 2, ui.RGBA(90, 105, 115, 220))
	c.TextCentered(r.X+r.W/2, r.Y+10, i18n.T("death.title_screen"), 1.4, ui.TextColor)
}

func (g *Game) drawButton(r ui.Rect, label string, enabled bool) {
	g.drawButtonScale(r, label, enabled, 1.8)
}

func (g *Game) drawButtonScale(r ui.Rect, label string, enabled bool, scale float32) {
	c := g.canv
	hover := enabled && r.Contains(g.mouseX, g.mouseY)
	bg := ui.RGBA(24, 30, 35, 235)
	border := ui.RGBA(92, 110, 120, 230)
	text := ui.TextColor
	if hover {
		bg = ui.RGBA(48, 58, 56, 245)
		border = ui.Accent
	}
	if !enabled {
		bg = ui.RGBA(18, 20, 22, 180)
		border = ui.RGBA(55, 60, 64, 180)
		text = ui.RGBA(110, 116, 120, 255)
	}
	c.Rect(r.X+4, r.Y+4, r.W, r.H, [4]float32{0, 0, 0, 0.35})
	c.Rect(r.X, r.Y, r.W, r.H, bg)
	c.Border(r.X, r.Y, r.W, r.H, 2, border)
	c.TextCentered(r.X+r.W/2, r.Y+r.H/2-8*scale/1.8, label, scale, text)
}

func (g *Game) drawItemIcon(x, y, size float32, st core.Stack, alpha float32) {
	if st.Empty() {
		return
	}
	c := g.canv
	col3 := st.Item.Def().Color
	base := [4]float32{col3[0], col3[1], col3[2], alpha}
	light := [4]float32{clampF(col3[0]*1.28, 0, 1), clampF(col3[1]*1.28, 0, 1), clampF(col3[2]*1.28, 0, 1), alpha}
	dark := [4]float32{col3[0] * 0.68, col3[1] * 0.68, col3[2] * 0.68, alpha}
	c.Rect(x+2, y+3, size-4, size-5, [4]float32{0, 0, 0, 0.35 * alpha})
	c.Rect(x+1, y+1, size-2, size-2, base)
	c.Rect(x+1, y+1, size-2, (size-2)*0.28, light)
	c.Rect(x+size*0.72, y+2, size*0.26, size-3, dark)
	c.Border(x+1, y+1, size-2, size-2, 1.5, [4]float32{0, 0, 0, 0.35 * alpha})
	name := i18n.ItemName(st.Item)
	if name != "" {
		letter := string([]rune(name)[0])
		c.TextCentered(x+size/2, y+size/2-9, letter, 1.0, [4]float32{1, 1, 1, alpha})
	}
	if st.Count > 1 {
		cnt := fmt.Sprint(st.Count)
		c.Text(x+size-c.TextWidth(cnt, 0.75)-2, y+size-16, cnt, 0.75, ui.TextColor)
	}
}

func (g *Game) canCraft(r recipe) bool {
	for _, in := range r.Inputs {
		if g.inv.Count(in.Item) < in.Count {
			return false
		}
	}
	return true
}

func recipeInputs(r recipe) string {
	out := ""
	for i, in := range r.Inputs {
		if i > 0 {
			out += " + "
		}
		out += fmt.Sprintf("%dx %s", in.Count, i18n.ItemName(in.Item))
	}
	return out
}

// Layout helpers -----------------------------------------------------------

func (g *Game) menuRects() []ui.Rect {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	bw, bh := float32(360), float32(54)
	x := (w - bw) / 2
	y := h*0.50 + 20
	out := make([]ui.Rect, 3)
	for i := range out {
		out[i] = ui.Rect{X: x, Y: y + float32(i)*(bh+12), W: bw, H: bh}
	}
	return out
}

func (g *Game) menuLangRect() ui.Rect {
	w := float32(g.gfx.W)
	return ui.Rect{X: w - 286, Y: 26, W: 250, H: 46}
}

func (g *Game) pauseRects() []ui.Rect {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	bw, bh := float32(360), float32(50)
	x := (w - bw) / 2
	y := h * 0.34
	out := make([]ui.Rect, 4)
	for i := range out {
		out[i] = ui.Rect{X: x, Y: y + float32(i)*(bh+12), W: bw, H: bh}
	}
	return out
}

func (g *Game) respawnRect() ui.Rect {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	return ui.Rect{X: w/2 - 180, Y: h * 0.48, W: 360, H: 56}
}

func (g *Game) deadTitleRect() ui.Rect {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	return ui.Rect{X: w/2 - 180, Y: h*0.48 + 72, W: 360, H: 46}
}

func (g *Game) invPanelRect() ui.Rect {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	pw, ph := float32(760), float32(460)
	if pw > w-32 {
		pw = w - 32
	}
	if ph > h-32 {
		ph = h - 32
	}
	return ui.Rect{X: (w - pw) / 2, Y: (h - ph) / 2, W: pw, H: ph}
}

func (g *Game) invSlotRect(vis int) ui.Rect {
	p := g.invPanelRect()
	col := vis % 9
	row := vis / 9
	x := p.X + 24 + float32(col)*48
	y := p.Y + 76 + float32(row)*48
	if row >= 3 {
		y += 18
	}
	return ui.Rect{X: x, Y: y, W: 42, H: 42}
}

func (g *Game) invCloseRect() ui.Rect {
	p := g.invPanelRect()
	return ui.Rect{X: p.X + p.W - 38, Y: p.Y + 12, W: 26, H: 26}
}

func (g *Game) recipeRects() []ui.Rect {
	p := g.invPanelRect()
	x := p.X + 474
	y := p.Y + 70
	out := make([]ui.Rect, len(recipes))
	for i := range out {
		out[i] = ui.Rect{X: x, Y: y + float32(i)*40, W: 240, H: 36}
	}
	return out
}

func visualToSlot(vis int) int {
	row, col := vis/9, vis%9
	if row < 3 {
		return 9 + row*9 + col
	}
	return col
}

func worldChunkX(x float32) int { return int(math.Floor(float64(x))) / 16 }
func worldChunkZ(z float32) int { return int(math.Floor(float64(z))) / 16 }

// Entity and block-highlight geometry, rendered through the block atlas
// pipeline so it shares fog, depth and lighting with terrain.
func (g *Game) buildEntityMeshes() (world.MeshData, world.MeshData) {
	var opaque, transparent world.MeshData
	if g.mobs != nil {
		for _, mob := range g.mobs.Mobs {
			if mob.Dead {
				continue
			}
			flash := clampF(mob.HitFlash/0.25, 0, 1)
			for _, box := range mob.Boxes() {
				col := box.Color
				if flash > 0 {
					col[0] = col[0] + (1-col[0])*flash*0.7
					col[1] = col[1] * (1 - flash*0.3)
					col[2] = col[2] * (1 - flash*0.3)
				}
				appendEntityBox(&opaque, box.Center.Sub(box.Size.Mul(0.5)), box.Center.Add(box.Size.Mul(0.5)), col)
			}
		}
		for _, d := range g.mobs.Items {
			if d.Dead {
				continue
			}
			bob := float32(math.Sin(float64(d.Age*4))) * 0.06
			c := d.Body.Pos.Add(mathx.Vec3{Y: 0.18})
			c.Y += bob
			def := d.Item.Def()
			switch d.Visual() {
			case entity.DropVisualBlock:
				half := def.Block
				appendBlockBox(&opaque, c.Sub(mathx.Vec3{X: 0.14, Y: 0.14, Z: 0.14}), c.Add(mathx.Vec3{X: 0.14, Y: 0.14, Z: 0.14}), half)
			case entity.DropVisualGlass:
				appendBlockBox(&transparent, c.Sub(mathx.Vec3{X: 0.14, Y: 0.14, Z: 0.14}), c.Add(mathx.Vec3{X: 0.14, Y: 0.14, Z: 0.14}), def.Block)
			case entity.DropVisualPlant:
				appendPlant(&opaque, c, 0.34, 0.42, def.Block)
			case entity.DropVisualSword:
				blade := uint8(core.TileWood)
				if d.Item == core.ItemIronSword {
					blade = uint8(core.TileWhite)
				}
				appendSword(&opaque, c, blade)
			case entity.DropVisualPickaxe:
				head := uint8(core.TileGray)
				if d.Item == core.ItemIronPickaxe {
					head = uint8(core.TileWhite)
				}
				appendPickaxe(&opaque, c, head)
			default:
				col := [4]float32{def.Color[0], def.Color[1], def.Color[2], 1}
				appendEntityBox(&opaque, c.Sub(mathx.Vec3{X: 0.14, Y: 0.14, Z: 0.14}), c.Add(mathx.Vec3{X: 0.14, Y: 0.14, Z: 0.14}), col)
			}
		}
	}
	if g.player != nil && g.targetValid {
		x, y, z := g.breakPos[0], g.breakPos[1], g.breakPos[2]
		min := mathx.V3(float32(x)-0.003, float32(y)-0.003, float32(z)-0.003)
		max := mathx.V3(float32(x)+1.003, float32(y)+1.003, float32(z)+1.003)
		appendOutline(&opaque, min, max, [4]float32{1, 0.88, 0.35, 1})
	}
	return opaque, transparent
}

func appendPlant(m *world.MeshData, c mathx.Vec3, width, height float32, block core.BlockID) {
	u0, v0, u1, v1 := world.TileUV(block.Tile(2))
	hx, hz := width/2, width/2
	top, bottom := c.Y+height/2, c.Y-height/2
	quads := [2][4]mathx.Vec3{
		{
			{X: c.X - hx, Y: bottom, Z: c.Z - hz},
			{X: c.X + hx, Y: bottom, Z: c.Z + hz},
			{X: c.X + hx, Y: top, Z: c.Z + hz},
			{X: c.X - hx, Y: top, Z: c.Z - hz},
		},
		{
			{X: c.X - hx, Y: bottom, Z: c.Z + hz},
			{X: c.X + hx, Y: bottom, Z: c.Z - hz},
			{X: c.X + hx, Y: top, Z: c.Z - hz},
			{X: c.X - hx, Y: top, Z: c.Z + hz},
		},
	}
	uvs := [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}
	for _, q := range quads {
		base := uint32(len(m.Vertices))
		for i := 0; i < 4; i++ {
			m.Vertices = append(m.Vertices, world.Vertex{
				X: q[i].X, Y: q[i].Y, Z: q[i].Z,
				U: u0 + (u1-u0)*uvs[i][0], V: v0 + (v1-v0)*uvs[i][1],
				Light: 0.92,
			})
		}
		m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}

func appendTileBox(m *world.MeshData, center, size mathx.Vec3, tile uint8) {
	appendBoxFaces(m, center.Sub(size.Mul(0.5)), center.Add(size.Mul(0.5)), func(int) uint8 { return tile })
}

func appendSword(m *world.MeshData, c mathx.Vec3, bladeTile uint8) {
	appendTileBox(m, c.Add(mathx.Vec3{Y: -0.09}), mathx.Vec3{X: 0.04, Y: 0.15, Z: 0.04}, core.TileBrown)
	appendTileBox(m, c.Add(mathx.Vec3{Y: -0.18}), mathx.Vec3{X: 0.08, Y: 0.04, Z: 0.05}, core.TileDark)
	appendTileBox(m, c.Add(mathx.Vec3{Y: -0.02}), mathx.Vec3{X: 0.17, Y: 0.035, Z: 0.05}, core.TileDark)
	appendTileBox(m, c.Add(mathx.Vec3{Y: 0.10}), mathx.Vec3{X: 0.07, Y: 0.22, Z: 0.025}, bladeTile)
	appendTileBox(m, c.Add(mathx.Vec3{Y: 0.225}), mathx.Vec3{X: 0.04, Y: 0.045, Z: 0.02}, core.TileWhite)
}

func appendPickaxe(m *world.MeshData, c mathx.Vec3, headTile uint8) {
	appendTileBox(m, c, mathx.Vec3{X: 0.04, Y: 0.34, Z: 0.04}, core.TileBrown)
	appendTileBox(m, c.Add(mathx.Vec3{Y: 0.15}), mathx.Vec3{X: 0.28, Y: 0.06, Z: 0.06}, headTile)
	appendTileBox(m, c.Add(mathx.Vec3{X: -0.15, Y: 0.14}), mathx.Vec3{X: 0.05, Y: 0.09, Z: 0.055}, core.TileWhite)
	appendTileBox(m, c.Add(mathx.Vec3{X: 0.15, Y: 0.14}), mathx.Vec3{X: 0.05, Y: 0.09, Z: 0.055}, core.TileWhite)
}

type entityPaletteEntry struct {
	tile uint8
	col  [3]float32
}

var entityPalette = []entityPaletteEntry{
	{core.TileWhite, [3]float32{0.93, 0.93, 0.90}},
	{core.TilePig, [3]float32{0.94, 0.52, 0.56}},
	{core.TileDark, [3]float32{0.18, 0.14, 0.13}},
	{core.TileRed, [3]float32{0.80, 0.19, 0.19}},
	{core.TileGold, [3]float32{1.00, 0.86, 0.36}},
	{core.TileGray, [3]float32{0.55, 0.55, 0.58}},
	{core.TileBlack, [3]float32{0.07, 0.07, 0.09}},
	{core.TileBrown, [3]float32{0.44, 0.31, 0.18}},
	{core.TilePink, [3]float32{0.91, 0.29, 0.42}},
	{core.TileGreen, [3]float32{0.25, 0.55, 0.24}},
}

func tileForColor(c [4]float32) uint8 {
	best := 0
	bestD := float32(1e9)
	for i, e := range entityPalette {
		dr, dg, db := c[0]-e.col[0], c[1]-e.col[1], c[2]-e.col[2]
		d := dr*dr + dg*dg + db*db
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return entityPalette[best].tile
}

type entityFace struct {
	corners [4][3]float32
	uv      [4][2]float32
	light   float32
}

var entityFaces = [6]entityFace{
	{corners: [4][3]float32{{1, 0, 1}, {1, 0, 0}, {1, 1, 0}, {1, 1, 1}}, uv: [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}, light: 0.78},
	{corners: [4][3]float32{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}, {0, 1, 0}}, uv: [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}, light: 0.70},
	{corners: [4][3]float32{{0, 1, 1}, {1, 1, 1}, {1, 1, 0}, {0, 1, 0}}, uv: [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}, light: 1.00},
	{corners: [4][3]float32{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}}, uv: [4][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, light: 0.46},
	{corners: [4][3]float32{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}}, uv: [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}, light: 0.88},
	{corners: [4][3]float32{{1, 0, 0}, {0, 0, 0}, {0, 1, 0}, {1, 1, 0}}, uv: [4][2]float32{{0, 1}, {1, 1}, {1, 0}, {0, 0}}, light: 0.82},
}

func appendEntityBox(m *world.MeshData, min, max mathx.Vec3, col [4]float32) {
	tile := tileForColor(col)
	appendBoxFaces(m, min, max, func(int) uint8 { return tile })
}

// appendBlockBox renders a dropped block with the same per-face atlas tiles
// as a block placed in the world.
func appendBlockBox(m *world.MeshData, min, max mathx.Vec3, block core.BlockID) {
	appendBoxFaces(m, min, max, block.Tile)
}

func appendBoxFaces(m *world.MeshData, min, max mathx.Vec3, tileAt func(int) uint8) {
	for fi, f := range entityFaces {
		u0, v0, u1, v1 := world.TileUV(tileAt(fi))
		base := uint32(len(m.Vertices))
		for i := 0; i < 4; i++ {
			p := f.corners[i]
			x := min.X + (max.X-min.X)*p[0]
			y := min.Y + (max.Y-min.Y)*p[1]
			z := min.Z + (max.Z-min.Z)*p[2]
			m.Vertices = append(m.Vertices, world.Vertex{
				X: x, Y: y, Z: z,
				U:     u0 + (u1-u0)*f.uv[i][0],
				V:     v0 + (v1-v0)*f.uv[i][1],
				Light: f.light,
			})
		}
		m.Indices = append(m.Indices, base, base+1, base+2, base, base+2, base+3)
	}
}

func appendOutline(m *world.MeshData, min, max mathx.Vec3, col [4]float32) {
	t := float32(0.018)
	xs := [2]float32{min.X, max.X - t}
	ys := [2]float32{min.Y, max.Y - t}
	zs := [2]float32{min.Z, max.Z - t}
	for _, x := range xs {
		for _, z := range zs {
			appendEntityBox(m, mathx.V3(x, min.Y, z), mathx.V3(x+t, max.Y, z+t), col)
		}
	}
	for _, y := range ys {
		for _, z := range zs {
			appendEntityBox(m, mathx.V3(min.X, y, z), mathx.V3(max.X, y+t, z+t), col)
		}
	}
	for _, x := range xs {
		for _, y := range ys {
			appendEntityBox(m, mathx.V3(x, y, min.Z), mathx.V3(x+t, y+t, max.Z), col)
		}
	}
}
