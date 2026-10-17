package world

import "sevg/internal/core"

const maxWaterLevel = 7

func (w *World) queueFluidLocked(x, y, z int) {
	if y < 0 || y >= ChunkHeight {
		return
	}
	w.fluidQueue[[3]int{x, y, z}] = true
	for _, d := range [6][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}} {
		ny := y + d[1]
		if ny >= 0 && ny < ChunkHeight {
			w.fluidQueue[[3]int{x + d[0], ny, z + d[2]}] = true
		}
	}
}

func (w *World) waterLevelUnlocked(x, y, z int) (uint8, bool) {
	if y < 0 || y >= ChunkHeight {
		return 0, false
	}
	cp := ChunkOf(x, z)
	c := w.chunks[cp.Key()]
	if c == nil {
		return 0, false
	}
	i := index(Mod(x, ChunkSize), y, Mod(z, ChunkSize))
	if c.Blocks[i] != core.Water {
		return 0, false
	}
	return c.WaterLevel[i], true
}

func (w *World) setWaterLocked(x, y, z int, level uint8) bool {
	if y < 0 || y >= ChunkHeight {
		return false
	}
	cp := ChunkOf(x, z)
	c := w.chunks[cp.Key()]
	if c == nil {
		return false
	}
	lx, lz := Mod(x, ChunkSize), Mod(z, ChunkSize)
	i := index(lx, y, lz)
	if c.Blocks[i] == core.Water && c.WaterLevel[i] == level {
		return false
	}
	c.Blocks[i] = core.Water
	c.WaterLevel[i] = level
	c.Dirty = true
	w.markFluidDirtyLocked(cp, lx, lz)
	w.queueFluidLocked(x, y, z)
	return true
}

func (w *World) clearWaterLocked(x, y, z int) bool {
	cp := ChunkOf(x, z)
	c := w.chunks[cp.Key()]
	if c == nil {
		return false
	}
	lx, lz := Mod(x, ChunkSize), Mod(z, ChunkSize)
	i := index(lx, y, lz)
	if c.Blocks[i] != core.Water {
		return false
	}
	c.Blocks[i] = core.Air
	c.WaterLevel[i] = 0
	c.Dirty = true
	w.markFluidDirtyLocked(cp, lx, lz)
	w.queueFluidLocked(x, y, z)
	return true
}

func (w *World) markFluidDirtyLocked(cp ChunkPos, lx, lz int) {
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
}

// TickFluids processes a bounded batch of the active fluid cells. The bound
// keeps streaming and rendering responsive even when a large body of water is
// opened at once.
func (w *World) TickFluids() {
	w.mu.Lock()
	defer w.mu.Unlock()
	const batch = 256
	n := 0
	for p := range w.fluidQueue {
		if n >= batch {
			break
		}
		delete(w.fluidQueue, p)
		w.updateFluidLocked(p[0], p[1], p[2])
		n++
	}
}

func (w *World) updateFluidLocked(x, y, z int) {
	b := w.blockUnlocked(x, y, z)
	if b != core.Water {
		return
	}
	level, ok := w.waterLevelUnlocked(x, y, z)
	if !ok {
		return
	}
	supplierX, supplierZ := -1, -1
	if level > 0 {
		desired, supplied, sx, _, sz := w.suppliedWaterLevelLocked(x, y, z, level)
		if !supplied {
			w.clearWaterLocked(x, y, z)
			return
		}
		supplierX, supplierZ = sx, sz
		if desired < level {
			w.setWaterLocked(x, y, z, desired)
			level = desired
		}
	}

	// Falling water always fills the cell below first.
	if y > 0 && w.blockUnlocked(x, y-1, z) == core.Air {
		w.setWaterLocked(x, y-1, z, 1)
	}

	next := level + 1
	if level == 0 {
		next = 1
	}
	if next > maxWaterLevel {
		return
	}
	for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		nx, nz := x+d[0], z+d[1]
		// Do not grow back into the cell that supplied this flowing water.
		if nx == supplierX && nz == supplierZ {
			continue
		}
		if w.blockUnlocked(nx, y, nz) == core.Air {
			w.setWaterLocked(nx, y, nz, next)
		}
	}
}

func (w *World) suppliedWaterLevelLocked(x, y, z int, current uint8) (uint8, bool, int, int, int) {
	if y+1 < ChunkHeight && w.blockUnlocked(x, y+1, z) == core.Water {
		return 1, true, x, y + 1, z
	}
	best := uint8(255)
	bx, bz := 0, 0
	for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		if lvl, ok := w.waterLevelUnlocked(x+d[0], y, z+d[1]); ok && lvl < current && lvl < best {
			best = lvl
			bx, bz = x+d[0], z+d[1]
		}
	}
	if best == 255 {
		return 0, false, 0, 0, 0
	}
	return best + 1, true, bx, y, bz
}

func (w *World) WaterLevelAt(wx, wy, wz int) uint8 {
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
	return c.WaterLevel[index(Mod(wx, ChunkSize), wy, Mod(wz, ChunkSize))]
}

// drainLocalWaterLocked removes a flowing-water component that lost its last
// source. It is only invoked when a source block itself is replaced, so a
// normal ocean with neighbouring sources is left intact.
func (w *World) drainLocalWaterLocked(x, y, z int) {
	dirs := [6][3]int{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}}
	seen := make(map[[3]int]bool)
	queue := make([][3]int, 0, 32)
	start := func(nx, ny, nz int) {
		if ny < 0 || ny >= ChunkHeight {
			return
		}
		p := [3]int{nx, ny, nz}
		if !seen[p] {
			seen[p] = true
			queue = append(queue, p)
		}
	}
	for _, d := range dirs {
		if lvl, ok := w.waterLevelUnlocked(x+d[0], y+d[1], z+d[2]); ok && lvl > 0 {
			start(x+d[0], y+d[1], z+d[2])
		}
	}
	for head := 0; head < len(queue) && head < 4096; head++ {
		p := queue[head]
		if lvl, ok := w.waterLevelUnlocked(p[0], p[1], p[2]); !ok || lvl == 0 {
			continue
		}
		for _, d := range dirs {
			nx, ny, nz := p[0]+d[0], p[1]+d[1], p[2]+d[2]
			if lvl, ok := w.waterLevelUnlocked(nx, ny, nz); ok && lvl == 0 {
				return // still connected to another source
			} else if ok && lvl > 0 {
				key := [3]int{nx, ny, nz}
				if !seen[key] {
					seen[key] = true
					queue = append(queue, key)
				}
			}
		}
	}
	for _, p := range queue[:min(len(queue), 4096)] {
		w.clearWaterLocked(p[0], p[1], p[2])
	}
}
