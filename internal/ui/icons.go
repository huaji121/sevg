package ui

import (
	"image"
	"math"
	"sort"

	"sevg/internal/core"
)

const itemIconSize = 40

func buildItemIcons(atlas []byte) map[core.ItemID]*image.RGBA {
	out := make(map[core.ItemID]*image.RGBA)
	if len(atlas) != core.AtlasW*core.AtlasH*4 {
		return out
	}
	for item := core.ItemID(1); item < core.ItemCount; item++ {
		def := item.Def()
		var img *image.RGBA
		switch {
		case def.Kind == core.KindBlock && def.Block.Cross():
			img = plantIcon(atlas, def.Block)
		case def.Kind == core.KindBlock:
			img = blockIcon(atlas, def.Block)
		case item == core.ItemWoodSword || item == core.ItemIronSword:
			img = swordIcon(def.Color)
		case item == core.ItemStonePickaxe || item == core.ItemIronPickaxe:
			img = pickaxeIcon(def.Color)
		case item == core.ItemApple:
			img = appleIcon()
		case item == core.ItemCoal:
			img = coalIcon()
		case item == core.ItemStick:
			img = stickIcon()
		case item == core.ItemIronIngot:
			img = ingotIcon()
		default:
			img = materialIcon(def.Color)
		}
		if img != nil {
			// Tools, food and materials are rendered as hard-edged pixel
			// sprites in the world. Keeping anti-aliased alpha here would
			// produce a translucent fringe around the held/dropped item.
			if def.Kind != core.KindBlock {
				hardenIconAlpha(img)
			}
			out[item] = img
		}
	}
	return out
}

type iconPoint struct{ X, Y int }

func newIcon() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, itemIconSize, itemIconSize))
}

func blockIcon(atlas []byte, block core.BlockID) *image.RGBA {
	img := newIcon()
	top := [4]iconPoint{{20, 3}, {37, 11}, {20, 19}, {3, 11}}
	left := [4]iconPoint{{3, 11}, {20, 19}, {20, 37}, {3, 29}}
	right := [4]iconPoint{{20, 19}, {37, 11}, {37, 29}, {20, 37}}
	minAlpha := uint8(0)
	if block == core.Glass || block == core.Ice {
		minAlpha = 155
	}
	rasterFace(img, top, atlas, block.Tile(2), 1.0, minAlpha)
	rasterFace(img, left, atlas, block.Tile(4), 0.78, minAlpha)
	rasterFace(img, right, atlas, block.Tile(0), 0.62, minAlpha)
	return img
}

func plantIcon(atlas []byte, block core.BlockID) *image.RGBA {
	img := newIcon()
	tile := block.Tile(2)
	for y := 2; y < itemIconSize-2; y++ {
		for x := 2; x < itemIconSize-2; x++ {
			u := float32(x-2) / float32(itemIconSize-6)
			v := float32(y-2) / float32(itemIconSize-6)
			r, g, b, a := sampleTile(atlas, tile, u, v)
			if a > 0 {
				setPixel(img, x, y, r, g, b, a)
			}
		}
	}
	return img
}

func rasterFace(dst *image.RGBA, q [4]iconPoint, atlas []byte, tile uint8, shade float32, minAlpha uint8) {
	rasterTri(dst, q[0], q[1], q[2], [2]float32{0, 0}, [2]float32{1, 0}, [2]float32{1, 1}, atlas, tile, shade, minAlpha)
	rasterTri(dst, q[0], q[2], q[3], [2]float32{0, 0}, [2]float32{1, 1}, [2]float32{0, 1}, atlas, tile, shade, minAlpha)
}

func rasterTri(dst *image.RGBA, p0, p1, p2 iconPoint, uv0, uv1, uv2 [2]float32, atlas []byte, tile uint8, shade float32, minAlpha uint8) {
	area := edge(p0, p1, p2)
	if area == 0 {
		return
	}
	minX := minInt3(p0.X, p1.X, p2.X)
	maxX := maxInt3(p0.X, p1.X, p2.X)
	minY := minInt3(p0.Y, p1.Y, p2.Y)
	maxY := maxInt3(p0.Y, p1.Y, p2.Y)
	for y := max(minY, 0); y <= min(maxY, itemIconSize-1); y++ {
		for x := max(minX, 0); x <= min(maxX, itemIconSize-1); x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			w0 := edgeF(float64(p1.X), float64(p1.Y), float64(p2.X), float64(p2.Y), px, py)
			w1 := edgeF(float64(p2.X), float64(p2.Y), float64(p0.X), float64(p0.Y), px, py)
			w2 := edgeF(float64(p0.X), float64(p0.Y), float64(p1.X), float64(p1.Y), px, py)
			if area > 0 {
				if w0 < 0 || w1 < 0 || w2 < 0 {
					continue
				}
			} else if w0 > 0 || w1 > 0 || w2 > 0 {
				continue
			}
			a := float64(area)
			u := float32((w0*float64(uv0[0]) + w1*float64(uv1[0]) + w2*float64(uv2[0])) / a)
			v := float32((w0*float64(uv0[1]) + w1*float64(uv1[1]) + w2*float64(uv2[1])) / a)
			r, g, b, alpha := sampleTile(atlas, tile, u, v)
			if alpha == 0 {
				continue
			}
			if alpha < minAlpha {
				alpha = minAlpha
			}
			setPixel(dst, x, y, shadeByte(r, shade), shadeByte(g, shade), shadeByte(b, shade), alpha)
		}
	}
}

