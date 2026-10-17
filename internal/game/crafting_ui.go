package game

import (
	"fmt"

	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/ui"
)

func (g *Game) craftSlotRect(i int) ui.Rect {
	p := g.invPanelRect()
	cx, cy := p.X+500, p.Y+78
	return ui.Rect{X: cx + float32(i%2)*50, Y: cy + float32(i/2)*50, W: 42, H: 42}
}

func (g *Game) craftOutputRect() ui.Rect {
	p := g.invPanelRect()
	return ui.Rect{X: p.X + 638, Y: p.Y + 103, W: 42, H: 42}
}

func (g *Game) recipeBookButtonRect() ui.Rect {
	p := g.invPanelRect()
	return ui.Rect{X: p.X + 486, Y: p.Y + 212, W: 210, H: 38}
}

func (g *Game) recipeBookPanelRect() ui.Rect {
	p := g.invPanelRect()
	w := float32(620)
	h := float32(430)
	if w > p.W-20 {
		w = p.W - 20
	}
	if h > p.H-20 {
		h = p.H - 20
	}
	return ui.Rect{X: p.X + (p.W-w)/2, Y: p.Y + (p.H-h)/2, W: w, H: h}
}

func (g *Game) recipeBookCloseRect() ui.Rect {
	p := g.recipeBookPanelRect()
	return ui.Rect{X: p.X + p.W - 36, Y: p.Y + 12, W: 24, H: 24}
}

func (g *Game) recipeBookRowRects() []ui.Rect {
	p := g.recipeBookPanelRect()
	list := g.recipeBookRecipes()
	out := make([]ui.Rect, len(list))
	for i := range out {
		out[i] = ui.Rect{X: p.X + 20, Y: p.Y + 66 + float32(i)*44, W: p.W - 40, H: 40}
	}
	return out
}

func (g *Game) drawCraftingColumn(c *ui.Canvas, p ui.Rect) {
	cx := p.X + 486
	c.Text(cx, p.Y+14, i18n.T("inventory.crafting"), 1.8, ui.Accent)
	c.Text(cx, p.Y+47, i18n.T("inventory.handmade"), 0.8, ui.MutedColor)

	out := core.Stack{}
	if r, ok := g.matchedGridRecipe(); ok {
		out = r.Output
	}
	for i := 0; i < 4; i++ {
		r := g.craftSlotRect(i)
		bg := ui.RGBA(16, 20, 24, 220)
		bd := ui.RGBA(70, 82, 90, 200)
		if r.Contains(g.mouseX, g.mouseY) {
			bg = ui.RGBA(40, 47, 53, 235)
			bd = ui.Accent
		}
		g.drawInventorySlot(r, g.craftGrid[i], bg, bd, 0)
	}
	c.Text(cx+101, p.Y+102, ">", 1.8, ui.MutedColor)
	or := g.craftOutputRect()
	bd := ui.RGBA(90, 105, 115, 230)
	bg := ui.RGBA(24, 30, 34, 230)
	if !out.Empty() {
		bd = ui.Accent
		bg = ui.RGBA(42, 48, 40, 240)
	}
	if or.Contains(g.mouseX, g.mouseY) {
		bg = ui.RGBA(50, 58, 62, 245)
	}
	g.drawInventorySlot(or, out, bg, bd, 0)

	g.drawButtonScale(g.recipeBookButtonRect(), i18n.T("inventory.recipe_book"), true, 1.1)
	c.Text(cx, p.Y+266, i18n.T("inventory.recipe_hint"), 0.72, ui.MutedColor)

	if !g.recipeBookOpen {
		c.Text(cx, p.Y+424, i18n.T("inventory.tip1"), 0.8, ui.MutedColor)
		c.Text(cx, p.Y+440, i18n.T("inventory.tip2"), 0.8, ui.MutedColor)
	}
}

func (g *Game) drawRecipeBook(w, h float32) {
	c := g.canv
	p := g.recipeBookPanelRect()
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.68})
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(20, 24, 28, 250), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+20, p.Y+14, i18n.T("inventory.recipe_book"), 1.8, ui.TextColor)
	c.Text(p.X+20, p.Y+44, i18n.T("inventory.recipe_hint"), 0.8, ui.MutedColor)
	closeR := g.recipeBookCloseRect()
	c.Rect(closeR.X, closeR.Y, closeR.W, closeR.H, ui.RGBA(180, 60, 60, 230))
	g.drawCenteredLabel(closeR, "X", 1.4, ui.TextColor)

	for i, r := range g.recipeBookRowRects() {
		rec := g.recipeBookRecipes()[i]
		ok := g.canCraft(rec)
		bg := ui.RGBA(32, 38, 43, 235)
		bd := ui.RGBA(80, 94, 104, 220)
		if ok {
			bg = ui.RGBA(38, 52, 43, 240)
			bd = ui.RGBA(100, 160, 105, 220)
		}
		if r.Contains(g.mouseX, g.mouseY) {
			bg = ui.RGBA(54, 64, 58, 245)
			bd = ui.Accent
		}
		c.Rect(r.X, r.Y, r.W, r.H, bg)
		c.Border(r.X, r.Y, r.W, r.H, 1, bd)
		g.drawItemIcon(r.X+6, r.Y+5, 30, rec.Output, 1)
		c.Text(r.X+44, r.Y+5, i18n.ItemName(rec.Output.Item), 1.0, ui.TextColor)
		c.Text(r.X+44, r.Y+21, recipeInputs(rec), 0.68, ui.MutedColor)
		c.Text(r.X+r.W-38, r.Y+10, fmt.Sprintf("x%d", rec.Output.Count), 0.9, ui.Accent)
	}
}
