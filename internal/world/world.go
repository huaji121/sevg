package world

import (
	"math"
	"sync"

	"sevg/internal/core"
	"sevg/internal/mathx"
)

type Chunk struct {
	Pos        ChunkPos
	Blocks     []core.BlockID
	SkyLight   []uint8
	BlockLight []uint8
	WaterLevel []uint8
	Dirty      bool
}

func NewChunk(pos ChunkPos) *Chunk {
	return &Chunk{
		Pos:        pos,
		Blocks:     make([]core.BlockID, ChunkArea),
		SkyLight:   make([]uint8, ChunkArea),
		BlockLight: make([]uint8, ChunkArea),
		WaterLevel: make([]uint8, ChunkArea),
		Dirty:      true,
	}
}

func (c *Chunk) GetLocal(x, y, z int) core.BlockID {
	if x < 0 || x >= ChunkSize || z < 0 || z >= ChunkSize || y < 0 || y >= ChunkHeight {
		return core.Air
	}
	return c.Blocks[index(x, y, z)]
}

// RebuildSkyLight recomputes direct vertical sky light for this chunk. It is
// intentionally column based: roofs and caves are dark, while the surface is
// fully lit. Leaves and water attenuate the column.
func (c *Chunk) RebuildSkyLight() {
	for x := 0; x < ChunkSize; x++ {
		for z := 0; z < ChunkSize; z++ {
			level := 15
			for y := ChunkHeight - 1; y >= 0; y-- {
				i := index(x, y, z)
				b := c.Blocks[i]
				if b.Opaque() {
					level = 0
					c.SkyLight[i] = 0
					continue
				}
				switch b {
				case core.Water:
					if level > 0 {
						level--
					}
				case core.Leaves:
					if level > 1 {
						level -= 2
					} else {
						level = 0
					}
				}
				c.SkyLight[i] = uint8(level)
			}
		}
	}
}

func (c *Chunk) SetLocal(x, y, z int, b core.BlockID) bool {
	if x < 0 || x >= ChunkSize || z < 0 || z >= ChunkSize || y < 0 || y >= ChunkHeight {
		return false
	}
	i := index(x, y, z)
	if c.Blocks[i] == b {
		return false
	}
	c.Blocks[i] = b
	c.Dirty = true
	return true
}

type EditRecord struct {
	Chunk ChunkPos
	Index int
	Block core.BlockID
}

type World struct {
	mu           sync.RWMutex
	Seed         int64
	Gen          *Generator
	chunks       map[[2]int]*Chunk
	edits        map[[2]int]map[int]core.BlockID
	fluidQueue   map[[3]int]bool
	lightSources int
}

func New(seed int64) *World {
	return &World{
		Seed:       seed,
		Gen:        NewGenerator(seed),
		chunks:     make(map[[2]int]*Chunk),
		edits:      make(map[[2]int]map[int]core.BlockID),
		fluidQueue: make(map[[3]int]bool),
	}
}

// EnsureChunk generates a chunk if it does not exist and reports whether it
// was newly created.
func (w *World) EnsureChunk(pos ChunkPos) (*Chunk, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ensureLocked(pos)
}

func (w *World) ensureLocked(pos ChunkPos) (*Chunk, bool) {
	key := pos.Key()
	if c := w.chunks[key]; c != nil {
		return c, false
	}
	c := NewChunk(pos)
	c.Blocks = w.Gen.GenerateChunk(pos)
	if ed := w.edits[key]; ed != nil {
		for i, b := range ed {
			if i >= 0 && i < len(c.Blocks) {
				c.Blocks[i] = b
			}
		}
	}
	c.Dirty = true
	w.chunks[key] = c
	c.RebuildSkyLight()
	needLight := w.lightSources > 0
	if !needLight {
		for _, b := range c.Blocks {
			if b.Def().Light > 0 {
				needLight = true
				break
			}
		}
	}
	if needLight {
		w.rebuildBlockLightLocked()
	}
	for _, n := range []ChunkPos{{pos.X + 1, pos.Z}, {pos.X - 1, pos.Z}, {pos.X, pos.Z + 1}, {pos.X, pos.Z - 1}} {
		if nc := w.chunks[n.Key()]; nc != nil {
			nc.Dirty = true
		}
	}
	return c, true
}

