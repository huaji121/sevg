package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"sevg/internal/i18n"
	"sevg/internal/ui"
)

const settingsPath = "settings.json"

type videoSettings struct {
	RenderDist     int   `json:"render_distance"`
	FPSLimit       int   `json:"fps_limit"`
	FOV            int   `json:"fov"`
	Fullscreen     bool  `json:"fullscreen"`
	SmoothLighting *bool `json:"smooth_lighting,omitempty"`
	DynamicLight   *bool `json:"dynamic_light,omitempty"`
	VSync          *bool `json:"vsync,omitempty"`
}

var fpsOptions = []int{0, 30, 60, 75, 120, 144, 240}

func loadVideoSettings(opts *Options) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return
	}
	var s videoSettings
	if json.Unmarshal(data, &s) != nil {
		return
	}
	if s.RenderDist > 0 {
		opts.RenderDist = s.RenderDist
	}
	opts.FPSLimit = s.FPSLimit
	if s.FOV > 0 {
		opts.FOV = s.FOV
	}
	opts.Fullscreen = s.Fullscreen
	if s.SmoothLighting != nil {
		opts.SmoothLighting = *s.SmoothLighting
	}
	if s.DynamicLight != nil {
		opts.DynamicLight = *s.DynamicLight
	}
	if s.VSync != nil {
		opts.VSync = *s.VSync
	}
}

func saveVideoSettings(opts Options) {
	if opts.SmokeFrames > 0 {
		return
	}
	s := videoSettings{
		RenderDist:     opts.RenderDist,
		FPSLimit:       opts.FPSLimit,
		FOV:            opts.FOV,
		Fullscreen:     opts.Fullscreen,
		SmoothLighting: &opts.SmoothLighting,
		DynamicLight:   &opts.DynamicLight,
		VSync:          &opts.VSync,
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(settingsPath)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, ".sevg-settings-*")
	if err != nil {
		return
	}
	name := tmp.Name()
	defer os.Remove(name)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return
	}
	if err := tmp.Close(); err != nil {
		return
	}
	_ = os.Rename(name, settingsPath)
}

