package ui

import (
	"testing"

	"sevg/internal/core"
)

func TestBuildItemIcons(t *testing.T) {
	icons := buildItemIcons()
	items := []core.ItemID{
		core.ItemPlanks, core.ItemGlass, core.ItemTallGrass,
		core.ItemStonePickaxe, core.ItemWoodSword, core.ItemApple,
		core.ItemCoal, core.ItemStick, core.ItemIronIngot,
	}
	for _, item := range items {
		img := icons[item]
		if img == nil {
			t.Fatalf("missing icon for %v", item)
		}
		if img.Rect.Dx() != itemIconSize || img.Rect.Dy() != itemIconSize {
			t.Fatalf("icon %v has size %v", item, img.Rect)
		}
		opaque := false
		for i := 3; i < len(img.Pix); i += 4 {
			if img.Pix[i] != 0 {
				opaque = true
				break
			}
		}
		if !opaque {
			t.Fatalf("icon %v is fully transparent", item)
		}
	}
}

func TestFontAtlasPacksIconsAndSpaces(t *testing.T) {
	fa := BuildFontAtlas([]rune("A中"))
	if _, ok := fa.Icons[core.ItemPlanks]; !ok {
		t.Fatal("plank icon was not packed into the font atlas")
	}
	if len(fa.IconSides[core.ItemPlanks]) == 0 {
		t.Fatal("plank silhouette sides were not recorded")
	}
	if _, ok := fa.Glyphs[' ']; !ok {
		t.Fatal("space glyph missing from atlas")
	}
	if g, ok := fa.Glyphs['A']; !ok || g.U1 <= g.U0 || g.V1 <= g.V0 {
		t.Fatalf("A glyph has invalid UVs: %+v", g)
	}
}
