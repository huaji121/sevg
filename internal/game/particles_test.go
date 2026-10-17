package game

import (
	"testing"

	"sevg/internal/core"
	"sevg/internal/mathx"
	"sevg/internal/world"
)

func TestBreakParticlesSpawnAndExpire(t *testing.T) {
	g := newTestGame(t)
	g.spawnBlockParticles(core.Dirt, [3]int{8, 70, 8}, mathx.V3(0, 1, 0), 8)
	if len(g.particles) != 8 {
		t.Fatalf("particle count = %d", len(g.particles))
	}
	var mesh, transparent world.MeshData
	g.appendParticleMeshes(&mesh, &transparent)
	if len(mesh.Vertices) == 0 || len(mesh.Indices) == 0 {
		t.Fatal("particles produced no geometry")
	}
	g.updateParticles(2)
	if len(g.particles) != 0 {
		t.Fatalf("particles survived expiry: %d", len(g.particles))
	}
}
func TestTransparentParticlesUseBlendedPass(t *testing.T) {
	g := newTestGame(t)
	g.spawnBlockParticles(core.Glass, [3]int{8, 70, 8}, mathx.V3(0, 1, 0), 3)
	var opaque, transparent world.MeshData
	g.appendParticleMeshes(&opaque, &transparent)
	if len(opaque.Vertices) != 0 {
		t.Fatalf("glass particles entered the opaque pass: %d vertices", len(opaque.Vertices))
	}
	if len(transparent.Vertices) != 3*6*4 {
		t.Fatalf("glass particle vertices = %d, want %d", len(transparent.Vertices), 3*6*4)
	}
}