func normalizeVideoOptions(opts *Options) {
	opts.RenderDist = clampInt(opts.RenderDist, 3, 16)
	opts.FOV = clampInt(opts.FOV, 50, 110)
	opts.FPSLimit = normalizeFPSLimit(opts.FPSLimit)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func normalizeFPSLimit(v int) int {
	for _, o := range fpsOptions {
		if v == o {
			return v
		}
	}
	return 60
}

func fpsOptionIndex(v int) int {
	for i, o := range fpsOptions {
		if v == o {
			return i
		}
	}
	return 2
}

func (g *Game) openSettings(back State) {
	g.settingsBack = back
	g.setState(StateSettings)
}

func (g *Game) applyVideoChange(kind, delta int) {
	switch kind {
	case 0:
		g.opts.RenderDist = clampInt(g.opts.RenderDist+delta, 3, 16)
		g.streamT = 0
		g.unloadT = 0
	case 1:
		idx := fpsOptionIndex(g.opts.FPSLimit)
		idx = (idx + delta + len(fpsOptions)) % len(fpsOptions)
		g.opts.FPSLimit = fpsOptions[idx]
	case 2:
		g.opts.FOV = clampInt(g.opts.FOV+delta*2, 50, 110)
	case 3:
		g.setFullscreen(!g.fullscreen)
		return
	case 4:
		g.opts.SmoothLighting = !g.opts.SmoothLighting
		if g.world != nil {
			g.world.SetSmoothLighting(g.opts.SmoothLighting)
		}
	case 5:
		g.setVSync(!g.opts.VSync)
		return
	case 6:
		g.opts.DynamicLight = !g.opts.DynamicLight
		g.heldLightT = 0
	}
	saveVideoSettings(g.opts)
}

func (g *Game) setVSync(on bool) {
	g.opts.VSync = on
	if g.gfx != nil {
		if err := g.gfx.SetVSync(on); err != nil {
			g.notify(fmt.Sprintf("VSync: %v", err), 2.5)
		}
	}
	saveVideoSettings(g.opts)
}

func (g *Game) setFullscreen(on bool) {
	g.fullscreen = on
	g.opts.Fullscreen = on
	if g.Win != nil {
		_ = g.Win.SetFullscreen(on)
	}
	saveVideoSettings(g.opts)
}

type settingsLayout struct {
	panel   ui.Rect
	rows    [7]ui.Rect
	minus   [7]ui.Rect
	plus    [7]ui.Rect
	values  [7]ui.Rect
	full    ui.Rect
	smooth  ui.Rect
	vsync   ui.Rect
	dynamic ui.Rect
	back    ui.Rect
}

func (g *Game) settingsRects() settingsLayout {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	pw, ph := float32(760), float32(600)
	if pw > w-32 {
		pw = w - 32
	}
	if ph > h-32 {
		ph = h - 32
	}
	p := ui.Rect{X: (w - pw) / 2, Y: (h - ph) / 2, W: pw, H: ph}
	l := settingsLayout{panel: p}
	rowX := p.X + 28
	rowW := p.W - 56
	rowY := p.Y + 86
	rowH := float32(46)
	gap := float32(8)
	avail := p.H - 86 - 72
	if used := 7*rowH + 6*gap; used > avail {
		gap = 5
		rowH = (avail - 6*gap) / 7
		if rowH < 34 {
			rowH = 34
		}
	}
	for i := 0; i < 7; i++ {
		l.rows[i] = ui.Rect{X: rowX, Y: rowY + float32(i)*(rowH+gap), W: rowW, H: rowH}
		btnY := l.rows[i].Y + (rowH-30)/2
		l.minus[i] = ui.Rect{X: p.X + p.W - 250, Y: btnY, W: 38, H: 30}
		l.plus[i] = ui.Rect{X: p.X + p.W - 72, Y: btnY, W: 38, H: 30}
		l.values[i] = ui.Rect{X: l.minus[i].X + 44, Y: btnY, W: 130, H: 30}
	}
	for i := 3; i <= 6; i++ {
		l.values[i] = ui.Rect{X: p.X + p.W - 246, Y: l.rows[i].Y + (rowH-30)/2, W: 212, H: 30}
	}
	l.full, l.smooth, l.vsync, l.dynamic = l.values[3], l.values[4], l.values[5], l.values[6]
	l.back = ui.Rect{X: p.X + p.W/2 - 110, Y: p.Y + p.H - 62, W: 220, H: 42}
	return l
}

func (g *Game) drawSettings(w, h float32) {
	c := g.canv
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.62})
	l := g.settingsRects()
	p := l.panel
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(18, 23, 28, 248), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+28, p.Y+16, i18n.T("settings.title"), 2.0, ui.TextColor)
	c.Text(p.X+28, p.Y+60, i18n.T("settings.hint"), 0.8, ui.MutedColor)

	labels := []string{
		i18n.T("settings.render_distance"),
		i18n.T("settings.fps_limit"),
		i18n.T("settings.fov"),
		i18n.T("settings.fullscreen"),
		i18n.T("settings.smooth_lighting"),
		i18n.T("settings.vsync"),
		i18n.T("settings.dynamic_light"),
	}
	values := []string{
		fmt.Sprintf("%d", g.opts.RenderDist),
		fpsLimitLabel(g.opts.FPSLimit),
		fmt.Sprintf("%d", g.opts.FOV),
		onOffLabel(g.fullscreen),
		onOffLabel(g.opts.SmoothLighting),
		onOffLabel(g.opts.VSync),
		onOffLabel(g.opts.DynamicLight),
	}
	for i := 0; i < 7; i++ {
		r := l.rows[i]
		c.Rect(r.X, r.Y, r.W, r.H, ui.RGBA(12, 16, 20, 225))
		c.Border(r.X, r.Y, r.W, r.H, 1, ui.RGBA(65, 78, 86, 180))
		g.drawLeftLabel(r.X+20, r, labels[i], 1.15, ui.TextColor)
		if i < 3 {
			g.drawCenteredLabel(l.values[i], values[i], 1.2, ui.Accent)
			g.drawButtonScale(l.minus[i], "-", true, 1.3)
			g.drawButtonScale(l.plus[i], "+", true, 1.3)
		} else if i == 3 {
			g.drawButtonScale(l.full, values[i], true, 1.2)
		} else if i == 4 {
			g.drawButtonScale(l.smooth, values[i], true, 1.2)
		} else if i == 5 {
			g.drawButtonScale(l.vsync, values[i], true, 1.2)
		} else {
			g.drawButtonScale(l.dynamic, values[i], true, 1.2)
		}
	}
	g.drawButtonScale(l.back, i18n.T("settings.back"), true, 1.2)
}

func fpsLimitLabel(v int) string {
	if v <= 0 {
		return i18n.T("settings.unlimited")
	}
	return fmt.Sprintf("%d", v)
}

func onOffLabel(on bool) string {
	if on {
		return i18n.T("settings.on")
	}
	return i18n.T("settings.off")
}
