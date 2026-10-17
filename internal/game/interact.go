package game

import (
	"math"

	"sevg/internal/core"
	"sevg/internal/i18n"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

type recipe struct {
	Inputs []core.Stack
	Output core.Stack
}

var recipes = []recipe{
	{Inputs: []core.Stack{{Item: core.ItemLog, Count: 1}}, Output: core.Stack{Item: core.ItemPlanks, Count: 4}},
	{Inputs: []core.Stack{{Item: core.ItemPlanks, Count: 2}}, Output: core.Stack{Item: core.ItemStick, Count: 4}},
	{Inputs: []core.Stack{{Item: core.ItemCobble, Count: 4}}, Output: core.Stack{Item: core.ItemStoneBrick, Count: 4}},
	{Inputs: []core.Stack{{Item: core.ItemSand, Count: 4}}, Output: core.Stack{Item: core.ItemGlass, Count: 1}},
	{Inputs: []core.Stack{{Item: core.ItemCobble, Count: 3}, {Item: core.ItemStick, Count: 2}}, Output: core.Stack{Item: core.ItemStonePickaxe, Count: 1}},
	{Inputs: []core.Stack{{Item: core.ItemPlanks, Count: 2}, {Item: core.ItemStick, Count: 1}}, Output: core.Stack{Item: core.ItemWoodSword, Count: 1}},
	{Inputs: []core.Stack{{Item: core.ItemIronIngot, Count: 2}, {Item: core.ItemStick, Count: 1}}, Output: core.Stack{Item: core.ItemIronSword, Count: 1}},
	{Inputs: []core.Stack{{Item: core.ItemIronIngot, Count: 3}, {Item: core.ItemStick, Count: 2}}, Output: core.Stack{Item: core.ItemIronPickaxe, Count: 1}},
}

func (g *Game) handleActions(dt float32) {
	eye := g.player.Eye()
	dir := g.player.LookDir()
	blockHit := g.world.Raycast(eye, dir, 5.8)
	mob, mobDist, mobHit := g.mobs.PickMob(eye, dir, 5.8)
	mobFirst := mobHit && (!blockHit.Hit || mobDist < blockHit.Distance)

	if blockHit.Hit {
		g.breakPos = blockHit.Pos
		g.targetValid = true
	} else {
		g.targetValid = false
		g.breakActive = false
		g.breakProgress = 0
	}

	if g.mouseLeft {
		if mobFirst {
			g.breakProgress = 0
			g.breakActive = false
			if g.attackCool <= 0 {
				dmg := g.selectedDamage()
				if g.mode == ModeCreative {
					dmg = 1000
				}
				g.mobs.DamageMob(mob, dmg, g.player.Body.Pos)
				g.attackCool = 0.32
			}
		} else if blockHit.Hit {
			g.mineBlock(dt, blockHit)
		}
	} else {
		g.breakProgress = 0
		g.breakActive = false
	}

	if g.mouseRight && g.placeCool <= 0 && blockHit.Hit && !mobFirst {
		g.useItem(blockHit)
	}
}

func (g *Game) mineBlock(dt float32, hit world.RayHit) {
	if g.mode == ModeCreative {
		if g.breakCool > 0 {
			return
		}
		b := hit.Block
		if b == core.Air || b == core.Water || b == core.Bedrock {
			return
		}
		x, y, z := hit.Pos[0], hit.Pos[1], hit.Pos[2]
		if g.world.GetBlock(x, y, z) != b {
			return
		}
		g.world.SetBlock(x, y, z, core.Air)
		g.breakCool = 0.12
		g.breakActive = false
		g.breakProgress = 0
		return
	}
	hard := hit.Block.Hardness()
	if hard < 0 {
		g.breakProgress = 0
		return
	}
	if !g.breakActive || g.breakPos != hit.Pos {
		g.breakPos = hit.Pos
		g.breakProgress = 0
		g.breakActive = true
	}
	speed := g.selectedMining()
	if speed <= 0 {
		speed = 1
	}
	seconds := hard / speed
	if seconds < 0.055 {
		seconds = 0.055
	}
	g.breakProgress += dt / seconds
	if g.breakProgress < 1 {
		return
	}
	g.breakProgress = 0
	g.breakActive = false
	b := hit.Block
	if b == core.Air || b == core.Water || b == core.Bedrock {
		return
	}
	x, y, z := hit.Pos[0], hit.Pos[1], hit.Pos[2]
	if g.world.GetBlock(x, y, z) != b {
		return
	}
	g.world.SetBlock(x, y, z, core.Air)
	drop := dropForBlock(b)
	if drop != core.ItemNone {
		pos := mathx.Vec3{X: float32(x) + 0.5, Y: float32(y) + 0.35, Z: float32(z) + 0.5}
		ang := float32(g.frame%360) * 0.0174
		vel := mathx.Vec3{X: float32(math.Cos(float64(ang))) * 1.5, Y: 3.4, Z: float32(math.Sin(float64(ang))) * 1.5}
		g.mobs.Drop(drop, 1, pos, vel)
	}
}

func dropForBlock(b core.BlockID) core.ItemID {
	switch b {
	case core.Leaves, core.Glass, core.Ice, core.Water, core.Bedrock, core.Air:
		return core.ItemNone
	}
	if d := b.Def().Drop; d != core.ItemNone {
		return d
	}
	return core.ItemForBlock(b)
}

func (g *Game) selectedMining() float32 {
	s := g.inv.SelectedStack()
	if s.Empty() {
		return 1
	}
	if v := s.Item.Def().Mining; v > 0 {
		return v
	}
	return 1
}

func (g *Game) selectedDamage() float32 {
	s := g.inv.SelectedStack()
	if s.Empty() {
		return 1.5
	}
	if v := s.Item.Def().Damage; v > 0 {
		return v
	}
	return 1.5
}

func (g *Game) useItem(hit world.RayHit) {
	s := g.inv.SelectedStack()
	if s.Empty() {
		return
	}
	def := s.Item.Def()
	if def.Kind == core.KindFood {
		if g.player.Health >= g.player.MaxHP {
			g.notify(i18n.T("toast.full_health"), 1.2)
			return
		}
		g.player.Heal(def.Food)
		g.inv.TakeSelected(1)
		g.placeCool = 0.35
		g.notify(i18n.T("toast.ate", i18n.ItemName(s.Item)), 1.2)
		return
	}
	if def.Kind != core.KindBlock || def.Block == core.Air {
		return
	}
	nx, ny, nz := int(hit.Normal.X), int(hit.Normal.Y), int(hit.Normal.Z)
	x, y, z := hit.Pos[0]+nx, hit.Pos[1]+ny, hit.Pos[2]+nz
	if y < 0 || y >= world.ChunkHeight {
		return
	}
	cur := g.world.GetBlock(x, y, z)
	if cur != core.Air && cur != core.Water {
		return
	}
	if def.Block.Cross() {
		if !g.world.GetBlock(x, y-1, z).Solid() {
			return
		}
	}
	if g.player.IntersectsBlock(x, y, z) {
		g.notify(i18n.T("toast.no_room"), 1.2)
		return
	}
	for _, m := range g.mobs.Mobs {
		if m.Dead {
			continue
		}
		min, max := m.Bounds()
		if max.X > float32(x) && min.X < float32(x+1) && max.Y > float32(y) && min.Y < float32(y+1) && max.Z > float32(z) && min.Z < float32(z+1) {
			return
		}
	}
	if g.world.SetBlock(x, y, z, def.Block) {
		if g.mode != ModeCreative {
			g.inv.TakeSelected(1)
		}
		g.placeCool = 0.18
	}
}

func (g *Game) pickBlock() {
	if g.pickCool > 0 {
		return
	}
	g.pickCool = 0.25
	hit := g.world.Raycast(g.player.Eye(), g.player.LookDir(), 5.8)
	if !hit.Hit {
		return
	}
	item := core.ItemForBlock(hit.Block)
	if item == core.ItemNone {
		item = dropForBlock(hit.Block)
	}
	if item == core.ItemNone {
		return
	}
	for i := 0; i < core.HotbarSlots; i++ {
		if g.inv.Slots[i].Item == item {
			g.inv.Select(i)
			return
		}
	}
	for i := core.HotbarSlots; i < core.InventorySlots; i++ {
		if g.inv.Slots[i].Item == item {
			g.inv.Swap(i, g.inv.Selected)
			return
		}
	}
	g.notify(i18n.T("toast.no_item", i18n.ItemName(item)), 1.3)
}

func (g *Game) handleUIClick(right bool) {
	switch g.state {
	case StateMenu:
		if right {
			return
		}
		if g.menuLangRect().Contains(g.mouseX, g.mouseY) {
			g.toggleLanguage()
			return
		}
		rects := g.menuRects()
		for i, r := range rects {
			if !r.Contains(g.mouseX, g.mouseY) {
				continue
			}
			switch i {
			case 0:
				if err := g.startWorld(false); err != nil {
					g.notify(i18n.T("toast.could_not_start", err.Error()), 3)
				}
			case 1:
				if g.saveExists() {
					if err := g.startWorld(true); err != nil {
						g.notify(i18n.T("toast.could_not_load", err.Error()), 3)
					}
				}
			case 2:
				g.quit = true
			}
		}
	case StatePaused:
		if right {
			return
		}
		rects := g.pauseRects()
		for i, r := range rects {
			if !r.Contains(g.mouseX, g.mouseY) {
				continue
			}
			switch i {
			case 0:
				g.setState(StatePlaying)
			case 1:
				if err := g.saveGame(); err == nil {
					g.notify(i18n.T("toast.saved"), 1.5)
				}
			case 2:
				g.quitToTitle()
			case 3:
				g.quit = true
			}
		}
	case StateDead:
		if right {
			return
		}
		if g.respawnRect().Contains(g.mouseX, g.mouseY) {
			g.respawn()
		} else if g.deadTitleRect().Contains(g.mouseX, g.mouseY) {
			g.quitToTitle()
		}
	case StateInventory:
		if g.mode == ModeCreative {
			invTab, creativeTab := g.inventoryTabRects()
			if invTab.Contains(g.mouseX, g.mouseY) {
				g.inventoryTab = 0
				return
			}
			if creativeTab.Contains(g.mouseX, g.mouseY) {
				g.inventoryTab = 1
				return
			}
		}
		if g.mode == ModeCreative && g.inventoryTab == 1 {
			if g.creativeCloseRect().Contains(g.mouseX, g.mouseY) {
				g.heldStack = core.Stack{}
				g.setState(StatePlaying)
				return
			}
			for i := 0; i < 9; i++ {
				if g.creativeHotbarRect(i).Contains(g.mouseX, g.mouseY) {
					// Same slot interaction as the survival inventory:
					// pick up, place, merge, split, or swap.
					g.clickSlot(i, right)
					return
				}
			}
			if item := g.creativeItemAt(g.mouseX, g.mouseY); item != core.ItemNone {
				g.giveCreativeItem(item)
			}
			return
		}
		if g.invCloseRect().Contains(g.mouseX, g.mouseY) {
			if !g.heldStack.Empty() {
				left := g.inv.Add(g.heldStack.Item, g.heldStack.Count)
				if left > 0 {
					g.mobs.Drop(g.heldStack.Item, left, g.player.Body.Pos, mathx.Vec3{Y: 3})
				}
				g.heldStack = core.Stack{}
			}
			g.setState(StatePlaying)
			return
		}
		for vis := 0; vis < core.InventorySlots; vis++ {
			if g.invSlotRect(vis).Contains(g.mouseX, g.mouseY) {
				g.clickSlot(visualToSlot(vis), right)
				return
			}
		}
		for i, r := range g.recipeRects() {
			if r.Contains(g.mouseX, g.mouseY) {
				g.craft(i)
				return
			}
		}
	}
}

func (g *Game) clickSlot(idx int, right bool) {
	if idx < 0 || idx >= core.InventorySlots {
		return
	}
	slot := &g.inv.Slots[idx]
	if !right {
		switch {
		case g.heldStack.Empty() && !slot.Empty():
			g.heldStack = *slot
			*slot = core.Stack{}
		case !g.heldStack.Empty() && slot.Empty():
			*slot = g.heldStack
			g.heldStack = core.Stack{}
		case !g.heldStack.Empty() && slot.Item == g.heldStack.Item:
			max := slot.Item.MaxStack()
			n := max - slot.Count
			if n > g.heldStack.Count {
				n = g.heldStack.Count
			}
			slot.Count += n
			g.heldStack.Count -= n
			if g.heldStack.Count <= 0 {
				g.heldStack = core.Stack{}
			}
		case !g.heldStack.Empty():
			slotCopy := *slot
			*slot = g.heldStack
			g.heldStack = slotCopy
		}
		return
	}
	if g.heldStack.Empty() {
		if slot.Empty() {
			return
		}
		n := (slot.Count + 1) / 2
		g.heldStack = core.Stack{Item: slot.Item, Count: n}
		slot.Count -= n
		if slot.Count <= 0 {
			*slot = core.Stack{}
		}
		return
	}
	if slot.Empty() {
		slot.Item = g.heldStack.Item
		slot.Count = 1
		g.heldStack.Count--
	} else if slot.Item == g.heldStack.Item && slot.Count < slot.Item.MaxStack() {
		slot.Count++
		g.heldStack.Count--
	}
	if g.heldStack.Count <= 0 {
		g.heldStack = core.Stack{}
	}
}

func (g *Game) craft(index int) {
	if index < 0 || index >= len(recipes) {
		return
	}
	r := recipes[index]
	for _, in := range r.Inputs {
		if g.inv.Count(in.Item) < in.Count {
			g.notify(i18n.T("toast.missing", i18n.ItemName(in.Item)), 1.4)
			return
		}
	}
	for _, in := range r.Inputs {
		g.inv.Remove(in.Item, in.Count)
	}
	left := g.inv.Add(r.Output.Item, r.Output.Count)
	if left > 0 {
		g.mobs.Drop(r.Output.Item, left, g.player.Body.Pos, mathx.Vec3{Y: 3})
	}
	g.notify(i18n.T("toast.crafted", i18n.ItemName(r.Output.Item)), 1.2)
}
