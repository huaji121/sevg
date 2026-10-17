package render

import (
	"testing"

	"sevg/internal/core"
)

func TestAtlasContainsAllTiles(t *testing.T) {
	atlas := BuildAtlas()
	if len(atlas) != core.AtlasW*core.AtlasH*4 {
		t.Fatalf("atlas size %d", len(atlas))
	}
	// Ice should be a blue tile, not an empty or palette tile.
	cx := core.TileIce % core.AtlasCols
	cy := core.TileIce / core.AtlasCols
	px := cx*core.AtlasCell + core.AtlasPad + 8
	py := cy*core.AtlasCell + core.AtlasPad + 8
	i := (py*core.AtlasW + px) * 4
	if atlas[i+2] <= atlas[i] || atlas[i+3] < 200 {
		t.Fatalf("ice tile looks wrong: %v", atlas[i:i+4])
	}
}

func TestPaletteTilesOpaque(t *testing.T) {
	atlas := BuildAtlas()
	for tile := core.TileWhite; tile < core.TileTileCount; tile++ {
		cx := tile % core.AtlasCols
		cy := tile / core.AtlasCols
		px := cx*core.AtlasCell + core.AtlasPad + 8
		py := cy*core.AtlasCell + core.AtlasPad + 8
		i := (py*core.AtlasW + px) * 4
		if atlas[i+3] != 255 {
			t.Fatalf("palette tile %d alpha %d", tile, atlas[i+3])
		}
	}
}
func TestCrackTilesContainPixels(t *testing.T) {
	atlas := BuildAtlas()
	for tile := core.TileCrack0; tile <= core.TileCrack9; tile++ {
		cx := tile % core.AtlasCols
		cy := tile / core.AtlasCols
		base := cy*core.AtlasCell + core.AtlasPad
		count := 0
		for y := 0; y < core.AtlasTilePx; y++ {
			for x := 0; x < core.AtlasTilePx; x++ {
				i := ((base+y)*core.AtlasW + cx*core.AtlasCell + core.AtlasPad + x) * 4
				if atlas[i+3] > 0 {
					count++
				}
			}
		}
		if count == 0 {
			t.Fatalf("crack tile %d is empty", tile)
		}
	}
}