func sampleTile(atlas []byte, tile uint8, u, v float32) (uint8, uint8, uint8, uint8) {
	tx := int(iconClamp(u, 0, 0.999) * core.AtlasTilePx)
	ty := int(iconClamp(v, 0, 0.999) * core.AtlasTilePx)
	cx := int(tile) % core.AtlasCols
	cy := int(tile) / core.AtlasCols
	px := cx*core.AtlasCell + core.AtlasPad + tx
	py := cy*core.AtlasCell + core.AtlasPad + ty
	i := (py*core.AtlasW + px) * 4
	return atlas[i], atlas[i+1], atlas[i+2], atlas[i+3]
}

func swordIcon(col [3]float32) *image.RGBA {
	img := newIcon()
	dark := [4]uint8{48, 34, 20, 255}
	handle := [4]uint8{112, 76, 38, 255}
	blade := [4]uint8{byte(col[0] * 255), byte(col[1] * 255), byte(col[2] * 255), 255}
	thickLine(img, 8, 35, 17, 26, 3.2, dark)
	thickLine(img, 8, 35, 17, 26, 2.2, handle)
	thickLine(img, 10, 22, 24, 31, 2.6, dark)
	thickLine(img, 10, 22, 24, 31, 1.6, handle)
	thickLine(img, 17, 26, 35, 8, 3.3, dark)
	thickLine(img, 17, 26, 35, 8, 2.2, blade)
	thickLine(img, 31, 12, 35, 8, 1.6, [4]uint8{235, 235, 230, 255})
	return img
}

func pickaxeIcon(col [3]float32) *image.RGBA {
	img := newIcon()
	dark := [4]uint8{48, 34, 20, 255}
	handle := [4]uint8{112, 76, 38, 255}
	head := [4]uint8{byte(col[0] * 255), byte(col[1] * 255), byte(col[2] * 255), 255}
	thickLine(img, 9, 35, 26, 9, 3.3, dark)
	thickLine(img, 9, 35, 26, 9, 2.3, handle)
	thickLine(img, 10, 16, 24, 8, 4.2, dark)
	thickLine(img, 24, 8, 37, 16, 4.2, dark)
	thickLine(img, 10, 16, 24, 8, 3.0, head)
	thickLine(img, 24, 8, 37, 16, 3.0, head)
	thickLine(img, 10, 16, 11, 21, 2.4, head)
	thickLine(img, 37, 16, 35, 21, 2.4, head)
	return img
}

func appleIcon() *image.RGBA {
	img := newIcon()
	fillCircle(img, 19, 25, 10.5, [4]uint8{196, 30, 34, 255})
	fillCircle(img, 16, 21, 3.2, [4]uint8{238, 96, 90, 220})
	thickLine(img, 19, 16, 23, 8, 1.8, [4]uint8{96, 64, 30, 255})
	fillEllipse(img, 28, 10, 6, 3, [4]uint8{74, 146, 56, 255})
	return img
}

func coalIcon() *image.RGBA {
	img := newIcon()
	fillCircle(img, 18, 24, 9.5, [4]uint8{22, 22, 26, 255})
	fillCircle(img, 26, 20, 7, [4]uint8{30, 30, 34, 255})
	fillCircle(img, 25, 30, 6, [4]uint8{18, 18, 22, 255})
	fillCircle(img, 15, 19, 2.4, [4]uint8{92, 92, 100, 180})
	return img
}

func stickIcon() *image.RGBA {
	img := newIcon()
	thickLine(img, 9, 33, 31, 8, 3.2, [4]uint8{62, 42, 20, 255})
	thickLine(img, 9, 33, 31, 8, 2.2, [4]uint8{150, 105, 52, 255})
	thickLine(img, 13, 26, 18, 22, 0.9, [4]uint8{190, 145, 82, 255})
	return img
}

func ingotIcon() *image.RGBA {
	img := newIcon()
	poly := []iconPoint{{7, 25}, {25, 18}, {35, 24}, {17, 32}}
	fillPoly(img, poly, [4]uint8{138, 140, 148, 255})
	thickLine(img, 10, 25, 25, 20, 1.3, [4]uint8{210, 212, 218, 255})
	thickLine(img, 17, 31, 31, 24, 1.3, [4]uint8{86, 88, 96, 255})
	return img
}

