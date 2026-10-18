package game

import (
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/i18n"

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
		if !l.rows[i].Contains(l.sliders[i].X+1, l.sliders[i].Y+1) {
			t.Fatalf("slider %d is outside its row", i)
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
func TestSettingsCategories(t *testing.T) {
	g := newTestGame(t)
	g.opts.SmokeFrames = 1
	g.state = StateSettings
	g.bindingAction = 0
	g.onKey(sdl.K_I)
	if g.keys.Forward != sdl.K_I || g.opts.Keys.Forward != sdl.K_I || g.bindingAction != -1 {
		t.Fatalf("rebind failed: keys=%+v binding=%d", g.keys, g.bindingAction)
	}
	if g.bindingKey(0) != sdl.K_I {
		t.Fatal("bindingKey did not return rebound key")
	}
	g.opts.MasterVolume = 50
	g.applyAudioChange(10)
	if g.opts.MasterVolume != 60 {
		t.Fatalf("master volume = %d", g.opts.MasterVolume)
	}
	old := i18n.Current()
	defer i18n.Set(old)
	g.setLanguage(i18n.EN)
	if i18n.Current() != i18n.EN || g.opts.Language != i18n.EN {
		t.Fatalf("language = %q opts = %q", i18n.Current(), g.opts.Language)
	}
}
func TestAudioSliderDrag(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.opts.SmokeFrames = 1
	g.opts.MasterVolume = 70
	g.state = StateSettings
	g.settingsTab = 1
	l := g.settingsRects()
	g.mouseX = l.audioSlider.X + l.audioSlider.W*0.25
	g.mouseY = l.audioSlider.Y + l.audioSlider.H/2
	g.handleUIClick(false)
	if g.opts.MasterVolume < 24 || g.opts.MasterVolume > 26 {
		t.Fatalf("volume after click = %d", g.opts.MasterVolume)
	}
	g.mouseX = l.audioSlider.X + l.audioSlider.W*0.8
	g.updateSliderDrag()
	if g.opts.MasterVolume < 79 || g.opts.MasterVolume > 81 {
		t.Fatalf("volume after drag = %d", g.opts.MasterVolume)
	}
	g.onMouseUp(uint8(sdl.BUTTON_LEFT))
	if g.dragSlider != -1 {
		t.Fatalf("drag state = %d", g.dragSlider)
	}
}
func TestVideoSlidersAdjustValues(t *testing.T) {
	g := newTestGame(t)
	g.gfx = &render.Renderer{W: 1280, H: 720}
	g.opts.SmokeFrames = 1
	g.state = StateSettings
	g.settingsTab = 0
	l := g.settingsRects()
	g.mouseX = l.sliders[0].X + l.sliders[0].W/2
	g.mouseY = l.sliders[0].Y + l.sliders[0].H/2
	g.handleUIClick(false)
	if g.opts.RenderDist < 9 || g.opts.RenderDist > 10 {
		t.Fatalf("render distance = %d", g.opts.RenderDist)
	}
	g.dragSlider = -1
	g.mouseX = l.sliders[2].X + l.sliders[2].W*0.75
	g.mouseY = l.sliders[2].Y + l.sliders[2].H/2
	g.handleUIClick(false)
	if g.opts.FOV < 94 || g.opts.FOV > 96 {
		t.Fatalf("fov = %d", g.opts.FOV)
	}
}
