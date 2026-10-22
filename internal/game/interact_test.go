package game

import (
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/core"
	"sevg/internal/entity"
	"sevg/internal/mathx"
	"sevg/internal/player"
	"sevg/internal/world"
)

func newTestGame(t *testing.T) *Game {
	t.Helper()
	w := world.New(123)
	w.EnsureChunk(world.ChunkPos{})
	p := player.New(mathx.Vec3{X: 8.5, Y: 80, Z: 8.5})
	g := &Game{
		world:         w,
		player:        p,
		mobs:          entity.NewManager(123),
		canv:          nil,
		keys:          DefaultKeybinds(),
		bindingAction: -1,
		dragSlider:    -1,
	}
	g.inv.FillStarter()
	return g
}

func TestMineBlockProducesDrop(t *testing.T) {
	g := newTestGame(t)
	g.world.SetBlock(8, 70, 8, core.Dirt)
	hit := world.RayHit{Hit: true, Block: core.Dirt, Pos: [3]int{8, 70, 8}, Normal: mathx.Vec3{Y: 1}, Distance: 2}
	for i := 0; i < 20; i++ {
		g.mineBlock(0.1, hit)
	}
	if b := g.world.GetBlock(8, 70, 8); b != core.Air {
		t.Fatalf("block not broken: %v", b)
	}
	if g.mobs.ItemCount() != 1 {
		t.Fatalf("expected one item drop, got %d", g.mobs.ItemCount())
	}
	if g.mobs.Items[0].Item != core.ItemDirt {
		t.Fatalf("wrong drop: %v", g.mobs.Items[0].Item)
	}
}

func TestPlaceBlockConsumesItem(t *testing.T) {
	g := newTestGame(t)
	g.world.SetBlock(8, 100, 8, core.Stone)
	g.player.Body.Pos = mathx.Vec3{X: 12.5, Y: 102, Z: 12.5}
	g.inv = core.Inventory{}
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 4}
	hit := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{8, 100, 8}, Normal: mathx.Vec3{Y: 1}}
	g.useItem(hit)
	if b := g.world.GetBlock(8, 101, 8); b != core.Planks {
		t.Fatalf("block not placed: %v", b)
	}
	if got := g.inv.Count(core.ItemPlanks); got != 3 {
		t.Fatalf("inventory count %d", got)
	}
}

func TestDropMapping(t *testing.T) {
	if dropForBlock(core.Grass) != core.ItemDirt {
		t.Fatal("grass should drop dirt")
	}
	if dropForBlock(core.Stone) != core.ItemCobble {
		t.Fatal("stone should drop cobble")
	}
	if dropForBlock(core.Glass) != core.ItemNone {
		t.Fatal("glass should drop nothing")
	}
}
func TestCreativeMiningIsInstantAndDropsNothing(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeCreative
	g.player.Creative = true
	g.world.SetBlock(8, 70, 8, core.Stone)
	hit := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{8, 70, 8}, Normal: mathx.Vec3{Y: 1}, Distance: 2}
	g.mineBlock(0.016, hit)
	if b := g.world.GetBlock(8, 70, 8); b != core.Air {
		t.Fatalf("creative block not broken: %v", b)
	}
	if g.mobs.ItemCount() != 0 {
		t.Fatalf("creative mining dropped %d items", g.mobs.ItemCount())
	}
}

func TestCreativePlacementDoesNotConsume(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeCreative
	g.player.Creative = true
	g.world.SetBlock(8, 70, 8, core.Stone)
	g.player.Body.Pos = mathx.Vec3{X: 12.5, Y: 72, Z: 12.5}
	g.inv = core.Inventory{}
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 1}
	hit := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{8, 70, 8}, Normal: mathx.Vec3{Y: 1}}
	g.useItem(hit)
	if got := g.inv.Count(core.ItemPlanks); got != 1 {
		t.Fatalf("creative placement consumed item, count %d", got)
	}
}

