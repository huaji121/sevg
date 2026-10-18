package game

import (
	"testing"

	"sevg/internal/render"
)

func TestNormalizeVideoOptions(t *testing.T) {
	opts := Options{RenderDist: 99, FOV: 20, FPSLimit: 33}
	normalizeVideoOptions(&opts)
	if opts.RenderDist != 16 || opts.FOV != 50 || opts.FPSLimit != 60 {
		t.Fatalf("normalized options = %+v", opts)
	}
}

func TestFPSLimitCycle(t *testing.T) {
	g := newTestGame(t)
	g.opts.SmokeFrames = 1
	g.opts.FPSLimit = 0
	g.applyVideoChange(1, -1)
	if g.opts.FPSLimit != 240 {
		t.Fatalf("fps wrap down = %d", g.opts.FPSLimit)
	}
	g.applyVideoChange(1, 1)
	if g.opts.FPSLimit != 0 {
		t.Fatalf("fps wrap up = %d", g.opts.FPSLimit)
	}
}

func TestSettingsLayout(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	l := g.settingsRects()
	if l.panel.W <= 0 || l.back.W <= 0 {
		t.Fatal("settings panel has invalid geometry")
	}
	for i := 0; i < 3; i++ {
		if !l.rows[i].Contains(l.minus[i].X+1, l.minus[i].Y+1) {
			t.Fatalf("minus button %d is outside its row", i)
		}
	}
	if !l.rows[6].Contains(l.dynamic.X+1, l.dynamic.Y+1) {
		t.Fatal("dynamic light button is outside its row")
	}
	if l.rows[6].Y+l.rows[6].H > l.back.Y {
		t.Fatal("settings rows overlap the back button")
	}
}
func TestDynamicLightToggle(t *testing.T) {
	g := newTestGame(t)
	g.opts.SmokeFrames = 1
	g.opts.DynamicLight = true
	g.world.SetDynamicLight(8, 70, 8, 14)
	if g.world.BlockLightAt(8, 70, 8) == 0 {
		t.Fatal("dynamic light was not applied")
	}
	g.applyVideoChange(6, 0)
	if g.opts.DynamicLight {
		t.Fatal("dynamic light option did not turn off")
	}
	g.updateHeldLight(0.1)
	if got := g.world.BlockLightAt(8, 70, 8); got != 0 {
		t.Fatalf("dynamic light remains after disabling: %d", got)
	}
}
func TestSettingsLayoutSmallWindow(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 800, H: 500}
	l := g.settingsRects()
	if l.panel.H > 468 {
		t.Fatalf("settings panel too tall: %f", l.panel.H)
	}
	if l.rows[6].Y+l.rows[6].H > l.back.Y {
		t.Fatalf("rows overlap back button: row end %f back %f", l.rows[6].Y+l.rows[6].H, l.back.Y)
	}
}
