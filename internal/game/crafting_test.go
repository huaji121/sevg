package game

import (
	"testing"

	"sevg/internal/core"
)

func TestTwoByTwoCraftingGrid(t *testing.T) {
	g := newTestGame(t)
	g.inv = core.Inventory{}
	g.craftGrid[0] = core.Stack{Item: core.ItemLog, Count: 1}
	g.takeCraftOutput()
	if got := g.inv.Count(core.ItemPlanks); got != 4 {
		t.Fatalf("planks = %d", got)
	}
	if !g.craftGrid[0].Empty() {
		t.Fatalf("grid not consumed: %+v", g.craftGrid)
	}
}

func TestRecipeBookOnlyTwoByTwo(t *testing.T) {
	g := newTestGame(t)
	for _, r := range g.recipeBookRecipes() {
		if !recipeFits2x2(r) {
			t.Fatalf("recipe book contains 3x3 recipe %+v", r)
		}
	}
	for _, r := range g.recipeBookRecipes() {
		if r.Output.Item == core.ItemStonePickaxe || r.Output.Item == core.ItemIronPickaxe {
			t.Fatalf("pickaxe should require a crafting table: %+v", r)
		}
	}
}

func TestRecipeBookFill(t *testing.T) {
	g := newTestGame(t)
	g.inv = core.Inventory{}
	g.inv.Add(core.ItemLog, 2)
	if !g.fillCraftingRecipe(0) {
		t.Fatal("could not fill planks recipe")
	}
	if g.craftGrid[0].Item != core.ItemLog || g.craftGrid[0].Count != 1 {
		t.Fatalf("craft grid = %+v", g.craftGrid)
	}
	if got := g.inv.Count(core.ItemLog); got != 1 {
		t.Fatalf("remaining log = %d", got)
	}
	g.takeCraftOutput()
	if got := g.inv.Count(core.ItemPlanks); got != 4 {
		t.Fatalf("planks = %d", got)
	}
}
func TestCraftingConsumesFullInputCount(t *testing.T) {
	g := newTestGame(t)
	g.inv = core.Inventory{}
	g.craftGrid[0] = core.Stack{Item: core.ItemCobble, Count: 4}
	g.takeCraftOutput()
	if got := g.inv.Count(core.ItemStoneBrick); got != 4 {
		t.Fatalf("stone bricks = %d", got)
	}
	if !g.craftGrid[0].Empty() {
		t.Fatalf("cobble remaining: %+v", g.craftGrid[0])
	}
}
func TestCraftingWithStackLargerThanRecipe(t *testing.T) {
	g := newTestGame(t)
	g.inv = core.Inventory{}
	g.craftGrid[0] = core.Stack{Item: core.ItemLog, Count: 2}
	g.takeCraftOutput()
	if got := g.inv.Count(core.ItemPlanks); got != 4 {
		t.Fatalf("planks = %d", got)
	}
	if g.craftGrid[0].Item != core.ItemLog || g.craftGrid[0].Count != 1 {
		t.Fatalf("remaining stack = %+v", g.craftGrid[0])
	}
	if _, ok := g.matchedGridRecipe(); !ok {
		t.Fatal("remaining log should still match the recipe")
	}
}

func TestCraftingWithExtraItemDoesNotMatch(t *testing.T) {
	g := newTestGame(t)
	g.craftGrid[0] = core.Stack{Item: core.ItemLog, Count: 2}
	g.craftGrid[1] = core.Stack{Item: core.ItemSand, Count: 1}
	if _, ok := g.matchedGridRecipe(); ok {
		t.Fatal("unrelated item should prevent a match")
	}
}
