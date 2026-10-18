package world

import "testing"

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
				if h < 145 {
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
