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
