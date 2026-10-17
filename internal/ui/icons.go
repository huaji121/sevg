package ui

import (
	"fmt"
	"image"
	"image/png"

	"sevg/assets"
	"sevg/internal/core"
)

const itemIconSize = 40

// Item icons are editable PNG assets. The order matches core.ItemID values
// starting at 1.
var itemIconFiles = []string{
	"grass_block", "dirt", "stone", "cobblestone", "sand", "oak_log", "leaves", "oak_planks",
	"coal_ore", "iron_ore", "snow", "cactus", "glass", "gravel", "flower", "tall_grass",
	"bricks", "stone_bricks", "ice", "obsidian", "stick", "apple", "coal", "iron_ingot",
	"wooden_sword", "stone_pickaxe", "iron_sword", "iron_pickaxe", "wool",
}

func buildItemIcons() map[core.ItemID]*image.RGBA {
	out := make(map[core.ItemID]*image.RGBA)
	if len(itemIconFiles) != int(core.ItemCount)-1 {
		panic(fmt.Sprintf("item icon manifest has %d entries, expected %d", len(itemIconFiles), int(core.ItemCount)-1))
	}
	for i, name := range itemIconFiles {
		item := core.ItemID(i + 1)
		path := "textures/items/" + name + ".png"
		f, err := assets.Textures.Open(path)
		if err != nil {
			panic(fmt.Sprintf("open icon %s: %v", path, err))
		}
		src, err := png.Decode(f)
		f.Close()
		if err != nil {
			panic(fmt.Sprintf("decode icon %s: %v", path, err))
		}
		b := src.Bounds()
		img := image.NewRGBA(image.Rect(0, 0, itemIconSize, itemIconSize))
		for y := 0; y < itemIconSize; y++ {
			for x := 0; x < itemIconSize; x++ {
				if x < b.Dx() && y < b.Dy() {
					img.Set(x, y, src.At(b.Min.X+x, b.Min.Y+y))
				}
			}
		}
		out[item] = img
	}
	return out
}

func buildIconSides(img *image.RGBA, g Glyph) []IconSide {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	const threshold = 40
	opaque := func(x, y int) bool {
		if x < 0 || y < 0 || x >= w || y >= h {
			return false
		}
		return img.Pix[y*img.Stride+x*4+3] > threshold
	}
	var sides []IconSide
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if !opaque(x, y) {
				continue
			}
			u := g.U0 + (float32(x)+0.5)/float32(w)*(g.U1-g.U0)
			v := g.V0 + (float32(y)+0.5)/float32(h)*(g.V1-g.V0)
			xL := float32(x)/float32(w) - 0.5
			xR := float32(x+1)/float32(w) - 0.5
			yT := 0.5 - float32(y)/float32(h)
			yB := 0.5 - float32(y+1)/float32(h)
			if !opaque(x-1, y) {
				sides = append(sides, IconSide{AX: xL, AY: yB, BX: xL, BY: yT, U: u, V: v})
			}
			if !opaque(x+1, y) {
				sides = append(sides, IconSide{AX: xR, AY: yB, BX: xR, BY: yT, U: u, V: v})
			}
			if !opaque(x, y-1) {
				sides = append(sides, IconSide{AX: xL, AY: yT, BX: xR, BY: yT, U: u, V: v})
			}
			if !opaque(x, y+1) {
				sides = append(sides, IconSide{AX: xR, AY: yB, BX: xL, BY: yB, U: u, V: v})
			}
		}
	}
	return sides
}