func TestCreativeDamageIsIgnored(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeCreative
	g.player.Creative = true
	before := g.player.Health
	if g.player.Damage(5) {
		t.Fatal("creative player took damage")
	}
	if g.player.Health != before {
		t.Fatalf("health changed: %v -> %v", before, g.player.Health)
	}
}
func TestDoubleSpaceTogglesCreativeFlight(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.mode = ModeCreative
	g.player.Creative = true
	g.onKey(sdl.K_SPACE)
	if g.spaceTapTimer <= 0 {
		t.Fatal("first space tap did not arm the double-tap timer")
	}
	g.onKey(sdl.K_SPACE)
	if !g.player.Flying {
		t.Fatal("double space did not enable flight")
	}
	g.onKey(sdl.K_SPACE)
	g.onKey(sdl.K_SPACE)
	if g.player.Flying {
		t.Fatal("double space did not disable flight")
	}
}
func TestMiningProgressResetsWhenTargetChanges(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeSurvival
	g.world.SetBlock(2, 70, 0, core.Stone)
	g.world.SetBlock(3, 70, 0, core.Stone)
	first := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{2, 70, 0}}
	second := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{3, 70, 0}}
	g.mineBlock(0.2, first)
	before := g.breakProgress
	if before <= 0 || g.breakPos != first.Pos {
		t.Fatalf("first target progress=%f pos=%v", before, g.breakPos)
	}
	g.mineBlock(0.05, second)
	if g.breakPos != second.Pos {
		t.Fatalf("target did not move: %v", g.breakPos)
	}
	if g.breakProgress >= before {
		t.Fatalf("progress was not reset: before=%f after=%f", before, g.breakProgress)
	}
}

func TestMiningProgressResetsWhenBlockChanges(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeSurvival
	g.world.SetBlock(2, 70, 0, core.Stone)
	hit := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{2, 70, 0}}
	g.mineBlock(0.2, hit)
	before := g.breakProgress
	g.world.SetBlock(2, 70, 0, core.Dirt)
	hit.Block = core.Dirt
	g.mineBlock(0.05, hit)
	if g.breakBlock != core.Dirt {
		t.Fatalf("break block did not update: %v", g.breakBlock)
	}
	if g.breakProgress >= before {
		t.Fatalf("progress survived block replacement: before=%f after=%f", before, g.breakProgress)
	}
}
func TestPlacePassThroughBlockInsidePlayer(t *testing.T) {
	g := newTestGame(t)
	g.player.Body.Pos = mathx.Vec3{X: 12.5, Y: 102, Z: 12.5}
	g.inv = core.Inventory{}
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemTorch, Count: 4}
	g.world.SetBlock(12, 101, 12, core.Stone)
	hit := world.RayHit{Hit: true, Block: core.Stone, Pos: [3]int{11, 102, 12}, Normal: mathx.Vec3{X: 1}}
	if !g.player.IntersectsBlock(12, 102, 12) {
		t.Fatal("test target is not inside the player")
	}
	g.useItem(hit)
	if b := g.world.GetBlock(12, 102, 12); b != core.Torch {
		t.Fatalf("pass-through block was not placed inside player: %v", b)
	}
	if got := g.inv.Count(core.ItemTorch); got != 3 {
		t.Fatalf("torch count = %d", got)
	}
}

func TestCreativePickBlockCopiesBlock(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeCreative
	g.player.Creative = true
	g.inv = core.Inventory{}
	g.inv.Selected = 2
	g.world.SetBlock(8, 101, 8, core.Stone)
	g.player.Body.Pos = mathx.Vec3{X: 8.5, Y: 100, Z: 10.5}
	g.player.Yaw = 0
	g.player.Pitch = 0
	g.pickBlock()
	got := g.inv.Slots[2]
	if got.Item != core.ItemStone || got.Count != core.ItemStone.MaxStack() {
		t.Fatalf("creative pick block = %+v", got)
	}
}

func TestCreativePickBlockWorksOnLeaves(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeCreative
	g.player.Creative = true
	g.inv = core.Inventory{}
	g.inv.Selected = 0
	g.world.SetBlock(8, 101, 8, core.Leaves)
	g.player.Body.Pos = mathx.Vec3{X: 8.5, Y: 100, Z: 10.5}
	g.player.Yaw = 0
	g.player.Pitch = 0
	g.pickBlock()
	if got := g.inv.Slots[0]; got.Item != core.ItemLeaves {
		t.Fatalf("creative pick on leaves = %+v", got)
	}
}

