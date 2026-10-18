package render

import (
	"bytes"
	"fmt"
	"image/png"
	"os"
	"sync"

	"sevg/assets"
	"sevg/internal/core"
)

// Textures are kept as individual PNG files in categorized assets/textures folders.
// The order here matches the core.Tile* constants.
var customSkinPath string

// SetSkinPath selects an external 64x64 Minecraft-layout skin texture used for the player model.
// It must be called before the first BuildAtlas call.
func SetSkinPath(path string) { customSkinPath = path }

var blockTextureFiles = [core.TileTileCount]string{
	"blocks/grass_top", "blocks/grass_side", "blocks/dirt", "blocks/stone", "blocks/cobble", "blocks/sand", "blocks/water",
	"blocks/log_side", "blocks/log_top", "blocks/leaves", "blocks/planks", "blocks/coal_ore", "blocks/iron_ore",
	"blocks/snow", "blocks/snow_side", "blocks/cactus_side", "blocks/cactus_top", "blocks/glass", "blocks/bedrock",
	"blocks/gravel", "blocks/flower", "blocks/tall_grass", "blocks/brick", "blocks/stone_brick", "blocks/ice",
	"blocks/obsidian", "blocks/wool", "blocks/torch", "blocks/crafting_table",
	"effects/crack_0", "effects/crack_1", "effects/crack_2", "effects/crack_3", "effects/crack_4",
	"effects/crack_5", "effects/crack_6", "effects/crack_7", "effects/crack_8", "effects/crack_9",
	"palette/white", "entities/pig_skin", "palette/dark", "palette/red", "palette/gold", "palette/gray",
	"palette/black", "palette/brown", "palette/pink", "palette/green", "entities/player_skin_map",
	"sky/sun", "sky/moon", "blocks/chest",
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
		path := "textures/" + name + ".png"
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
		b := src.Bounds()
		for y := 0; y < core.AtlasTilePx; y++ {
			for x := 0; x < core.AtlasTilePx; x++ {
				sx := b.Min.X + x*b.Dx()/core.AtlasTilePx
				sy := b.Min.Y + y*b.Dy()/core.AtlasTilePx
				r, g, b, a := src.At(sx, sy).RGBA()
				i := ((tileY+y)*core.AtlasW + tileX + x) * 4
				img[i], img[i+1], img[i+2], img[i+3] =
					byte(r>>8), byte(g>>8), byte(b>>8), byte(a>>8)
			}
		}
	}

	// Optional external skin overrides the built-in bright test skin.
	if customSkinPath != "" {
		data, err := os.ReadFile(customSkinPath)
		if err != nil {
			panic(fmt.Sprintf("open skin %s: %v", customSkinPath, err))
		}
		src, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			panic(fmt.Sprintf("decode skin %s: %v", customSkinPath, err))
		}
		b := src.Bounds()
		tile := int(core.TilePlayerSkinMap)
		cx, cy := tile%core.AtlasCols, tile/core.AtlasCols
		tileX := cx*core.AtlasCell + core.AtlasPad
		tileY := cy*core.AtlasCell + core.AtlasPad
		for y := 0; y < core.AtlasTilePx; y++ {
			for x := 0; x < core.AtlasTilePx; x++ {
				sx := b.Min.X + x*b.Dx()/core.AtlasTilePx
				sy := b.Min.Y + y*b.Dy()/core.AtlasTilePx
				r, g, bl, a := src.At(sx, sy).RGBA()
				i := ((tileY+y)*core.AtlasW + tileX + x) * 4
				img[i], img[i+1], img[i+2], img[i+3] = byte(r>>8), byte(g>>8), byte(bl>>8), byte(a>>8)
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