func (w *World) HasChunk(pos ChunkPos) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.chunks[pos.Key()] != nil
}

func (w *World) Chunk(pos ChunkPos) *Chunk {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.chunks[pos.Key()]
}

func (w *World) GetBlock(wx, wy, wz int) core.BlockID {
	if wy < 0 || wy >= ChunkHeight {
		return core.Air
	}
	cp := ChunkOf(wx, wz)
	w.mu.RLock()
	defer w.mu.RUnlock()
	c := w.chunks[cp.Key()]
	if c == nil {
		return core.Air
	}
	return c.Blocks[index(Mod(wx, ChunkSize), wy, Mod(wz, ChunkSize))]
}

func (w *World) blockUnlocked(wx, wy, wz int) core.BlockID {
	if wy < 0 || wy >= ChunkHeight {
		return core.Air
	}
	cp := ChunkOf(wx, wz)
	c := w.chunks[cp.Key()]
	if c == nil {
		return core.Air
	}
	return c.Blocks[index(Mod(wx, ChunkSize), wy, Mod(wz, ChunkSize))]
}

// SetBlock edits a block, records the change for saving, and marks this chunk
// and any bordering neighbours dirty.
func (w *World) SetBlock(wx, wy, wz int, b core.BlockID) bool {
	if wy < 0 || wy >= ChunkHeight {
		return false
	}
	cp := ChunkOf(wx, wz)
	lx, lz := Mod(wx, ChunkSize), Mod(wz, ChunkSize)
	w.mu.Lock()
	defer w.mu.Unlock()
	c, _ := w.ensureLocked(cp)
	i := index(lx, wy, lz)
	old := c.Blocks[i]
	oldLevel := c.WaterLevel[i]
	if !c.SetLocal(lx, wy, lz, b) {
		return false
	}
	c.WaterLevel[i] = 0
	w.queueFluidLocked(wx, wy, wz)
	if old == core.Water && oldLevel == 0 && b != core.Water {
		w.drainLocalWaterLocked(wx, wy, wz)
	}
	c.RebuildSkyLight()
	w.rebuildBlockLightLocked()
	key := cp.Key()
	if w.edits[key] == nil {
		w.edits[key] = make(map[int]core.BlockID)
	}
	w.edits[key][index(lx, wy, lz)] = b
	// Faces and ambient occlusion cross chunk boundaries.
	if lx == 0 {
		w.markDirtyLocked(ChunkPos{cp.X - 1, cp.Z})
	}
	if lx == ChunkSize-1 {
		w.markDirtyLocked(ChunkPos{cp.X + 1, cp.Z})
	}
	if lz == 0 {
		w.markDirtyLocked(ChunkPos{cp.X, cp.Z - 1})
	}
	if lz == ChunkSize-1 {
		w.markDirtyLocked(ChunkPos{cp.X, cp.Z + 1})
	}
	return true
}

func (w *World) markDirtyLocked(p ChunkPos) {
	if c := w.chunks[p.Key()]; c != nil {
		c.Dirty = true
	}
}

func (w *World) MarkDirty(p ChunkPos) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.markDirtyLocked(p)
}

func (w *World) ClearDirty(p ChunkPos) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if c := w.chunks[p.Key()]; c != nil {
		c.Dirty = false
	}
}

func (w *World) ChunkDirty(p ChunkPos) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	c := w.chunks[p.Key()]
	return c != nil && c.Dirty
}

func (w *World) MarkNeighborsDirty(p ChunkPos) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.markDirtyLocked(p)
	w.markDirtyLocked(ChunkPos{p.X + 1, p.Z})
	w.markDirtyLocked(ChunkPos{p.X - 1, p.Z})
	w.markDirtyLocked(ChunkPos{p.X, p.Z + 1})
	w.markDirtyLocked(ChunkPos{p.X, p.Z - 1})
}

// DirtyChunks returns the chunks that need remeshing, closest first.
func (w *World) DirtyChunks(center ChunkPos) []ChunkPos {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var out []ChunkPos
	for _, c := range w.chunks {
		if c.Dirty {
			out = append(out, c.Pos)
		}
	}
	sortByDistance(out, center)
	return out
}

func (w *World) Unload(p ChunkPos) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.chunks, p.Key())
}

func (w *World) Loaded() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.chunks)
}