func TestSurvivalPickBlockNeedsTheItem(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeSurvival
	g.player.Creative = false
	g.inv = core.Inventory{}
	g.inv.Selected = 2
	g.world.SetBlock(8, 101, 8, core.Stone)
	g.player.Body.Pos = mathx.Vec3{X: 8.5, Y: 100, Z: 10.5}
	g.player.Yaw = 0
	g.player.Pitch = 0
	g.pickBlock()
	if !g.inv.Slots[2].Empty() {
		t.Fatalf("survival pick block created an item: %+v", g.inv.Slots[2])
	}
}

func TestEatingTakesTime(t *testing.T) {
	g := newTestGame(t)
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemApple, Count: 2}
	g.player.Food = 8
	g.player.Saturation = 0
	before := g.inv.Count(core.ItemApple)

	// A single frame of use must not consume the item.
	g.updateEating(0.05, core.ItemApple)
	if g.inv.Count(core.ItemApple) != before {
		t.Fatal("item was consumed instantly")
	}
	if len(g.particles) == 0 {
		t.Fatal("no chewing particles spawned")
	}
	if !g.useActive {
		t.Fatal("use is not active while holding the item")
	}

	// Holding until the use time elapses finishes the first apple.
	for guard := 0; g.useActive && guard < 200; guard++ {
		g.updateEating(0.05, core.ItemApple)
	}
	if g.useActive {
		t.Fatal("use did not finish after the use time elapsed")
	}
	if g.player.Food <= 8 {
		t.Fatalf("eating did not restore hunger: %f", g.player.Food)
	}
	if got := g.inv.Count(core.ItemApple); got != before-1 {
		t.Fatalf("apple count = %d, want %d", got, before-1)
	}
}

func TestEatingBlockedWhenFull(t *testing.T) {
	g := newTestGame(t)
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemApple, Count: 2}
	g.player.Food = g.player.MaxFood
	g.player.Health = 10
	before := g.inv.Count(core.ItemApple)
	for i := 0; i < 10; i++ {
		g.updateEating(0.05, core.ItemApple)
	}
	if got := g.inv.Count(core.ItemApple); got != before {
		t.Fatalf("apple consumed while full: %d", got)
	}
	if g.useActive {
		t.Fatal("use should not start while full")
	}
}

func TestSolidBlockStillRequiresSpace(t *testing.T) {
	g := newTestGame(t)
	g.player.Body.Pos = mathx.Vec3{X: 12.5, Y: 102, Z: 12.5}
	g.inv = core.Inventory{}
	g.inv.Selected = 0
	g.inv.Slots[0] = core.Stack{Item: core.ItemPlanks, Count: 4}
	hit := world.RayHit{Hit: true, Block: core.Dirt, Pos: [3]int{11, 102, 12}, Normal: mathx.Vec3{X: 1}}
	g.useItem(hit)
	if b := g.world.GetBlock(12, 102, 12); b != core.Air {
		t.Fatalf("solid block was placed inside player: %v", b)
	}
	if got := g.inv.Count(core.ItemPlanks); got != 4 {
		t.Fatalf("planks count = %d", got)
	}
}
func TestUnsupportedPlantDropsItem(t *testing.T) {
	g := newTestGame(t)
	g.world.SetBlock(8, 100, 8, core.Stone)
	g.world.SetBlock(8, 101, 8, core.Flower)
	g.world.TakeBlockBreaks()
	g.world.SetBlock(8, 100, 8, core.Air)
	g.processBlockBreaks()
	if g.mobs.ItemCount() != 1 {
		t.Fatalf("plant drop count = %d", g.mobs.ItemCount())
	}
	if got := g.mobs.Items[0].Item; got != core.ItemFlower {
		t.Fatalf("plant drop = %v", got)
	}
}
func TestCreativePlantSupportBreakDoesNotDrop(t *testing.T) {
	g := newTestGame(t)
	g.mode = ModeCreative
	g.world.SetBlock(8, 110, 8, core.Stone)
	g.world.SetBlock(8, 111, 8, core.TallGrass)
	g.world.TakeBlockBreaks()
	g.world.SetBlock(8, 110, 8, core.Air)
	g.processBlockBreaks()
	if g.mobs.ItemCount() != 0 {
		t.Fatalf("creative support break dropped %d items", g.mobs.ItemCount())
	}
}
