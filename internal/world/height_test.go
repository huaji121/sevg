package world

import (
	"testing"

	"sevg/internal/core"
)

func TestHighMountainsInTallWorld(t *testing.T) {
	if ChunkHeight != 256 {
		t.Fatalf("chunk height = %d", ChunkHeight)
	}
	g := NewGenerator(12345)
	maxH := 0
	high := 0
	for x := -2048; x <= 2048; x += 16 {
		for z := -2048; z <= 2048; z += 16 {
			h, b := g.HeightAt(x, z)
			if h < 0 || h >= ChunkHeight {
				t.Fatalf("height %d outside world at %d,%d", h, x, z)
			}
			if h > maxH {
				maxH = h
			}
			if b == BiomeHighMountains {
				high++
				if h < SeaLevel+97 {
					t.Fatalf("high mountain too low: %d", h)
				}
			}
		}
	}
	if maxH < 180 {
		t.Fatalf("tallest peak = %d, want alpine terrain", maxH)
	}
	if high == 0 {
		t.Fatal("no high mountain biome generated")
	}
}

func TestOreNeverExposedOnSurface(t *testing.T) {
	g := NewGenerator(12345)
	ore := 0
	for _, cp := range []ChunkPos{{0, 0}, {3, -2}, {-5, 7}, {8, 8}} {
		blocks := g.GenerateChunk(cp)
		for lx := 0; lx < ChunkSize; lx++ {
			for lz := 0; lz < ChunkSize; lz++ {
				wx, wz := cp.X*ChunkSize+lx, cp.Z*ChunkSize+lz
				h, _ := g.HeightAt(wx, wz)
				if h < 0 || h >= ChunkHeight {
					continue
				}
				b := blocks[index(lx, h, lz)]
				if b == core.CoalOre || b == core.IronOre {
					t.Fatalf("ore on the surface at %d,%d: %v", wx, wz, b)
				}
			}
		}
		for _, b := range blocks {
			if b == core.CoalOre || b == core.IronOre {
				ore++
			}
		}
	}
	if ore == 0 {
		t.Fatal("no ore generated underground")
	}
}
