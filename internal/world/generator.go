package world

import (
	"math"

	"sevg/internal/core"
)

const (
	ChunkSize   = 16
	ChunkHeight = 256
	SeaLevel    = 70
	ChunkArea   = ChunkSize * ChunkSize * ChunkHeight
)

type Biome uint8

const (
	BiomePlains Biome = iota
	BiomeForest
	BiomeDesert
	BiomeSnow
	BiomeMountains
	BiomeHighMountains
	BiomeOcean
)

func (b Biome) Name() string {
	switch b {
	case BiomeForest:
		return "Forest"
	case BiomeDesert:
		return "Desert"
	case BiomeSnow:
		return "Snow"
	case BiomeMountains:
		return "Mountains"
	case BiomeHighMountains:
		return "High Mountains"
	case BiomeOcean:
		return "Ocean"
	default:
		return "Plains"
	}
}

type ChunkPos struct {
	X, Z int
}

func (p ChunkPos) Key() [2]int { return [2]int{p.X, p.Z} }

func ChunkOf(x, z int) ChunkPos {
	return ChunkPos{floorDiv(x, ChunkSize), floorDiv(z, ChunkSize)}
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func Mod(a, b int) int {
	m := a % b
	if m < 0 {
		m += b
	}
	return m
}

// Generator deterministically turns a world seed into terrain and features.
type Generator struct {
	Seed   int64
	height *Noise
	detail *Noise
	ridge  *Noise
	alpine *Noise
	temp   *Noise
	moist  *Noise
	cave   *Noise
	ore    *Noise
	tree   *Noise
}

func NewGenerator(seed int64) *Generator {
	return &Generator{
		Seed:   seed,
		height: NewNoise(seed + 11),
		detail: NewNoise(seed + 29),
		ridge:  NewNoise(seed + 47),
		alpine: NewNoise(seed + 251),
		temp:   NewNoise(seed + 71),
		moist:  NewNoise(seed + 97),
		cave:   NewNoise(seed + 131),
		ore:    NewNoise(seed + 173),
		tree:   NewNoise(seed + 211),
	}
}

func smoothstep(a, b, x float64) float64 {
	t := (x - a) / (b - a)
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return t * t * (3 - 2*t)
}

// HeightAt returns the terrain surface height and biome for a world column.
func (g *Generator) HeightAt(wx, wz int) (int, Biome) {
	x, z := float64(wx), float64(wz)
	continent := g.height.Fbm2(x*0.0032, z*0.0032, 4, 2.0, 0.5)
	detail := g.detail.Fbm2(x*0.021, z*0.021, 4, 2.0, 0.5)
	ridge := g.ridge.Ridged2(x*0.0055+100, z*0.0055-70, 4)
	mountainMask := smoothstep(0.02, 0.55, continent)
	// The base height tracks SeaLevel so raising sea level lifts the whole
	// terrain (deeper underground) instead of flooding the world.
	h := float64(SeaLevel) + continent*19.0 + detail*5.5 + mountainMask*ridge*32.0

	// Alpine ranges rise from the highest continental areas. The ridged noise
	// creates sharp connecting peaks instead of rounded hills.
	alpineNoise := g.alpine.Fbm2(x*0.0021+500, z*0.0021-300, 4, 2.0, 0.5)
	alpineMask := smoothstep(0.08, 0.50, continent) * (0.75 + 0.25*smoothstep(-0.2, 0.3, alpineNoise))
	alpineRidge := g.alpine.Ridged2(x*0.0031+410, z*0.0031-310, 5)
	h += alpineMask * (alpineRidge*175.0 + detail*8.0)

	temp := g.temp.Fbm2(x*0.0028-41, z*0.0028+19, 3, 2.0, 0.5)
	moist := g.moist.Fbm2(x*0.0036+77, z*0.0036-31, 3, 2.0, 0.5)

	biome := BiomePlains
	switch {
	case h >= float64(SeaLevel+97):
		biome = BiomeHighMountains
	case h > float64(SeaLevel+30):
		biome = BiomeMountains
	case temp > 0.30 && moist < -0.05:
		biome = BiomeDesert
	case temp < -0.24:
		biome = BiomeSnow
	case moist > 0.12:
		biome = BiomeForest
	}
	if h < float64(SeaLevel-1) {
		biome = BiomeOcean
	}
	return int(math.Round(h)), biome
}

func (g *Generator) surfaceBlock(biome Biome, h int) core.BlockID {
	switch biome {
	case BiomeDesert:
		return core.Sand
	case BiomeSnow:
		if h > SeaLevel+62 {
			return core.Snow
		}
		return core.SnowyGrass
	case BiomeHighMountains:
		if h > SeaLevel+137 {
			return core.Snow
		}
		if h > SeaLevel+102 {
			return core.Stone
		}
		return core.Grass
	case BiomeMountains:
		if h > SeaLevel+77 {
			return core.Snow
		}
		if h > SeaLevel+52 {
			return core.Stone
		}
		return core.Grass
	case BiomeOcean:
		if h < SeaLevel-7 {
			return core.Gravel
		}
		return core.Sand
	default:
		return core.Grass
	}
}

func (g *Generator) GenerateChunk(cp ChunkPos) []core.BlockID {
	blocks := make([]core.BlockID, ChunkArea)
	baseX, baseZ := cp.X*ChunkSize, cp.Z*ChunkSize

	for lx := 0; lx < ChunkSize; lx++ {
		for lz := 0; lz < ChunkSize; lz++ {
			wx, wz := baseX+lx, baseZ+lz
			h, biome := g.HeightAt(wx, wz)
			if h >= ChunkHeight {
				h = ChunkHeight - 1
			}
			surface := g.surfaceBlock(biome, h)
			for y := 0; y <= h; y++ {
				var b core.BlockID
				switch {
				case y == 0:
					b = core.Bedrock
				case y < h-4:
					b = core.Stone
				case y < h:
					b = core.Dirt
					if biome == BiomeDesert || biome == BiomeOcean {
						b = core.Sand
					} else if biome == BiomeMountains || biome == BiomeHighMountains {
						b = core.Stone
					}
				default:
					b = surface
				}
				blocks[index(lx, y, lz)] = b
			}
			// Water fills all depressions up to sea level.
			for y := h + 1; y <= SeaLevel; y++ {
				blocks[index(lx, y, lz)] = core.Water
			}
		}
	}

	// Carve caves and decorate with ores.
	for lx := 0; lx < ChunkSize; lx++ {
		for lz := 0; lz < ChunkSize; lz++ {
			wx, wz := baseX+lx, baseZ+lz
			h, _ := g.HeightAt(wx, wz)
			for y := 2; y < ChunkHeight-1; y++ {
				i := index(lx, y, lz)
				b := blocks[i]
				if b != core.Stone && b != core.Dirt && b != core.Sand && b != core.Grass && b != core.SnowyGrass {
					continue
				}
				c := g.cave.Fbm3(float64(wx)*0.034, float64(y)*0.052, float64(wz)*0.034, 3, 2.0, 0.5)
				if c > 0.43 {
					blocks[i] = core.Air
					continue
				}
				// Ores only replace buried stone so no ore ever shows up as the
				// terrain surface (e.g. coal in grass).
				if b == core.Stone && y < h && y < SeaLevel+24 {
					r := g.ore.Unit3(wx, y, wz)
					if y < SeaLevel && r < 0.010 {
						blocks[i] = core.IronOre
					} else if r < 0.025 {
						blocks[i] = core.CoalOre
					} else if r > 0.986 {
						blocks[i] = core.Gravel
					}
				}
			}
		}
	}

	g.decorate(blocks, cp)
	return blocks
}

func index(x, y, z int) int { return x + z*ChunkSize + y*ChunkSize*ChunkSize }

func (g *Generator) decorate(blocks []core.BlockID, cp ChunkPos) {
	baseX, baseZ := cp.X*ChunkSize, cp.Z*ChunkSize
	set := func(wx, y, wz int, b core.BlockID) {
		lx, lz := wx-baseX, wz-baseZ
		if lx < 0 || lx >= ChunkSize || lz < 0 || lz >= ChunkSize || y < 0 || y >= ChunkHeight {
			return
		}
		i := index(lx, y, lz)
		if blocks[i] == core.Air || b == core.Water {
			blocks[i] = b
		}
	}
	surface := func(wx, wz int) (int, Biome, core.BlockID) {
		h, biome := g.HeightAt(wx, wz)
		if h < 0 || h >= ChunkHeight-2 {
			return h, biome, core.Air
		}
		return h, biome, g.surfaceBlock(biome, h)
	}

	// Trees. Scan a margin so canopies cross chunk borders identically in
	// every chunk that they touch.
	for wx := baseX - 3; wx < baseX+ChunkSize+3; wx++ {
		for wz := baseZ - 3; wz < baseZ+ChunkSize+3; wz++ {
			h, biome, ground := surface(wx, wz)
			if ground != core.Grass && ground != core.SnowyGrass {
				continue
			}
			if h <= SeaLevel+0 {
				continue
			}
			var chance uint64
			switch biome {
			case BiomeForest:
				chance = 14
			case BiomePlains:
				chance = 63
			case BiomeSnow:
				chance = 47
			default:
				continue
			}
			if g.tree.Hash2(wx, wz)%chance != 0 {
				continue
			}
			// Keep trees away from very steep terrain.
			if h2, _ := g.HeightAt(wx+1, wz); abs(h2-h) > 2 {
				continue
			}
			th := 4 + int(g.tree.Hash2(wx, wz)%3)
			for y := 1; y <= th; y++ {
				set(wx, h+y, wz, core.Log)
			}
			top := h + th
			for dy := -2; dy <= 1; dy++ {
				r := 2
				if dy == 1 {
					r = 1
				}
				for dx := -r; dx <= r; dx++ {
					for dz := -r; dz <= r; dz++ {
						if dx == 0 && dz == 0 && dy <= 0 {
							continue
						}
						if abs(dx) == r && abs(dz) == r && g.tree.Hash3(wx+dx, top+dy, wz+dz)%3 == 0 {
							continue
						}
						set(wx+dx, top+dy, wz+dz, core.Leaves)
					}
				}
			}
			set(wx, top+2, wz, core.Leaves)
		}
	}

	// Small ground cover and desert cacti.
	for lx := 0; lx < ChunkSize; lx++ {
		for lz := 0; lz < ChunkSize; lz++ {
			wx, wz := baseX+lx, baseZ+lz
			h, biome := g.HeightAt(wx, wz)
			if h <= SeaLevel || h+3 >= ChunkHeight {
				continue
			}
			top := blocks[index(lx, h, lz)]
			above := blocks[index(lx, h+1, lz)]
			if above != core.Air {
				continue
			}
			r := g.tree.Hash3(wx, biomeInt(biome), wz)
			if top == core.Grass {
				if r%11 == 0 {
					set(wx, h+1, wz, core.Flower)
				} else if r%5 == 0 {
					set(wx, h+1, wz, core.TallGrass)
				}
			}
			if top == core.Sand && biome == BiomeDesert && r%97 == 0 {
				ch := 1 + int((r>>8)%3)
				for y := 1; y <= ch; y++ {
					set(wx, h+y, wz, core.Cactus)
				}
			}
		}
	}
}

func biomeInt(b Biome) int { return int(b) }

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
