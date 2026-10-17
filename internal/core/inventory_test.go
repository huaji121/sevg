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
