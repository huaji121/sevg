package render

import (
	"fmt"
	"image/png"
	"sync"

	"sevg/assets"
	"sevg/internal/core"
)

// Block textures are kept as individual PNG files in assets/textures/blocks.
// The order here matches the core.Tile* constants.
var blockTextureFiles = [core.TileTileCount]string{
	"grass_top", "grass_side", "dirt", "stone", "cobble", "sand", "water",
	"log_side", "log_top", "leaves", "planks", "coal_ore", "iron_ore",
	"snow", "snow_side", "cactus_side", "cactus_top", "glass", "bedrock",
	"gravel", "flower", "tall_grass", "brick", "stone_brick", "ice",
	"obsidian", "wool", "torch", "crafting_table",
	"crack_0", "crack_1", "crack_2", "crack_3", "crack_4",
	"crack_5", "crack_6", "crack_7", "crack_8", "crack_9",
	"white", "pig_skin", "dark", "red", "gold", "gray", "black", "brown",
	"pink", "green", "player_skin", "player_shirt", "player_pants", "player_hair",
}

// BuildAtlas assembles the embedded tile textures into the runtime atlas and
// fills the gutters by clamping to the nearest tile edge.
var (
	atlasOnce sync.Once
	atlasData []byte
)

func BuildAtlas() []byte {
	atlasOnce.Do(func() { atlasData = buildAtlas() })
	return atlasData
}

func buildAtlas() []byte {
	const size = core.AtlasW * core.AtlasH
	img := make([]byte, size*4)
	for tile, name := range blockTextureFiles {
		path := "textures/blocks/" + name + ".png"
		f, err := assets.Textures.Open(path)
		if err != nil {
			panic(fmt.Sprintf("open texture %s: %v", path, err))
		}
		src, err := png.Decode(f)
		f.Close()
		if err != nil {
			panic(fmt.Sprintf("decode texture %s: %v", path, err))
		}
		cx, cy := tile%core.AtlasCols, tile/core.AtlasCols
		tileX := cx*core.AtlasCell + core.AtlasPad
		tileY := cy*core.AtlasCell + core.AtlasPad
		for y := 0; y < core.AtlasTilePx; y++ {
			for x := 0; x < core.AtlasTilePx; x++ {
				r, g, b, a := src.At(x, y).RGBA()
				i := ((tileY+y)*core.AtlasW + tileX + x) * 4
				img[i], img[i+1], img[i+2], img[i+3] =
					byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8)
			}
		}
	}

	// Fill the gutters with clamped edge pixels to avoid atlas bleeding.
	for tile := 0; tile < core.TileTileCount; tile++ {
		cx := tile % core.AtlasCols
		cy := tile / core.AtlasCols
		tileX := cx * core.AtlasCell
		tileY := cy * core.AtlasCell
		at := func(x, y int) [4]byte {
			x = clamp(x, 0, core.AtlasTilePx-1)
			y = clamp(y, 0, core.AtlasTilePx-1)
			i := ((tileY+core.AtlasPad+y)*core.AtlasW + tileX + core.AtlasPad + x) * 4
			return [4]byte{img[i], img[i+1], img[i+2], img[i+3]}
		}
		for y := 0; y < core.AtlasCell; y++ {
			for x := 0; x < core.AtlasCell; x++ {
				c := at(x-core.AtlasPad, y-core.AtlasPad)
				i := ((tileY+y)*core.AtlasW + tileX + x) * 4
				img[i], img[i+1], img[i+2], img[i+3] = c[0], c[1], c[2], c[3]
			}
		}
	}
	return img
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
