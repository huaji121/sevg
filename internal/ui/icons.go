package ui

import (
	"fmt"
	"image"
	"image/png"

	"sevg/assets"
	"sevg/internal/core"
)

const itemIconSize = 40

// Flat item sprites are real texture assets. Block items with a 3D preview
// are generated from the block atlas instead of storing a second copy.
var flatItemIcons = map[core.ItemID]string{
	core.ItemStick:        "stick",
	core.ItemApple:        "apple",
	core.ItemCoal:         "coal",
	core.ItemIronIngot:    "iron_ingot",
	core.ItemWoodSword:    "wooden_sword",
	core.ItemStonePickaxe: "stone_pickaxe",
	core.ItemIronSword:    "iron_sword",
	core.ItemIronPickaxe:  "iron_pickaxe",
}

func buildItemIcons(atlas []byte) map[core.ItemID]*image.RGBA {
	out := make(map[core.ItemID]*image.RGBA)
	if len(atlas) != core.AtlasW*core.AtlasH*4 {
		return out
	}
	for item := core.ItemID(1); item < core.ItemCount; item++ {
		def := item.Def()
		if def.Kind != core.KindBlock {
			name, ok := flatItemIcons[item]
			if !ok {
				panic(fmt.Sprintf("missing flat icon asset for item %v", item))
			}
			out[item] = loadItemIcon(name)
			continue
		}
		if def.Block.Cross() {
			out[item] = plantIcon(atlas, def.Block)
		} else {
			out[item] = blockIcon(atlas, def.Block)
		}
	}
	return out
}

func loadItemIcon(name string) *image.RGBA {
	path := "textures/items/" + name + ".png"
	f, err := assets.Textures.Open(path)
	if err != nil {
		panic(fmt.Sprintf("open item icon %s: %v", path, err))
	}
	src, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(fmt.Sprintf("decode item icon %s: %v", path, err))
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
	return img
}

type iconPoint struct{ X, Y int }

func newIcon() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, itemIconSize, itemIconSize))
}

func blockIcon(atlas []byte, block core.BlockID) *image.RGBA {
	img := newIcon()
	def := block.Def()
	top := def.Tiles[2]
	side := def.Tiles[0]
	topQuad := [4]iconPoint{{20, 4}, {34, 11}, {20, 18}, {6, 11}}
	leftQuad := [4]iconPoint{{6, 11}, {20, 18}, {20, 34}, {6, 27}}
	rightQuad := [4]iconPoint{{20, 18}, {34, 11}, {34, 27}, {20, 34}}
	uv := [4][2]float32{{0, 0}, {1, 0}, {1, 1}, {0, 1}}
	rasterFace(img, topQuad, uv, atlas, top, 1.0)
	rasterFace(img, leftQuad, uv, atlas, side, 0.72)
	rasterFace(img, rightQuad, uv, atlas, side, 0.86)
	return img
}

func plantIcon(atlas []byte, block core.BlockID) *image.RGBA {
	img := newIcon()
	tile := block.Def().Tiles[0]
	const px, py, pw, ph = 5, 3, 30, 34
	for y := 0; y < ph; y++ {
		for x := 0; x < pw; x++ {
			u := (float32(x) + 0.5) / pw
			v := (float32(y) + 0.5) / ph
			r, g, b, a := sampleTile(atlas, tile, u, v)
			if a > 0 {
				setPixel(img, px+x, py+y, r, g, b, a)
			}
		}
	}
	return img
}

func rasterFace(dst *image.RGBA, q [4]iconPoint, uv [4][2]float32, atlas []byte, tile uint8, shade float32) {
	minX, minY := q[0].X, q[0].Y
	maxX, maxY := q[0].X, q[0].Y
	for _, p := range q {
		if p.X < minX {
			minX = p.X
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	tris := [2][3]int{{0, 1, 2}, {0, 2, 3}}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			for _, tri := range tris {
				a := q[tri[0]]
				b := q[tri[1]]
				c := q[tri[2]]
				w0, w1, w2, ok := bary(px, py, a, b, c)
				if !ok {
					continue
				}
				u := uv[tri[0]][0]*float32(w0) + uv[tri[1]][0]*float32(w1) + uv[tri[2]][0]*float32(w2)
				v := uv[tri[0]][1]*float32(w0) + uv[tri[1]][1]*float32(w1) + uv[tri[2]][1]*float32(w2)
				r, g, bl, al := sampleTile(atlas, tile, u, v)
				if al == 0 {
					break
				}
				r = shadeByte(r, shade)
				g = shadeByte(g, shade)
				bl = shadeByte(bl, shade)
				blendPixel(dst, x, y, r, g, bl, al)
				break
			}
		}
	}
}

func bary(px, py float64, a, b, c iconPoint) (float64, float64, float64, bool) {
	x0, y0 := float64(a.X), float64(a.Y)
	x1, y1 := float64(b.X), float64(b.Y)
	x2, y2 := float64(c.X), float64(c.Y)
	den := (y1-y2)*(x0-x2) + (x2-x1)*(y0-y2)
	if den == 0 {
		return 0, 0, 0, false
	}
	w0 := ((y1-y2)*(px-x2) + (x2-x1)*(py-y2)) / den
	w1 := ((y2-y0)*(px-x2) + (x0-x2)*(py-y2)) / den
	w2 := 1 - w0 - w1
	if w0 < -0.001 || w1 < -0.001 || w2 < -0.001 {
		return 0, 0, 0, false
	}
	return w0, w1, w2, true
}

func sampleTile(atlas []byte, tile uint8, u, v float32) (uint8, uint8, uint8, uint8) {
	x := int(u * core.AtlasTilePx)
	y := int(v * core.AtlasTilePx)
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x >= core.AtlasTilePx {
		x = core.AtlasTilePx - 1
	}
	if y >= core.AtlasTilePx {
		y = core.AtlasTilePx - 1
	}
	cx := int(tile) % core.AtlasCols
	cy := int(tile) / core.AtlasCols
	i := ((cy*core.AtlasCell+core.AtlasPad+y)*core.AtlasW +
		cx*core.AtlasCell + core.AtlasPad + x) * 4
	return atlas[i], atlas[i+1], atlas[i+2], atlas[i+3]
}

func setPixel(img *image.RGBA, x, y int, r, g, b, a uint8) {
	if x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() {
		return
	}
	i := y*img.Stride + x*4
	img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, a
}

func blendPixel(img *image.RGBA, x, y int, r, g, b, a uint8) {
	if a == 0 || x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() {
		return
	}
	i := y*img.Stride + x*4
	sa := float32(a) / 255
	da := float32(img.Pix[i+3]) / 255
	outA := sa + da*(1-sa)
	if outA <= 0 {
		return
	}
	img.Pix[i] = uint8((float32(r)*sa + float32(img.Pix[i])*da*(1-sa)) / outA)
	img.Pix[i+1] = uint8((float32(g)*sa + float32(img.Pix[i+1])*da*(1-sa)) / outA)
	img.Pix[i+2] = uint8((float32(b)*sa + float32(img.Pix[i+2])*da*(1-sa)) / outA)
	img.Pix[i+3] = uint8(outA * 255)
}

func shadeByte(v uint8, f float32) uint8 {
	n := float32(v) * f
	if n < 0 {
		n = 0
	}
	if n > 255 {
		n = 255
	}
	return uint8(n)
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