func materialIcon(col [3]float32) *image.RGBA {
	img := newIcon()
	c := [4]uint8{byte(col[0] * 255), byte(col[1] * 255), byte(col[2] * 255), 255}
	fillCircle(img, 20, 21, 10, [4]uint8{0, 0, 0, 120})
	fillCircle(img, 19, 20, 9, c)
	fillCircle(img, 16, 17, 3, [4]uint8{255, 255, 255, 70})
	return img
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

func hardenIconAlpha(img *image.RGBA) {
	for y := 0; y < img.Rect.Dy(); y++ {
		for x := 0; x < img.Rect.Dx(); x++ {
			i := y*img.Stride + x*4
			if img.Pix[i+3] < 128 {
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 0, 0, 0, 0
			} else {
				img.Pix[i+3] = 255
			}
		}
	}
}

func setPixel(img *image.RGBA, x, y int, r, g, b, a uint8) {
	if x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() {
		return
	}
	i := y*img.Stride + x*4
	img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, b, a
}

func blendPixel(img *image.RGBA, x, y int, r, g, b, a uint8) {
	if x < 0 || y < 0 || x >= img.Rect.Dx() || y >= img.Rect.Dy() || a == 0 {
		return
	}
	i := y*img.Stride + x*4
	da := int(img.Pix[i+3])
	sa := int(a)
	outA := sa + da*(255-sa)/255
	if outA == 0 {
		return
	}
	for k, sv := range [3]uint8{r, g, b} {
		dv := int(img.Pix[i+k])
		img.Pix[i+k] = uint8((int(sv)*sa*255 + dv*da*(255-sa)) / outA / 255)
	}
	img.Pix[i+3] = uint8(outA)
}

func fillCircle(img *image.RGBA, cx, cy, r float64, col [4]uint8) {
	for y := int(cy - r - 1); y <= int(cy+r+1); y++ {
		for x := int(cx - r - 1); x <= int(cx+r+1); x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			if d > r {
				continue
			}
			a := col[3]
			if d > r-1 {
				a = uint8(float64(a) * (r - d))
			}
			blendPixel(img, x, y, col[0], col[1], col[2], a)
		}
	}
}

func fillEllipse(img *image.RGBA, cx, cy, rx, ry float64, col [4]uint8) {
	for y := int(cy - ry - 1); y <= int(cy+ry+1); y++ {
		for x := int(cx - rx - 1); x <= int(cx+rx+1); x++ {
			dx := (float64(x) + 0.5 - cx) / rx
			dy := (float64(y) + 0.5 - cy) / ry
			if dx*dx+dy*dy <= 1 {
				blendPixel(img, x, y, col[0], col[1], col[2], col[3])
			}
		}
	}
}

func thickLine(img *image.RGBA, x0, y0, x1, y1, r float64, col [4]uint8) {
	minX := int(math.Min(x0, x1) - r - 1)
	maxX := int(math.Max(x0, x1) + r + 1)
	minY := int(math.Min(y0, y1) - r - 1)
	maxY := int(math.Max(y0, y1) + r + 1)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			d := distToSegment(float64(x)+0.5, float64(y)+0.5, x0, y0, x1, y1)
			if d > r {
				continue
			}
			a := col[3]
			if d > r-1 {
				a = uint8(float64(a) * (r - d))
			}
			blendPixel(img, x, y, col[0], col[1], col[2], a)
		}
	}
}

func fillPoly(img *image.RGBA, poly []iconPoint, col [4]uint8) {
	if len(poly) < 3 {
		return
	}
	minY, maxY := poly[0].Y, poly[0].Y
	for _, p := range poly[1:] {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}
	for y := minY; y <= maxY; y++ {
		var xs []float64
		py := float64(y) + 0.5
		for i := range poly {
			a, b := poly[i], poly[(i+1)%len(poly)]
			if (float64(a.Y) <= py && float64(b.Y) > py) || (float64(b.Y) <= py && float64(a.Y) > py) {
				t := (py - float64(a.Y)) / float64(b.Y-a.Y)
				xs = append(xs, float64(a.X)+t*float64(b.X-a.X))
			}
		}
		sort.Float64s(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			for x := int(math.Ceil(xs[i] - 0.5)); x <= int(math.Floor(xs[i+1]-0.5)); x++ {
				blendPixel(img, x, y, col[0], col[1], col[2], col[3])
			}
		}
	}
}

func distToSegment(px, py, x0, y0, x1, y1 float64) float64 {
	dx, dy := x1-x0, y1-y0
	if dx == 0 && dy == 0 {
		return math.Hypot(px-x0, py-y0)
	}
	t := ((px-x0)*dx + (py-y0)*dy) / (dx*dx + dy*dy)
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(px-(x0+t*dx), py-(y0+t*dy))
}

func edge(a, b, c iconPoint) float64 {
	return edgeF(float64(a.X), float64(a.Y), float64(b.X), float64(b.Y), float64(c.X), float64(c.Y))
}

func edgeF(ax, ay, bx, by, px, py float64) float64 {
	return (px-ax)*(by-ay) - (py-ay)*(bx-ax)
}

func minInt3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

func maxInt3(a, b, c int) int {
	if b > a {
		a = b
	}
	if c > a {
		a = c
	}
	return a
}

func shadeByte(v uint8, f float32) uint8 {
	return uint8(iconClamp(float32(v)*f, 0, 255))
}
func iconClamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
