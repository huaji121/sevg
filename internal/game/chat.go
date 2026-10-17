package game

import (
	"fmt"
	"strings"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/i18n"
	"sevg/internal/ui"
)

type chatMessage struct {
	text string
	age  float32
}

type chatLine struct {
	text string
	age  float32
}

const chatMaxMessages = 100

func (g *Game) openChat() {
	g.openChatWith("")
}

func (g *Game) openChatWith(prefix string) {
	if g.chatOpen || g.state != StatePlaying {
		return
	}
	g.chatOpen = true
	g.chatInput = []rune(prefix)
	g.chatCaret = len(g.chatInput)
	g.chatComposition = ""
	g.chatSuppressSlash = prefix == "/"
	g.chatScroll = 0
	g.mouseLeft = false
	g.mouseRight = false
	g.breakActive = false
	if g.Win != nil {
		_ = g.Win.SetRelativeMouseMode(false)
		_ = sdl.ShowCursor()
		_ = g.Win.StartTextInput()
	}
	g.updateChatInputArea()
}

func (g *Game) closeChat() {
	if !g.chatOpen {
		return
	}
	g.chatOpen = false
	g.chatInput = g.chatInput[:0]
	g.chatCaret = 0
	g.chatComposition = ""
	g.chatSuppressSlash = false
	g.chatScroll = 0
	if g.Win != nil {
		_ = g.Win.StopTextInput()
	}
	if g.state == StatePlaying && g.Win != nil {
		_ = g.Win.SetRelativeMouseMode(true)
		_ = sdl.HideCursor()
	}
}

func (g *Game) updateChatInputArea() {
	if g.Win == nil {
		return
	}
	w, h := g.gfx.W, g.gfx.H
	_ = g.Win.SetTextInputArea(&sdl.Rect{
		X: 20, Y: int32(h) - 116, W: int32(w) - 40, H: 30,
	}, 0)
}

func (g *Game) onChatKey(k sdl.Keycode) {
	switch k {
	case sdl.K_ESCAPE:
		g.closeChat()
	case sdl.K_RETURN, sdl.K_KP_ENTER:
		g.sendChat()
	case sdl.K_BACKSPACE:
		if g.chatCaret > 0 && len(g.chatInput) > 0 {
			g.chatInput = append(g.chatInput[:g.chatCaret-1], g.chatInput[g.chatCaret:]...)
			g.chatCaret--
		}
	case sdl.K_DELETE:
		if g.chatCaret < len(g.chatInput) {
			g.chatInput = append(g.chatInput[:g.chatCaret], g.chatInput[g.chatCaret+1:]...)
		}
	case sdl.K_LEFT:
		if g.chatCaret > 0 {
			g.chatCaret--
		}
	case sdl.K_RIGHT:
		if g.chatCaret < len(g.chatInput) {
			g.chatCaret++
		}
	case sdl.K_HOME:
		g.chatCaret = 0
	case sdl.K_END:
		g.chatCaret = len(g.chatInput)
	case sdl.K_V:
		if sdl.GetModState()&sdl.KMOD_CTRL != 0 {
			if text, err := sdl.GetClipboardText(); err == nil {
				text = strings.ReplaceAll(text, "\n", " ")
				g.chatInsert(text)
			}
		}
	}
}

func (g *Game) chatInsert(text string) {
	if !g.chatOpen {
		return
	}
	if g.chatSuppressSlash {
		g.chatSuppressSlash = false
		if text == "/" {
			return
		}
	}
	text = strings.ReplaceAll(text, "", "")
	if text == "" {
		return
	}
	runes := []rune(text)
	out := make([]rune, 0, len(g.chatInput)+len(runes))
	out = append(out, g.chatInput[:g.chatCaret]...)
	out = append(out, runes...)
	out = append(out, g.chatInput[g.chatCaret:]...)
	g.chatInput = out
	g.chatCaret += len(runes)
	g.chatComposition = ""
}

func (g *Game) sendChat() {
	text := strings.TrimSpace(string(g.chatInput))
	if strings.HasPrefix(text, "/") {
		g.executeCommand(text[1:])
		g.closeChat()
		return
	}
	if text != "" {
		g.addChat(fmt.Sprintf("<%s> %s", i18n.T("chat.you"), text))
	}
	g.closeChat()
}

func (g *Game) addChat(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	g.chatMessages = append(g.chatMessages, chatMessage{text: text})
	if len(g.chatMessages) > chatMaxMessages {
		g.chatMessages = g.chatMessages[len(g.chatMessages)-chatMaxMessages:]
	}
}

func (g *Game) ageChat(dt float32) {
	for i := range g.chatMessages {
		g.chatMessages[i].age += dt
	}
}

func (g *Game) chatLines(maxWidth, scale float32) []chatLine {
	out := make([]chatLine, 0, len(g.chatMessages))
	for _, m := range g.chatMessages {
		for _, line := range wrapChat(g.canv, m.text, maxWidth, scale) {
			out = append(out, chatLine{text: line, age: m.age})
		}
	}
	return out
}

