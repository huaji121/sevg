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
