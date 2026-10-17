package ui

import "testing"

func TestFontAtlasBuildsGlyphs(t *testing.T) {
	atlas := BuildFontAtlas([]rune("中文ABC"), nil)
	if len(atlas.Pixels) == 0 || atlas.W <= 0 || atlas.H <= 0 {
		t.Fatal("empty font atlas")
	}
	if _, ok := atlas.Glyphs['A']; !ok {
		t.Fatal("missing ASCII glyph")
	}
	if g, ok := atlas.Glyphs['中']; ok {
		if g.Advance <= 0 || g.W <= 0 {
			t.Fatalf("bad CJK glyph: %+v", g)
		}
	} else {
		t.Log("no CJK system font available; fallback atlas is ASCII-only")
	}
	if atlas.WhiteUV[0] <= 0 || atlas.WhiteUV[1] <= 0 {
		t.Fatal("bad white UV")
	}
}