func wrapChat(c *ui.Canvas, text string, maxWidth, scale float32) []string {
	var out []string
	for _, para := range strings.Split(text, "\n") {
		var cur []rune
		for _, r := range []rune(para) {
			cand := string(cur) + string(r)
			if len(cur) > 0 && c.TextWidth(cand, scale) > maxWidth {
				out = append(out, string(cur))
				cur = cur[:0]
			}
			cur = append(cur, r)
		}
		out = append(out, string(cur))
	}
	if len(out) == 0 {
		out = append(out, "")
	}
	return out
}

func (g *Game) scrollChat(delta int) {
	g.chatScroll += delta
	if g.chatScroll < 0 {
		g.chatScroll = 0
	}
}

func (g *Game) drawChat(w, h float32) {
	if g.state != StatePlaying || g.canv == nil {
		return
	}
	c := g.canv
	const scale = float32(1.0)
	const lineH = float32(20)

	if g.chatOpen {
		pw := float32(720)
		if w-40 < pw {
			pw = w - 40
		}
		if pw < 240 {
			pw = 240
		}
		ph := float32(260)
		if h-160 < ph {
			ph = h - 160
		}
		if ph < 150 {
			ph = 150
		}
		x := float32(20)
		y := h - ph - 64
		c.Panel(x, y, pw, ph, ui.RGBA(8, 10, 12, 205), ui.RGBA(170, 170, 170, 190))
		c.Text(x+14, y+8, i18n.T("chat.hint"), 0.85, ui.MutedColor)

		lines := g.chatLines(pw-28, scale)
		messageTop := y + 32
		messageBottom := y + ph - 58
		maxLines := int((messageBottom - messageTop) / lineH)
		if maxLines < 1 {
			maxLines = 1
		}
		maxScroll := len(lines) - maxLines
		if maxScroll < 0 {
			maxScroll = 0
		}
		if g.chatScroll > maxScroll {
			g.chatScroll = maxScroll
		}
		start := len(lines) - maxLines - g.chatScroll
		if start < 0 {
			start = 0
		}
		for i := start; i < len(lines) && i < start+maxLines; i++ {
			yy := messageTop + float32(i-start)*lineH
			c.Text(x+16, yy, lines[i].text, scale, ui.TextColor)
		}
		if maxScroll > 0 {
			trackX := x + pw - 10
			trackH := messageBottom - messageTop
			c.Rect(trackX, messageTop, 3, trackH, [4]float32{1, 1, 1, 0.12})
			thumbH := trackH * float32(maxLines) / float32(len(lines))
			if thumbH < 20 {
				thumbH = 20
			}
			thumbY := messageTop + (trackH-thumbH)*(1-float32(g.chatScroll)/float32(maxScroll))
			c.Rect(trackX, thumbY, 3, thumbH, [4]float32{1, 1, 1, 0.45})
		}

		inputY := y + ph - 40
		c.Rect(x+12, inputY-6, pw-24, 1, [4]float32{1, 1, 1, 0.20})
		prompt := "> "
		input := string(g.chatInput)
		comp := g.chatComposition
		maxWidth := pw - 32
		visible := input
		for len(visible) > 0 && c.TextWidth(prompt+visible+comp, 1.2) > maxWidth {
			_, size := decodeFirstRune(visible)
			visible = visible[size:]
		}
		if input == "" && comp == "" {
			c.Text(x+16, inputY, prompt, 1.2, ui.MutedColor)
			c.Text(x+16+c.TextWidth(prompt, 1.2), inputY, i18n.T("chat.placeholder"), 1.2, ui.MutedColor)
		} else {
			c.Text(x+16, inputY, prompt, 1.2, ui.MutedColor)
			tx := x + 16 + c.TextWidth(prompt, 1.2)
			c.Text(tx, inputY, visible, 1.2, ui.TextColor)
			if comp != "" {
				c.Text(tx+c.TextWidth(visible, 1.2), inputY, comp, 1.2, ui.RGBA(255, 214, 120, 255))
			}
		}
		if g.frame/30%2 == 0 {
			cx := x + 16 + c.TextWidth(prompt+visible+comp, 1.2) + 1
			c.Rect(cx, inputY+2, 2, 22, [4]float32{1, 1, 1, 0.9})
		}
		return
	}

	lines := g.chatLines(minF(w-40, 700), scale)
	if len(lines) == 0 {
		return
	}
	maxLines := 8
	start := len(lines) - maxLines
	if start < 0 {
		start = 0
	}
	baseY := h - 92
	for i := start; i < len(lines); i++ {
		alpha := float32(1)
		if lines[i].age > 8 {
			alpha = 1 - (lines[i].age-8)/2
		}
		if alpha <= 0 {
			continue
		}
		row := len(lines) - 1 - i
		yy := baseY - float32(row)*lineH
		tw := c.TextWidth(lines[i].text, scale)
		c.Rect(16, yy-2, tw+12, lineH, [4]float32{0, 0, 0, 0.30 * alpha})
		c.Text(22, yy, lines[i].text, scale, [4]float32{1, 1, 1, alpha})
	}
}

func decodeFirstRune(s string) (rune, int) {
	for i, r := range s {
		if i == 0 {
			return r, len(string(r))
		}
	}
	return 0, 0
}

func minF(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
