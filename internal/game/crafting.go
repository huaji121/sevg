package game

import (
	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
)

func recipeWidth(r recipe) int {
	w := 0
	for _, row := range r.Pattern {
		if len([]rune(row)) > w {
			w = len([]rune(row))
		}
	}
	return w
}

func recipeHeight(r recipe) int { return len(r.Pattern) }

func recipeFitsGrid(r recipe, size int) bool {
	return recipeWidth(r) <= size && recipeHeight(r) <= size
}

func recipeFits2x2(r recipe) bool { return recipeFitsGrid(r, 2) }

func (g *Game) recipeBookRecipes() []recipe {
	size := g.craftSize
	if size < 2 {
		size = 2
	}
	out := make([]recipe, 0, len(recipes))
	for _, r := range recipes {
		if recipeFitsGrid(r, size) {
			out = append(out, r)
		}
	}
	return out
}

func (g *Game) craftGridN() int {
	if g.craftSize == 3 {
		return 3
	}
	return 2
}

// recipeAt checks whether the shaped recipe occupies the grid at the given
// offset. Stack sizes above one are allowed in required cells; each occupied
// pattern cell consumes one item.
func (g *Game) recipeAt(r recipe, ox, oy int) bool {
	n := g.craftGridN()
	w, h := recipeWidth(r), recipeHeight(r)
	rows := make([][]rune, 0, h)
	for _, row := range r.Pattern {
		rows = append(rows, []rune(row))
	}
	for gy := 0; gy < n; gy++ {
		for gx := 0; gx < n; gx++ {
			st := g.craftGrid[gy*n+gx]
			ch := rune(' ')
			if gy >= oy && gy < oy+h && gx >= ox && gx < ox+w {
				row := rows[gy-oy]
				if gx-ox < len(row) {
					ch = row[gx-ox]
				}
			}
			item, required := r.Key[ch]
			if !required || ch == ' ' {
				if !st.Empty() {
					return false
				}
				continue
			}
			if st.Item != item || st.Count < 1 {
				return false
			}
		}
	}
	return true
}

func (g *Game) matchedGridRecipeAt() (recipe, int, int, bool) {
	n := g.craftGridN()
	for _, r := range recipes {
		if !recipeFitsGrid(r, n) {
			continue
		}
		for oy := 0; oy+recipeHeight(r) <= n; oy++ {
			for ox := 0; ox+recipeWidth(r) <= n; ox++ {
				if g.recipeAt(r, ox, oy) {
					return r, ox, oy, true
				}
			}
		}
	}
	return recipe{}, 0, 0, false
}

func (g *Game) matchedGridRecipe() (recipe, bool) {
	r, _, _, ok := g.matchedGridRecipeAt()
	return r, ok
}

func (g *Game) takeCraftOutput() {
	r, ox, oy, ok := g.matchedGridRecipeAt()
	if !ok {
		return
	}
	n := g.craftGridN()
	rows := make([][]rune, 0, len(r.Pattern))
	for _, row := range r.Pattern {
		rows = append(rows, []rune(row))
	}
	for gy, row := range rows {
		for gx, ch := range row {
			if ch == ' ' {
				continue
			}
			if _, required := r.Key[ch]; !required {
				continue
			}
			idx := (oy+gy)*n + (ox + gx)
			if idx < 0 || idx >= len(g.craftGrid) {
				continue
			}
			st := &g.craftGrid[idx]
			if st.Count > 0 {
				st.Count--
			}
			if st.Count <= 0 {
				*st = core.Stack{}
			}
		}
	}
	left := g.inv.Add(r.Output.Item, r.Output.Count)
	if left > 0 && g.mobs != nil && g.player != nil {
		g.mobs.Drop(r.Output.Item, left, g.player.Body.Pos, mathx.Vec3{Y: 3})
	}
	g.playSound("craft")
	g.notify(i18n.T("toast.crafted", i18n.ItemName(r.Output.Item)), 1.2)
}

func (g *Game) returnCraftGrid() {
	for i := range g.craftGrid {
		st := g.craftGrid[i]
		if st.Empty() {
			continue
		}
		left := g.inv.Add(st.Item, st.Count)
		if left > 0 && g.mobs != nil && g.player != nil {
			g.mobs.Drop(st.Item, left, g.player.Body.Pos, mathx.Vec3{Y: 3})
		}
		g.craftGrid[i] = core.Stack{}
	}
}

func (g *Game) fillCraftingRecipe(index int) bool {
	list := g.recipeBookRecipes()
	if index < 0 || index >= len(list) {
		return false
	}
	r := list[index]
	g.returnCraftGrid()
	for _, in := range r.Inputs {
		if g.inv.Count(in.Item) < in.Count {
			g.notify(i18n.T("toast.missing", i18n.ItemName(in.Item)), 1.4)
			return false
		}
	}
	n := g.craftGridN()
	for gy, row := range r.Pattern {
		for gx, ch := range row {
			if ch == ' ' {
				continue
			}
			item, ok := r.Key[ch]
			if !ok {
				continue
			}
			if g.inv.Remove(item, 1) > 0 {
				idx := gy*n + gx
				if idx >= 0 && idx < len(g.craftGrid) {
					g.craftGrid[idx] = core.Stack{Item: item, Count: 1}
				}
			}
		}
	}
	g.recipeBookOpen = false
	return true
}
