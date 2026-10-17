package render

import (
	"math/rand"

	"sevg/internal/core"
)

type rgba [4]uint8

func (c rgba) shade(f float32) rgba {
	return rgba{uint8(float32(c[0]) * f), uint8(float32(c[1]) * f), uint8(float32(c[2]) * f), c[3]}
}

func (c rgba) alpha(a uint8) rgba { c[3] = a; return c }

func mix(a, b rgba, t float32) rgba {
	return rgba{
		uint8(float32(a[0]) + (float32(b[0])-float32(a[0]))*t),
		uint8(float32(a[1]) + (float32(b[1])-float32(a[1]))*t),
		uint8(float32(a[2]) + (float32(b[2])-float32(a[2]))*t),
		uint8(float32(a[3]) + (float32(b[3])-float32(a[3]))*t),
	}
}

// BuildAtlas renders all block textures without any external image assets.
func BuildAtlas() []byte {
	const size = core.AtlasW * core.AtlasH
	img := make([]byte, size*4)
	put := func(tile, x, y int, c rgba) {
		if x < 0 || x >= core.AtlasTilePx || y < 0 || y >= core.AtlasTilePx {
			return
		}
		cx := tile % core.AtlasCols
		cy := tile / core.AtlasCols
		px := cx*core.AtlasCell + core.AtlasPad + x
		py := cy*core.AtlasCell + core.AtlasPad + y
		i := (py*core.AtlasW + px) * 4
		img[i], img[i+1], img[i+2], img[i+3] = c[0], c[1], c[2], c[3]
	}
	for tile := 0; tile < core.TileTileCount; tile++ {
		drawTile(tile, put)
	}
	// Fill the gutters with clamped edge pixels.
	for tile := 0; tile < core.TileTileCount; tile++ {
		cx := tile % core.AtlasCols
		cy := tile / core.AtlasCols
		tileX := cx * core.AtlasCell
		tileY := cy * core.AtlasCell
		at := func(x, y int) rgba {
			x = clamp(x, 0, core.AtlasTilePx-1)
			y = clamp(y, 0, core.AtlasTilePx-1)
			px := tileX + core.AtlasPad + x
			py := tileY + core.AtlasPad + y
			i := (py*core.AtlasW + px) * 4
			return rgba{img[i], img[i+1], img[i+2], img[i+3]}
		}
		for y := 0; y < core.AtlasCell; y++ {
			for x := 0; x < core.AtlasCell; x++ {
				c := at(x-core.AtlasPad, y-core.AtlasPad)
				px := tileX + x
				py := tileY + y
				i := (py*core.AtlasW + px) * 4
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

func drawTile(tile int, put func(tile, x, y int, c rgba)) {
	rng := rand.New(rand.NewSource(int64(tile*7919 + 17)))
	noise := func(base rgba, amount int) rgba {
		d := rng.Intn(amount*2+1) - amount
		return rgba{
			uint8(clamp(int(base[0])+d, 0, 255)),
			uint8(clamp(int(base[1])+d, 0, 255)),
			uint8(clamp(int(base[2])+d, 0, 255)),
			base[3],
		}
	}
	fill := func(base rgba, amount int) {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				put(tile, x, y, noise(base, amount))
			}
		}
	}
	clear := func() {
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				put(tile, x, y, rgba{0, 0, 0, 0})
			}
		}
	}
	switch tile {
	case core.TileGrassTop:
		fill(rgba{85, 152, 61, 255}, 14)
		for i := 0; i < 34; i++ {
			x, y := rng.Intn(16), rng.Intn(16)
			put(tile, x, y, mix(rgba{98, 171, 70, 255}, rgba{72, 130, 49, 255}, rng.Float32()))
		}
	case core.TileGrassSide:
		fill(rgba{112, 79, 50, 255}, 10)
		for x := 0; x < 16; x++ {
			h := 4 + rng.Intn(2)
			for y := 0; y < h; y++ {
				put(tile, x, y, noise(rgba{88, 154, 62, 255}, 12))
			}
			if rng.Intn(3) == 0 {
				put(tile, x, h, rgba{70, 122, 47, 255})
			}
		}
	case core.TileDirt:
		fill(rgba{119, 84, 54, 255}, 16)
		for i := 0; i < 22; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{87, 61, 39, 255})
		}
	case core.TileStone:
		fill(rgba{126, 126, 130, 255}, 12)
		for i := 0; i < 8; i++ {
			x, y := rng.Intn(15), rng.Intn(15)
			put(tile, x, y, rgba{98, 98, 103, 255})
			if rng.Intn(2) == 0 {
				put(tile, x+1, y, rgba{107, 107, 112, 255})
			}
		}
	case core.TileCobble:
		fill(rgba{108, 108, 112, 255}, 12)
		for i := 0; i < 12; i++ {
			x0, y0 := rng.Intn(12), rng.Intn(12)
			w, h := 2+rng.Intn(4), 1+rng.Intn(3)
			c := noise(rgba{138, 138, 142, 255}, 12)
			for y := y0; y < y0+h && y < 16; y++ {
				for x := x0; x < x0+w && x < 16; x++ {
					put(tile, x, y, c)
				}
			}
		}
		for i := 0; i < 14; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{75, 75, 80, 255})
		}
	case core.TileSand:
		fill(rgba{214, 197, 143, 255}, 12)
		for i := 0; i < 18; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{193, 176, 124, 255})
		}
	case core.TileWater:
		fill(rgba{49, 104, 190, 185}, 9)
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if (x+y*3)%11 < 2 {
					put(tile, x, y, rgba{78, 139, 219, 185})
				}
			}
		}
	case core.TileLogSide:
		fill(rgba{112, 80, 45, 255}, 8)
		for x := 0; x < 16; x++ {
			if x%4 == 0 {
				for y := 0; y < 16; y++ {
					put(tile, x, y, rgba{78, 53, 30, 255})
				}
			} else if x%7 == 3 {
				for y := 0; y < 16; y++ {
					if rng.Intn(4) != 0 {
						put(tile, x, y, rgba{129, 95, 54, 255})
					}
				}
			}
		}
	case core.TileLogTop:
		fill(rgba{154, 121, 74, 255}, 8)
		for r := 1; r < 8; r += 2 {
			for a := 0; a < 48; a++ {
				x := 7 + int(float32(r)*cosInt(a))
				y := 7 + int(float32(r)*sinInt(a))
				put(tile, clamp(x, 0, 15), clamp(y, 0, 15), rgba{106, 78, 44, 255})
			}
		}
	case core.TileLeaves:
		clear()
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if rng.Intn(100) < 9 && x > 0 && x < 15 && y > 0 && y < 15 {
					continue
				}
				c := mix(rgba{46, 105, 41, 255}, rgba{72, 139, 56, 255}, rng.Float32())
				put(tile, x, y, c)
			}
		}
	case core.TilePlanks:
		fill(rgba{171, 126, 69, 255}, 6)
		for y := 3; y < 16; y += 4 {
			for x := 0; x < 16; x++ {
				put(tile, x, y, rgba{111, 79, 43, 255})
			}
			off := rng.Intn(8)
			for y2 := y - 3; y2 < y; y2++ {
				put(tile, off, y2, rgba{123, 89, 49, 255})
			}
		}
	case core.TileCoalOre:
		fill(rgba{124, 124, 128, 255}, 10)
		for i := 0; i < 7; i++ {
			x0, y0 := rng.Intn(13), rng.Intn(13)
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					if rng.Intn(3) != 0 {
						put(tile, x0+dx, y0+dy, rgba{38, 38, 42, 255})
					}
				}
			}
		}
	case core.TileIronOre:
		fill(rgba{126, 122, 118, 255}, 10)
		for i := 0; i < 7; i++ {
			x0, y0 := rng.Intn(13), rng.Intn(13)
			for dy := 0; dy < 3; dy++ {
				for dx := 0; dx < 3; dx++ {
					if rng.Intn(3) != 0 {
						put(tile, x0+dx, y0+dy, rgba{191, 154, 119, 255})
					}
				}
			}
		}
	case core.TileSnow:
		fill(rgba{235, 241, 247, 255}, 7)
		for i := 0; i < 10; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{213, 226, 240, 255})
		}
	case core.TileSnowSide:
		fill(rgba{119, 91, 65, 255}, 10)
		for x := 0; x < 16; x++ {
			h := 4 + rng.Intn(2)
			for y := 0; y < h; y++ {
				put(tile, x, y, noise(rgba{235, 241, 247, 255}, 6))
			}
		}
	case core.TileCactusSide:
		fill(rgba{48, 108, 53, 255}, 8)
		for x := 0; x < 16; x++ {
			if x%5 == 1 {
				for y := 0; y < 16; y++ {
					put(tile, x, y, rgba{33, 78, 38, 255})
				}
			}
			if x%5 == 3 {
				for y := 2; y < 16; y += 4 {
					put(tile, x, y, rgba{209, 222, 160, 255})
				}
			}
		}
	case core.TileCactusTop:
		fill(rgba{55, 120, 59, 255}, 7)
		for i := 0; i < 10; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{40, 94, 45, 255})
		}
		for y := 4; y < 12; y++ {
			for x := 4; x < 12; x++ {
				if (x+y)%3 == 0 {
					put(tile, x, y, rgba{38, 87, 42, 255})
				}
			}
		}
	case core.TileGlass:
		clear()
		for i := 0; i < 16; i++ {
			put(tile, i, 0, rgba{213, 237, 242, 180})
			put(tile, i, 15, rgba{213, 237, 242, 150})
			put(tile, 0, i, rgba{213, 237, 242, 180})
			put(tile, 15, i, rgba{213, 237, 242, 150})
		}
		for i := 1; i < 15; i++ {
			for j := 1; j < 15; j++ {
				put(tile, i, j, rgba{186, 224, 235, 34})
			}
		}
		for i := 0; i < 6; i++ {
			put(tile, 3+i, 4+i, rgba{235, 248, 250, 120})
			put(tile, 9+i, 3+i, rgba{235, 248, 250, 80})
		}
	case core.TileBedrock:
		fill(rgba{55, 55, 60, 255}, 14)
		for i := 0; i < 18; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{30, 30, 34, 255})
		}
	case core.TileGravel:
		fill(rgba{112, 104, 98, 255}, 18)
		for i := 0; i < 26; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), mix(rgba{132, 126, 119, 255}, rgba{82, 75, 70, 255}, rng.Float32()))
		}
	case core.TileFlower:
		clear()
		for y := 8; y < 16; y++ {
			put(tile, 7, y, rgba{53, 118, 45, 255})
			put(tile, 6, y, rgba{43, 99, 38, 255})
		}
		put(tile, 8, 12, rgba{70, 140, 51, 255})
		put(tile, 5, 11, rgba{70, 140, 51, 255})
		for y := 4; y < 9; y++ {
			for x := 4; x < 10; x++ {
				if (x-7)*(x-7)+(y-6)*(y-6) < 6 {
					put(tile, x, y, mix(rgba{226, 74, 108, 255}, rgba{244, 158, 178, 255}, rng.Float32()))
				}
			}
		}
		put(tile, 7, 6, rgba{250, 219, 105, 255})
	case core.TileTallGrass:
		clear()
		for i := 0; i < 9; i++ {
			x := 2 + rng.Intn(12)
			h := 5 + rng.Intn(9)
			c := mix(rgba{63, 131, 48, 255}, rgba{93, 158, 62, 255}, rng.Float32())
			for y := 15; y > 15-h; y-- {
				put(tile, x, y, c)
				if rng.Intn(3) == 0 {
					x += 1
				} else if rng.Intn(3) == 0 {
					x -= 1
				}
			}
		}
	case core.TileBrick:
		fill(rgba{172, 79, 66, 255}, 7)
		for y := 0; y < 16; y++ {
			if y%4 == 0 {
				for x := 0; x < 16; x++ {
					put(tile, x, y, rgba{199, 186, 165, 255})
				}
			}
		}
		for y := 0; y < 16; y++ {
			off := 0
			if (y/4)%2 == 1 {
				off = 4
			}
			for x := off; x < 16; x += 8 {
				put(tile, x, y, rgba{199, 186, 165, 255})
			}
		}
	case core.TileStoneBrick:
		fill(rgba{126, 126, 128, 255}, 8)
		for i := 0; i < 4; i++ {
			y := i * 4
			for x := 0; x < 16; x++ {
				put(tile, x, y, rgba{87, 87, 91, 255})
			}
		}
		for i := 0; i < 4; i++ {
			x := i * 4
			for y := 0; y < 16; y++ {
				put(tile, x, y, rgba{87, 87, 91, 255})
			}
		}
	case core.TileIce:
		fill(rgba{148, 197, 226, 205}, 8)
		for i := 0; i < 9; i++ {
			x, y := rng.Intn(14), rng.Intn(14)
			for j := 0; j < 3; j++ {
				put(tile, x+j, y+j, rgba{201, 230, 244, 210})
			}
		}
	case core.TileObsidian:
		fill(rgba{36, 25, 47, 255}, 8)
		for i := 0; i < 20; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), mix(rgba{67, 46, 88, 255}, rgba{24, 16, 34, 255}, rng.Float32()))
		}
	case core.TileWool:
		fill(rgba{231, 231, 224, 255}, 6)
		for i := 0; i < 20; i++ {
			put(tile, rng.Intn(16), rng.Intn(16), rgba{207, 207, 199, 255})
		}
	case core.TileWhite:
		fill(rgba{238, 238, 234, 255}, 2)
	case core.TilePig:
		fill(rgba{239, 133, 144, 255}, 3)
	case core.TileDark:
		fill(rgba{46, 36, 33, 255}, 2)
	case core.TileRed:
		fill(rgba{205, 48, 48, 255}, 3)
	case core.TileGold:
		fill(rgba{255, 220, 92, 255}, 2)
	case core.TileGray:
		fill(rgba{140, 140, 148, 255}, 3)
	case core.TileBlack:
		fill(rgba{18, 18, 22, 255}, 1)
	case core.TileBrown:
		fill(rgba{112, 80, 45, 255}, 3)
	case core.TilePink:
		fill(rgba{232, 74, 108, 255}, 3)
	case core.TileGreen:
		fill(rgba{63, 140, 61, 255}, 3)
	}
}

func cosInt(a int) float32 {
	switch a % 8 {
	case 0:
		return 1
	case 1:
		return 0.7
	case 2:
		return 0
	case 3:
		return -0.7
	case 4:
		return -1
	case 5:
		return -0.7
	case 6:
		return 0
	default:
		return 0.7
	}
}

func sinInt(a int) float32 {
	return cosInt((a + 6) % 8)
}
