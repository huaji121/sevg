package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/i18n"
	"sevg/internal/ui"
)

const settingsPath = "settings.json"

// Keybinds stores the configurable gameplay keys.
type Keybinds struct {
	Forward   sdl.Keycode `json:"forward"`
	Back      sdl.Keycode `json:"back"`
	Left      sdl.Keycode `json:"left"`
	Right     sdl.Keycode `json:"right"`
	Jump      sdl.Keycode `json:"jump"`
	Inventory sdl.Keycode `json:"inventory"`
	Drop      sdl.Keycode `json:"drop"`
	Chat      sdl.Keycode `json:"chat"`
}

func DefaultKeybinds() Keybinds {
	return Keybinds{
		Forward: sdl.K_W, Back: sdl.K_S, Left: sdl.K_A, Right: sdl.K_D,
		Jump: sdl.K_SPACE, Inventory: sdl.K_E, Drop: sdl.K_Q, Chat: sdl.K_T,
	}
}

type videoSettings struct {
	RenderDist     int       `json:"render_distance"`
	FPSLimit       int       `json:"fps_limit"`
	FOV            int       `json:"fov"`
	Fullscreen     bool      `json:"fullscreen"`
	SmoothLighting *bool     `json:"smooth_lighting,omitempty"`
	DynamicLight   *bool     `json:"dynamic_light,omitempty"`
	VSync          *bool     `json:"vsync,omitempty"`
	MasterVolume   *int      `json:"master_volume,omitempty"`
	Language       string    `json:"language,omitempty"`
	Keys           *Keybinds `json:"keys,omitempty"`
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
	if s.MasterVolume != nil {
		opts.MasterVolume = clampInt(*s.MasterVolume, 0, 100)
	}
	if s.Language == i18n.ZH || s.Language == i18n.EN {
		opts.Language = s.Language
	}
	if s.Keys != nil {
		opts.Keys = *s.Keys
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
		MasterVolume:   &opts.MasterVolume,
		Language:       opts.Language,
		Keys:           &opts.Keys,
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
	tabs    [4]ui.Rect
	back    ui.Rect
	rows    [7]ui.Rect
	values  [7]ui.Rect
	sliders [3]ui.Rect
	full    ui.Rect
	smooth  ui.Rect
	vsync   ui.Rect
	dynamic ui.Rect

	audioValue  ui.Rect
	audioSlider ui.Rect
	lang        [2]ui.Rect
	keyRows     [8]ui.Rect
	keyButtons  [8]ui.Rect
}

func (g *Game) settingsRects() settingsLayout {
	w, h := float32(g.gfx.W), float32(g.gfx.H)
	pw, ph := float32(760), float32(640)
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
	tabGap := float32(8)
	tabW := (rowW - tabGap*3) / 4
	for i := 0; i < 4; i++ {
		l.tabs[i] = ui.Rect{X: rowX + float32(i)*(tabW+tabGap), Y: p.Y + 82, W: tabW, H: 32}
	}
	rowY := p.Y + 124
	rowH := float32(46)
	gap := float32(8)
	avail := p.H - 124 - 70
	if used := 7*rowH + 6*gap; used > avail {
		gap = 5
		rowH = (avail - 6*gap) / 7
		if rowH < 32 {
			rowH = 32
		}
	}
	for i := 0; i < 7; i++ {
		l.rows[i] = ui.Rect{X: rowX, Y: rowY + float32(i)*(rowH+gap), W: rowW, H: rowH}
		btnY := l.rows[i].Y + (rowH-30)/2
		l.values[i] = ui.Rect{X: p.X + p.W - 150, Y: btnY, W: 110, H: 30}
	}
	for i := 0; i < 3; i++ {
		l.sliders[i] = ui.Rect{X: p.X + 300, Y: l.rows[i].Y + (rowH-30)/2, W: p.W - 300 - 160, H: 30}
	}
	for i := 3; i <= 6; i++ {
		l.values[i] = ui.Rect{X: p.X + p.W - 246, Y: l.rows[i].Y + (rowH-30)/2, W: 212, H: 30}
	}
	l.full, l.smooth, l.vsync, l.dynamic = l.values[3], l.values[4], l.values[5], l.values[6]

	audioY := l.rows[0].Y + (rowH-30)/2
	l.audioValue = ui.Rect{X: p.X + p.W - 150, Y: audioY, W: 110, H: 30}
	l.audioSlider = ui.Rect{X: rowX + 190, Y: audioY, W: rowW - 190 - 160, H: 30}

	l.lang[0] = ui.Rect{X: rowX + 40, Y: rowY + 60, W: rowW - 80, H: 52}
	l.lang[1] = ui.Rect{X: rowX + 40, Y: rowY + 128, W: rowW - 80, H: 52}

	colW := (rowW - 12) / 2
	const keyRowH = float32(54)
	for i := 0; i < 8; i++ {
		col, row := i%2, i/2
		l.keyRows[i] = ui.Rect{X: rowX + float32(col)*(colW+12), Y: rowY + float32(row)*(keyRowH+8), W: colW, H: keyRowH}
		l.keyButtons[i] = ui.Rect{X: l.keyRows[i].X + l.keyRows[i].W - 168, Y: l.keyRows[i].Y + 12, W: 148, H: 30}
	}
	l.back = ui.Rect{X: p.X + p.W/2 - 110, Y: p.Y + p.H - 62, W: 220, H: 42}
	return l
}

func (g *Game) drawSettings(w, h float32) {
	c := g.canv
	c.Rect(0, 0, w, h, [4]float32{0, 0, 0, 0.62})
	l := g.settingsRects()
	p := l.panel
	c.Panel(p.X, p.Y, p.W, p.H, ui.RGBA(18, 23, 28, 248), ui.RGBA(255, 196, 82, 220))
	c.Text(p.X+28, p.Y+10, i18n.T("settings.title"), 2.0, ui.TextColor)
	c.Text(p.X+28, p.Y+54, i18n.T("settings.hint"), 0.78, ui.MutedColor)

	tabLabels := []string{
		i18n.T("settings.tab_video"), i18n.T("settings.tab_audio"),
		i18n.T("settings.tab_language"), i18n.T("settings.tab_controls"),
	}
	for i, r := range l.tabs {
		bg := ui.RGBA(16, 20, 24, 230)
		bd := ui.RGBA(70, 82, 90, 200)
		col := ui.MutedColor
		if i == g.settingsTab {
			bg = ui.RGBA(44, 51, 58, 245)
			bd = ui.Accent
			col = ui.TextColor
		}
		c.Rect(r.X, r.Y, r.W, r.H, bg)
		c.Border(r.X, r.Y, r.W, r.H, 1, bd)
		g.drawCenteredLabel(r, tabLabels[i], 1.0, col)
	}

	switch g.settingsTab {
	case 0:
		g.drawVideoSettings(l)
	case 1:
		g.drawAudioSettings(l)
	case 2:
		g.drawLanguageSettings(l)
	case 3:
		g.drawControlSettings(l)
	}
	g.drawButtonScale(l.back, i18n.T("settings.back"), true, 1.2)
}

func (g *Game) drawVideoSettings(l settingsLayout) {
	c := g.canv
	labels := []string{
		i18n.T("settings.render_distance"), i18n.T("settings.fps_limit"),
		i18n.T("settings.fov"), i18n.T("settings.fullscreen"),
		i18n.T("settings.smooth_lighting"), i18n.T("settings.vsync"),
		i18n.T("settings.dynamic_light"),
	}
	values := []string{
		fmt.Sprintf("%d", g.opts.RenderDist), fpsLimitLabel(g.opts.FPSLimit),
		fmt.Sprintf("%d", g.opts.FOV), onOffLabel(g.fullscreen),
		onOffLabel(g.opts.SmoothLighting), onOffLabel(g.opts.VSync),
		onOffLabel(g.opts.DynamicLight),
	}
	for i := 0; i < 7; i++ {
		r := l.rows[i]
		c.Rect(r.X, r.Y, r.W, r.H, ui.RGBA(12, 16, 20, 225))
		c.Border(r.X, r.Y, r.W, r.H, 1, ui.RGBA(65, 78, 86, 180))
		g.drawLeftLabel(r.X+20, r, labels[i], 1.15, ui.TextColor)
		switch i {
		case 0, 1, 2:
			g.drawCenteredLabel(l.values[i], values[i], 1.2, ui.Accent)
			g.drawSlider(l.sliders[i], g.videoSliderFraction(i))
		case 3:
			g.drawButtonScale(l.full, values[i], true, 1.2)
		case 4:
			g.drawButtonScale(l.smooth, values[i], true, 1.2)
		case 5:
			g.drawButtonScale(l.vsync, values[i], true, 1.2)
		case 6:
			g.drawButtonScale(l.dynamic, values[i], true, 1.2)
		}
	}
}

func (g *Game) drawAudioSettings(l settingsLayout) {
	c := g.canv
	r := l.rows[0]
	c.Rect(r.X, r.Y, r.W, r.H, ui.RGBA(12, 16, 20, 225))
	c.Border(r.X, r.Y, r.W, r.H, 1, ui.RGBA(65, 78, 86, 180))
	g.drawLeftLabel(r.X+20, r, i18n.T("settings.master_volume"), 1.15, ui.TextColor)
	g.drawCenteredLabel(l.audioValue, fmt.Sprintf("%d%%", g.opts.MasterVolume), 1.2, ui.Accent)
	g.drawSlider(l.audioSlider, float32(g.opts.MasterVolume)/100)
}

func (g *Game) drawLanguageSettings(l settingsLayout) {
	langs := []string{i18n.ZH, i18n.EN}
	for i, r := range l.lang {
		cur := i18n.Current() == langs[i]
		bg := ui.RGBA(16, 20, 24, 225)
		bd := ui.RGBA(65, 78, 86, 180)
		if cur {
			bg = ui.RGBA(42, 50, 56, 235)
			bd = ui.Accent
		}
		g.canv.Rect(r.X, r.Y, r.W, r.H, bg)
		g.canv.Border(r.X, r.Y, r.W, r.H, 2, bd)
		g.drawCenteredLabel(r, i18n.LocaleName(langs[i]), 1.2, ui.TextColor)
	}
}

func (g *Game) drawControlSettings(l settingsLayout) {
	labels := []string{
		i18n.T("settings.control_forward"), i18n.T("settings.control_back"),
		i18n.T("settings.control_left"), i18n.T("settings.control_right"),
		i18n.T("settings.control_jump"), i18n.T("settings.control_inventory"),
		i18n.T("settings.control_drop"), i18n.T("settings.control_chat"),
	}
	for i, r := range l.keyRows {
		g.canv.Rect(r.X, r.Y, r.W, r.H, ui.RGBA(12, 16, 20, 225))
		g.canv.Border(r.X, r.Y, r.W, r.H, 1, ui.RGBA(65, 78, 86, 180))
		g.drawLeftLabel(r.X+18, r, labels[i], 1.05, ui.TextColor)
		label := keybindLabel(g.bindingKey(i))
		if g.bindingAction == i {
			label = i18n.T("settings.press_key")
		}
		g.drawButtonScale(l.keyButtons[i], label, true, 1.0)
	}
}

func (g *Game) bindingKey(i int) sdl.Keycode {
	switch i {
	case 0:
		return g.keys.Forward
	case 1:
		return g.keys.Back
	case 2:
		return g.keys.Left
	case 3:
		return g.keys.Right
	case 4:
		return g.keys.Jump
	case 5:
		return g.keys.Inventory
	case 6:
		return g.keys.Drop
	case 7:
		return g.keys.Chat
	}
	return 0
}

func (g *Game) setBinding(i int, k sdl.Keycode) {
	switch i {
	case 0:
		g.keys.Forward = k
	case 1:
		g.keys.Back = k
	case 2:
		g.keys.Left = k
	case 3:
		g.keys.Right = k
	case 4:
		g.keys.Jump = k
	case 5:
		g.keys.Inventory = k
	case 6:
		g.keys.Drop = k
	case 7:
		g.keys.Chat = k
	}
	g.opts.Keys = g.keys
	saveVideoSettings(g.opts)
}

func keybindLabel(k sdl.Keycode) string {
	if k == 0 {
		return "?"
	}
	return k.KeyName()
}

func (g *Game) setLanguage(lang string) {
	if lang != i18n.ZH && lang != i18n.EN {
		return
	}
	i18n.Set(lang)
	g.opts.Language = lang
	if g.Win != nil {
		_ = g.Win.SetTitle(i18n.T("window.title"))
	}
	saveVideoSettings(g.opts)
}

func (g *Game) applyAudioChange(delta int) {
	g.opts.MasterVolume = clampInt(g.opts.MasterVolume+delta, 0, 100)
	if g.audio != nil {
		g.audio.SetMasterVolume(float32(g.opts.MasterVolume) / 100)
	}
	saveVideoSettings(g.opts)
}

func (g *Game) drawSlider(r ui.Rect, t float32) {
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	c := g.canv
	c.Rect(r.X, r.Y+r.H/2-4, r.W, 8, ui.RGBA(40, 47, 54, 235))
	c.Border(r.X, r.Y+r.H/2-4, r.W, 8, 1, ui.RGBA(80, 95, 105, 220))
	fill := r.W * t
	c.Rect(r.X, r.Y+r.H/2-4, fill, 8, ui.Accent)
	kx := r.X + fill - 7
	if kx < r.X {
		kx = r.X
	}
	if kx+14 > r.X+r.W {
		kx = r.X + r.W - 14
	}
	c.Rect(kx, r.Y+r.H/2-8, 14, 16, ui.TextColor)
}

func (g *Game) videoSliderFraction(i int) float32 {
	switch i {
	case 0:
		return float32(g.opts.RenderDist-3) / 13
	case 1:
		return float32(fpsOptionIndex(g.opts.FPSLimit)) / float32(len(fpsOptions)-1)
	case 2:
		return float32(g.opts.FOV-50) / 60
	}
	return 0
}

func sliderPercent(r ui.Rect, x float32) int {
	if r.W <= 0 {
		return 0
	}
	t := (x - r.X) / r.W
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}
	return int(t*100 + 0.5)
}

func (g *Game) setMasterVolume(v int) {
	g.opts.MasterVolume = clampInt(v, 0, 100)
	if g.audio != nil {
		g.audio.SetMasterVolume(float32(g.opts.MasterVolume) / 100)
	}
}

func (g *Game) updateSliderDrag() {
	l := g.settingsRects()
	switch g.dragSlider {
	case 0:
		g.opts.RenderDist = 3 + sliderPercent(l.sliders[0], g.mouseX)*13/100
		g.streamT = 0
		g.unloadT = 0
	case 1:
		idx := sliderPercent(l.sliders[1], g.mouseX) * (len(fpsOptions) - 1) / 100
		g.opts.FPSLimit = fpsOptions[idx]
	case 2:
		g.opts.FOV = 50 + sliderPercent(l.sliders[2], g.mouseX)*60/100
	case 3:
		g.setMasterVolume(sliderPercent(l.audioSlider, g.mouseX))
	}
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
