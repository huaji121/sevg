package world

import (
	"testing"

	"sevg/internal/core"
)

func TestTreeCanopyCrossesChunkBorders(t *testing.T) {
	w := New(12345)
	for x := -2; x <= 2; x++ {
		for z := -2; z <= 2; z++ {
			w.EnsureChunk(ChunkPos{x, z})
		}
	}
	check := func(a, b *Chunk, ax, az, bx, bz int) {
		for y := 1; y < ChunkHeight-3; y++ {
			if a.Blocks[index(ax, y, az)] != core.Log {
				continue
			}
			if a.Blocks[index(ax, y-1, az)] == core.Log {
				continue
			}
			top := y
			for top+1 < ChunkHeight && a.Blocks[index(ax, top+1, az)] == core.Log {
				top++
			}
			if top+1 >= ChunkHeight {
				continue
			}
			if b.Blocks[index(bx, top+1, bz)] == core.Air {
				t.Fatalf("tree at chunk %v local(%d,%d) missing canopy in neighbour", a.Pos, ax, az)
			}
		}
	}
	for cx := -2; cx <= 1; cx++ {
		for cz := -2; cz <= 2; cz++ {
			a, _ := w.EnsureChunk(ChunkPos{cx, cz})
			b, _ := w.EnsureChunk(ChunkPos{cx + 1, cz})
			for z := 0; z < ChunkSize; z++ {
				check(a, b, ChunkSize-1, z, 0, z)
			}
			c, _ := w.EnsureChunk(ChunkPos{cx, cz + 1})
			for x := 0; x < ChunkSize; x++ {
				check(a, c, x, ChunkSize-1, x, 0)
			}
		}
	}
}
