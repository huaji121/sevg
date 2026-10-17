package game

import (
	"fmt"

	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/ui"
)

func (g *Game) craftSlotRect(i int) ui.Rect {
	p := g.invPanelRect()
	n := g.craftGridN()
	return ui.Rect{X: p.X + 486 + float32(i%n)*50, Y: p.Y + 82 + float32(i/n)*50, W: 42, H: 42}
}

func (g *Game) craftOutputRect() ui.Rect {
	p := g.invPanelRect()
	n := g.craftGridN()
	return ui.Rect{X: p.X + 486 + float32(n*50+18), Y: p.Y + 82 + float32(n*50-50)/2, W: 42, H: 42}
}

func (g *Game) recipeBookButtonRect() ui.Rect {
	p := g.invPanelRect()
	n := g.craftGridN()
	return ui.Rect{X: p.X + 486, Y: p.Y + 82 + float32(n*50+22), W: 210, H: 38}
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
	const cols = 2
	rows := (len(list) + cols - 1) / cols
	gap := float32(10)
	colW := (p.W - 40 - gap) / 2
	for i := range out {
		col := i / rows
		row := i % rows
		x := p.X + 20 + float32(col)*(colW+gap)
		out[i] = ui.Rect{X: x, Y: p.Y + 66 + float32(row)*44, W: colW, H: 40}
	}
	return out
}

func (g *Game) drawCraftingColumn(c *ui.Canvas, p ui.Rect) {
	n := g.craftGridN()
	title := i18n.T("inventory.crafting")
	sub := i18n.T("inventory.handmade")
	if n == 3 {
		title = i18n.T("inventory.crafting_table")
		sub = i18n.T("inventory.crafting_table_hint")
	}
	c.Text(p.X+486, p.Y+14, title, 1.8, ui.Accent)
	c.Text(p.X+486, p.Y+47, sub, 0.8, ui.MutedColor)

	out := core.Stack{}
	if r, ok := g.matchedGridRecipe(); ok {
		out = r.Output
	}
	for i := 0; i < n*n; i++ {
		r := g.craftSlotRect(i)
		bg := ui.RGBA(16, 20, 24, 220)
		bd := ui.RGBA(70, 82, 90, 200)
		if r.Contains(g.mouseX, g.mouseY) {
			bg = ui.RGBA(40, 47, 53, 235)
			bd = ui.Accent
		}
		g.drawInventorySlot(r, g.craftGrid[i], bg, bd, 0)
	}
	or := g.craftOutputRect()
	c.Text(p.X+486+float32(n*50)+2, or.Y+7, ">", 1.8, ui.MutedColor)
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
	bookHint := i18n.T("inventory.recipe_hint")
	if n == 3 {
		bookHint = i18n.T("inventory.recipe_hint_table")
	}
	c.Text(p.X+486, g.recipeBookButtonRect().Y+46, bookHint, 0.72, ui.MutedColor)

	if !g.recipeBookOpen {
		tip1, tip2 := i18n.T("inventory.tip1"), i18n.T("inventory.tip2")
		if n == 3 {
			tip1, tip2 = i18n.T("inventory.tip1_table"), i18n.T("inventory.tip2_table")
		}
		c.Text(p.X+486, p.Y+424, tip1, 0.8, ui.MutedColor)
		c.Text(p.X+486, p.Y+440, tip2, 0.8, ui.MutedColor)
	}
}

func (g *Game) drawRecipeBook(w, h float32) {
	c := g.canv
	p := g.recipeBookPanelRect()
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.68})
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(20, 24, 28, 250), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+20, p.Y+14, i18n.T("inventory.recipe_book"), 1.8, ui.TextColor)
	hint := i18n.T("inventory.recipe_hint")
	if g.craftSize == 3 {
		hint = i18n.T("inventory.recipe_hint_table")
	}
	c.Text(p.X+20, p.Y+44, hint, 0.8, ui.MutedColor)
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
		g.drawItemIcon(r.X+5, r.Y+5, 30, rec.Output, 1)
		c.Text(r.X+40, r.Y+4, i18n.ItemName(rec.Output.Item), 0.88, ui.TextColor)
		c.Text(r.X+40, r.Y+20, recipeInputs(rec), 0.60, ui.MutedColor)
		c.Text(r.X+r.W-34, r.Y+12, fmt.Sprintf("x%d", rec.Output.Count), 0.8, ui.Accent)
	}
}
