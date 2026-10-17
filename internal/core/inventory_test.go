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
