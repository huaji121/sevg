package game

import (
	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
)

const craftGridSize = 4

func recipeFits2x2(r recipe) bool {
	total := 0
	for _, in := range r.Inputs {
		total += in.Count
		if total > craftGridSize {
			return false
		}
	}
	return len(r.Inputs) <= craftGridSize
}

func (g *Game) recipeBookRecipes() []recipe {
	out := make([]recipe, 0, len(recipes))
	for _, r := range recipes {
		if recipeFits2x2(r) {
			out = append(out, r)
		}
	}
	return out
}

func (g *Game) gridCounts() map[core.ItemID]int {
	counts := make(map[core.ItemID]int)
	for _, st := range g.craftGrid {
		if !st.Empty() {
			counts[st.Item] += st.Count
		}
	}
	return counts
}

func (g *Game) matchedGridRecipe() (recipe, bool) {
	counts := g.gridCounts()
	if len(counts) == 0 {
		return recipe{}, false
	}
	for _, r := range recipes {
		if !recipeFits2x2(r) {
			continue
		}
		want := make(map[core.ItemID]int)
		for _, in := range r.Inputs {
			want[in.Item] += in.Count
		}
		if len(want) != len(counts) {
			continue
		}
		ok := true
		for item, n := range want {
			if counts[item] < n {
				ok = false
				break
			}
		}
		if ok {
			return r, true
		}
	}
	return recipe{}, false
}

func (g *Game) takeCraftOutput() {
	r, ok := g.matchedGridRecipe()
	if !ok {
		return
	}
	for _, in := range r.Inputs {
		need := in.Count
		for i := range g.craftGrid {
			if need <= 0 {
				break
			}
			st := &g.craftGrid[i]
			if st.Item != in.Item || st.Count <= 0 {
				continue
			}
			take := st.Count
			if take > need {
				take = need
			}
			st.Count -= take
			need -= take
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
	slot := 0
	for _, in := range r.Inputs {
		g.inv.Remove(in.Item, in.Count)
		g.craftGrid[slot] = core.Stack{Item: in.Item, Count: in.Count}
		slot++
	}
	g.recipeBookOpen = false
	return true
}
