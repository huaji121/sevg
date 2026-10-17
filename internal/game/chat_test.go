package game

import (
	"strings"
	"testing"

	"github.com/Zyko0/go-sdl3/sdl"

	"sevg/internal/i18n"
)

func TestChatEditingAndSend(t *testing.T) {
	g := newTestGame(t)
	g.chatOpen = true
	g.chatInsert("你好 chat")
	if got := string(g.chatInput); got != "你好 chat" {
		t.Fatalf("chat input = %q", got)
	}
	g.chatCaret = 3 // after "你好 "
	g.onChatKey(sdl.K_BACKSPACE)
	if string(g.chatInput) != "你好chat" {
		t.Fatalf("backspace result = %q", string(g.chatInput))
	}
	g.sendChat()
	if g.chatOpen {
		t.Fatal("chat should close after sending")
	}
	if len(g.chatMessages) != 1 || !strings.Contains(g.chatMessages[0].text, "你好chat") {
		t.Fatalf("messages = %+v", g.chatMessages)
	}
	i18n.Set("zh-CN")
}

func TestChatEscape(t *testing.T) {
	g := newTestGame(t)
	g.chatOpen = true
	g.chatInsert("draft")
	g.onChatKey(sdl.K_ESCAPE)
	if g.chatOpen || len(g.chatInput) != 0 {
		t.Fatal("escape did not close and clear chat")
	}
}
func TestGamemodeCommand(t *testing.T) {
	g := newTestGame(t)
	g.chatOpen = true
	g.chatInput = []rune("/gamemode 1")
	g.sendChat()
	if g.mode != ModeCreative || !g.player.Creative {
		t.Fatal("creative mode was not applied")
	}
	if len(g.chatMessages) == 0 {
		t.Fatal("gamemode command produced no system message")
	}
	g.chatOpen = true
	g.chatInput = []rune("/gamemode 0")
	g.sendChat()
	if g.mode != ModeSurvival || g.player.Creative {
		t.Fatal("survival mode was not applied")
	}
}