func (w *World) ChunkPositions() []ChunkPos {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]ChunkPos, 0, len(w.chunks))
	for _, c := range w.chunks {
		out = append(out, c.Pos)
	}
	return out
}

func (w *World) ExportEdits() []EditRecord {
	w.mu.RLock()
	defer w.mu.RUnlock()
	var out []EditRecord
	for key, m := range w.edits {
		for i, b := range m {
			out = append(out, EditRecord{Chunk: ChunkPos{key[0], key[1]}, Index: i, Block: b})
		}
	}
	return out
}

func (w *World) ImportEdits(records []EditRecord) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, r := range records {
		key := r.Chunk.Key()
		if w.edits[key] == nil {
			w.edits[key] = make(map[int]core.BlockID)
		}
		w.edits[key][r.Index] = r.Block
	}
}

// SurfaceHeight finds the top non-air, non-water block in a column. It will
// generate the containing chunk if needed.
func (w *World) SurfaceHeight(wx, wz int) int {
	w.EnsureChunk(ChunkOf(wx, wz))
	for y := ChunkHeight - 1; y >= 0; y-- {
		b := w.GetBlock(wx, y, wz)
		if b != core.Air && b != core.Water && b != core.Flower && b != core.TallGrass {
			return y
		}
	}
	return 0
}

func (w *World) FindSpawn() mathx.Vec3 {
	type candidate struct {
		x, z int
		h    int
		d    int
	}
	best := candidate{d: 1 << 30}
	for r := 0; r < 10; r++ {
		for x := -r * 8; x <= r*8; x += 8 {
			for z := -r * 8; z <= r*8; z += 8 {
				h, biome := w.Gen.HeightAt(x, z)
				if h <= SeaLevel || h >= ChunkHeight-8 || biome == BiomeOcean {
					continue
				}
				d := x*x + z*z
				if d < best.d {
					best = candidate{x, z, h, d}
				}
			}
		}
		if best.d < 1<<30 {
			break
		}
	}
	x, z := 0, 0
	if best.d < 1<<30 {
		x, z = best.x, best.z
		h, _ := w.Gen.HeightAt(x, z)
		return mathx.Vec3{X: float32(x) + 0.5, Y: float32(h) + 1.05, Z: float32(z) + 0.5}
	}
	// Ocean fallback: stand on the sea floor with a small platform.
	w.EnsureChunk(ChunkPos{0, 0})
	for y := ChunkHeight - 1; y >= 0; y-- {
		if w.GetBlock(0, y, 0) != core.Air {
			return mathx.Vec3{X: 0.5, Y: float32(y) + 1.05, Z: 0.5}
		}
	}
	return mathx.Vec3{X: 0.5, Y: 64, Z: 0.5}
}

// SafeSpawn finds a nearby column with solid ground and enough clear
// headroom, avoiding tree trunks and canopies. It generates chunks as needed.
func (w *World) SafeSpawn(near mathx.Vec3, radius int) mathx.Vec3 {
	cx, cz := int(math.Floor(float64(near.X))), int(math.Floor(float64(near.Z)))
	bestX, bestZ, bestH, bestD := cx, cz, -1, 1<<30
	for r := 0; r <= radius; r++ {
		for dz := -r; dz <= r; dz++ {
			for dx := -r; dx <= r; dx++ {
				if maxAbsInt(dx) != r && maxAbsInt(dz) != r {
					continue
				}
				x, z := cx+dx, cz+dz
				h, biome := w.Gen.HeightAt(x, z)
				if h <= SeaLevel || h >= ChunkHeight-8 || biome == BiomeOcean {
					continue
				}
				w.EnsureChunk(ChunkOf(x, z))
				if !w.GetBlock(x, h, z).Solid() || w.GetBlock(x, h, z) == core.Leaves {
					continue
				}
				clear := true
				for y := h + 1; y <= h+4; y++ {
					b := w.GetBlock(x, y, z)
					if b != core.Air && b != core.Flower && b != core.TallGrass {
						clear = false
						break
					}
				}
				if !clear {
					continue
				}
				d := dx*dx + dz*dz
				if d < bestD {
					bestD = d
					bestX, bestZ, bestH = x, z, h
				}
			}
		}
		if bestH >= 0 {
			break
		}
	}
	if bestH < 0 {
		return near
	}
	return mathx.Vec3{X: float32(bestX) + 0.5, Y: float32(bestH) + 1.05, Z: float32(bestZ) + 0.5}
}

