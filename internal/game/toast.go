package game

import "sevg/internal/ui"

type toastEntry struct {
	text      string
	remaining float32
	duration  float32
}

const (
	maxToasts   = 6
	toastHeight = float32(40)
	toastGap    = float32(6)
	toastFade   = float32(0.65)
)

func (g *Game) notify(s string, secs float32) {
	if s == "" {
		return
	}
	if secs <= 0 {
		secs = 2.5
	}
	if secs < 0.5 {
		secs = 0.5
	}
	if secs > 30 {
		secs = 30
	}
	entry := toastEntry{text: s, remaining: secs, duration: secs}
	g.toasts = append([]toastEntry{entry}, g.toasts...)
	if len(g.toasts) > maxToasts {
		g.toasts = g.toasts[:maxToasts]
	}
}

func (g *Game) updateToasts(dt float32) {
	if len(g.toasts) == 0 {
		return
	}
	out := g.toasts[:0]
	for _, t := range g.toasts {
		t.remaining -= dt
		if t.remaining > 0 {
			out = append(out, t)
		}
	}
	g.toasts = out
}

func (g *Game) drawToasts(w, h float32) {
	if len(g.toasts) == 0 {
		return
	}
	c := g.canv
	baseY := float32(58)
	if g.state != StatePlaying {
		baseY = 18
	}
	for i, t := range g.toasts {
		fade := float32(1)
		if t.remaining < toastFade {
			fade = t.remaining / toastFade
		}
		if fade < 0 {
			fade = 0
		}
		scale := float32(1.5)
		maxTextW := w - 64
		if tw := c.TextWidth(t.text, scale); tw > maxTextW && tw > 0 {
			scale *= maxTextW / tw
		}
		tw := c.TextWidth(t.text, scale) + 32
		y := baseY + float32(i)*(toastHeight+toastGap)
		x := w/2 - tw/2
		c.Panel(x, y, tw, toastHeight,
			ui.RGBA(20, 25, 30, uint8(230*fade)),
			ui.RGBA(255, 196, 82, uint8(220*fade)))
		col := [4]float32{1, 1, 1, fade}
		g.drawCenteredLabel(ui.Rect{X: x, Y: y, W: tw, H: toastHeight}, t.text, scale, col)
	}
}
