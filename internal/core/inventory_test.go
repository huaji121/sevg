package core

import "testing"

func TestInventoryAddRemove(t *testing.T) {
	var inv Inventory
	left := inv.Add(ItemCobble, 150)
	if left != 0 {
		t.Fatalf("leftover %d", left)
	}
	if inv.Count(ItemCobble) != 150 {
		t.Fatalf("count %d", inv.Count(ItemCobble))
	}
	inv.Remove(ItemCobble, 60)
	if inv.Count(ItemCobble) != 90 {
		t.Fatalf("count after remove %d", inv.Count(ItemCobble))
	}
}

func TestInventoryMaxStack(t *testing.T) {
	var inv Inventory
	inv.Add(ItemWoodSword, 3)
	if inv.Count(ItemWoodSword) != 3 {
		t.Fatalf("swords should occupy separate slots, got %d", inv.Count(ItemWoodSword))
	}
	nonEmpty := 0
	for _, s := range inv.Slots {
		if !s.Empty() {
			nonEmpty++
			if s.Count != 1 {
				t.Fatalf("sword stack has %d", s.Count)
			}
		}
	}
	if nonEmpty != 3 {
		t.Fatalf("expected 3 slots, got %d", nonEmpty)
	}
}

func TestItemForBlock(t *testing.T) {
	if ItemForBlock(Grass) != ItemGrass {
		t.Fatal("grass item mismatch")
	}
	if ItemForBlock(Stone) != ItemStone {
		t.Fatal("stone item mismatch")
	}
	if ItemForBlock(Water) != ItemNone {
		t.Fatal("water should not be an item")
	}
	if ItemForBlock(Wool) != ItemWool {
		t.Fatal("wool item mismatch")
	}
	if ItemForBlock(Leaves) != ItemLeaves {
		t.Fatal("leaves should have a placeable item form")
	}
}
func TestBlockFaceTiles(t *testing.T) {
	grass := BlockDefs[Grass]
	if grass.Tiles[2] != TileGrassTop {
		t.Fatalf("grass top tile = %d", grass.Tiles[2])
	}
	if grass.Tiles[3] != TileDirt {
		t.Fatalf("grass bottom tile = %d", grass.Tiles[3])
	}
	for _, side := range []int{0, 1, 4, 5} {
		if grass.Tiles[side] != TileGrassSide {
			t.Fatalf("grass side %d tile = %d", side, grass.Tiles[side])
		}
	}
	if Grass.Tile(2) != TileGrassTop || Grass.Tile(3) != TileDirt || Grass.Tile(4) != TileGrassSide {
		t.Fatal("BlockID.Tile does not select the face tile")
	}
	if Log.Tile(2) != TileLogTop || Log.Tile(0) != TileLogSide {
		t.Fatal("log face tiles are wrong")
	}
}
func TestStackLimitIs64(t *testing.T) {
	if got := ItemPlanks.MaxStack(); got != 64 {
		t.Fatalf("planks max stack = %d", got)
	}
	if got := ItemApple.MaxStack(); got != 64 {
		t.Fatalf("apple max stack = %d", got)
	}
	if got := ItemWoodSword.MaxStack(); got != 1 {
		t.Fatalf("sword max stack = %d", got)
	}
	var inv Inventory
	if left := inv.Add(ItemPlanks, 100); left != 0 {
		t.Fatalf("leftover = %d", left)
	}
	if got := inv.Count(ItemPlanks); got != 100 {
		t.Fatalf("count = %d", got)
	}
	maxSeen := 0
	nonEmpty := 0
	for _, s := range inv.Slots {
		if s.Item == ItemPlanks {
			if s.Count > maxSeen {
				maxSeen = s.Count
			}
			nonEmpty++
		}
	}
	if maxSeen != 64 || nonEmpty != 2 {
		t.Fatalf("planks slots = %d, max = %d", nonEmpty, maxSeen)
	}
}
func TestNormalizeSplitsLegacyOverstacks(t *testing.T) {
	var inv Inventory
	inv.Slots[0] = Stack{Item: ItemPlanks, Count: 99}
	inv.Normalize()
	if got := inv.Count(ItemPlanks); got != 99 {
		t.Fatalf("count after normalize = %d", got)
	}
	for _, s := range inv.Slots {
		if s.Item == ItemPlanks && s.Count > 64 {
			t.Fatalf("overstack remains: %+v", s)
		}
	}
}
