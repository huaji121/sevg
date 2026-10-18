package render

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
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
	for tile := core.TileWhite; tile < core.TileSun; tile++ {
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
func TestCustomSkinOverride(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "skin.png")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	old := customSkinPath
	customSkinPath = path
	defer func() { customSkinPath = old }()

	atlas := buildAtlas()
	tile := int(core.TilePlayerSkinMap)
	cx, cy := tile%core.AtlasCols, tile/core.AtlasCols
	i := ((cy*core.AtlasCell+core.AtlasPad)*core.AtlasW + cx*core.AtlasCell + core.AtlasPad) * 4
	if atlas[i] != 10 || atlas[i+1] != 20 || atlas[i+2] != 30 || atlas[i+3] != 255 {
		t.Fatalf("custom skin pixel = %v", atlas[i:i+4])
	}
}