// ClearSpawnArea removes vegetation around a fresh spawn so the first view
// is not buried inside a tree canopy.
func (w *World) ClearSpawnArea(pos mathx.Vec3, radius, height int) {
	cx, cz := int(math.Floor(float64(pos.X))), int(math.Floor(float64(pos.Z)))
	y0 := int(math.Floor(float64(pos.Y)))
	for dz := -radius; dz <= radius; dz++ {
		for dx := -radius; dx <= radius; dx++ {
			for y := y0; y <= y0+height; y++ {
				b := w.GetBlock(cx+dx, y, cz+dz)
				switch b {
				case core.Leaves, core.Log, core.TallGrass, core.Flower, core.Cactus:
					w.SetBlock(cx+dx, y, cz+dz, core.Air)
				}
			}
		}
	}
}

func maxAbsInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func sortByDistance(ps []ChunkPos, center ChunkPos) {
	for i := 1; i < len(ps); i++ {
		for j := i; j > 0; j-- {
			a, b := ps[j-1], ps[j]
			da := (a.X-center.X)*(a.X-center.X) + (a.Z-center.Z)*(a.Z-center.Z)
			db := (b.X-center.X)*(b.X-center.X) + (b.Z-center.Z)*(b.Z-center.Z)
			if da <= db {
				break
			}
			ps[j-1], ps[j] = ps[j], ps[j-1]
		}
	}
}

type lightNode struct {
	x, y, z int
	level   uint8
}

// rebuildBlockLightLocked clears and rebuilds block light across every loaded
// chunk. Light sources are sparse, so a single BFS is cheap even with many
// loaded chunks.
func (w *World) rebuildBlockLightLocked() {
	for _, c := range w.chunks {
		clear(c.BlockLight)
	}
	queue := make([]lightNode, 0, 64)
	sources := 0
	for _, c := range w.chunks {
		for i, b := range c.Blocks {
			l := b.Def().Light
			if l == 0 {
				continue
			}
			c.BlockLight[i] = l
			x := i % ChunkSize
			y := i / (ChunkSize * ChunkSize)
			z := (i / ChunkSize) % ChunkSize
			sources++
			queue = append(queue, lightNode{
				x: c.Pos.X*ChunkSize + x, y: y, z: c.Pos.Z*ChunkSize + z, level: l,
			})
		}
	}
	w.lightSources = sources
	dirs := [6][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}}
	for head := 0; head < len(queue); head++ {
		n := queue[head]
		next := n.level - 1
		if next == 0 {
			continue
		}
		for _, d := range dirs {
			x, y, z := n.x+d[0], n.y+d[1], n.z+d[2]
			if y < 0 || y >= ChunkHeight {
				continue
			}
			cp := ChunkOf(x, z)
			c := w.chunks[cp.Key()]
			if c == nil {
				continue
			}
			i := index(Mod(x, ChunkSize), y, Mod(z, ChunkSize))
			if c.Blocks[i].Opaque() || next <= c.BlockLight[i] {
				continue
			}
			c.BlockLight[i] = next
			queue = append(queue, lightNode{x: x, y: y, z: z, level: next})
		}
	}
}

func (w *World) SkyLightAt(wx, wy, wz int) uint8 {
	if wy < 0 || wy >= ChunkHeight {
		return 0
	}
	cp := ChunkOf(wx, wz)
	w.mu.RLock()
	defer w.mu.RUnlock()
	c := w.chunks[cp.Key()]
	if c == nil {
		return 15
	}
	return c.SkyLight[index(Mod(wx, ChunkSize), wy, Mod(wz, ChunkSize))]
}

func (w *World) BlockLightAt(wx, wy, wz int) uint8 {
	if wy < 0 || wy >= ChunkHeight {
		return 0
	}
	cp := ChunkOf(wx, wz)
	w.mu.RLock()
	defer w.mu.RUnlock()
	c := w.chunks[cp.Key()]
	if c == nil {
		return 0
	}
	return c.BlockLight[index(Mod(wx, ChunkSize), wy, Mod(wz, ChunkSize))]
}
