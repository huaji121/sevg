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
func TestSlashOpensChatWithCommandPrefix(t *testing.T) {
	g := newTestGame(t)
	g.state = StatePlaying
	g.onKey(sdl.K_SLASH)
	if !g.chatOpen || string(g.chatInput) != "/" || g.chatCaret != 1 {
		t.Fatalf("slash open state: open=%v input=%q caret=%d", g.chatOpen, string(g.chatInput), g.chatCaret)
	}
	// The same key press may also produce a TEXT_INPUT event; it must not
	// create a second slash.
	g.chatInsert("/")
	if string(g.chatInput) != "/" {
		t.Fatalf("duplicate slash: %q", string(g.chatInput))
	}
	g.chatInsert("gamemode 1")
	g.sendChat()
	if g.mode != ModeCreative {
		t.Fatal("slash command did not switch mode")
	}
}
func TestChatScroll(t *testing.T) {
	g := newTestGame(t)
	g.scrollChat(-5)
	if g.chatScroll != 0 {
		t.Fatalf("negative chat scroll: %d", g.chatScroll)
	}
	g.scrollChat(6)
	if g.chatScroll != 6 {
		t.Fatalf("chat scroll: %d", g.chatScroll)
	}
}
func TestChatArrowKeysMoveCaret(t *testing.T) {
	g := newTestGame(t)
	g.chatOpen = true
	g.chatInsert("abcd")
	g.onChatKey(sdl.K_LEFT)
	g.onChatKey(sdl.K_LEFT)
	if g.chatCaret != 2 {
		t.Fatalf("left caret = %d", g.chatCaret)
	}
	g.chatInsert("X")
	if string(g.chatInput) != "abXcd" || g.chatCaret != 3 {
		t.Fatalf("mid insertion = %q caret %d", string(g.chatInput), g.chatCaret)
	}
	g.onChatKey(sdl.K_HOME)
	if g.chatCaret != 0 {
		t.Fatalf("home caret = %d", g.chatCaret)
	}
	g.onChatKey(sdl.K_RIGHT)
	if g.chatCaret != 1 {
		t.Fatalf("right caret = %d", g.chatCaret)
	}
	g.onChatKey(sdl.K_END)
	if g.chatCaret != 5 {
		t.Fatalf("end caret = %d", g.chatCaret)
	}
}
func TestToastCommand(t *testing.T) {
	g := newTestGame(t)
	g.chatOpen = true
	g.chatInput = []rune("/toast 1.5 hello world")
	g.sendChat()
	if g.toast != "hello world" {
		t.Fatalf("toast = %q", g.toast)
	}
	if g.toastTime != 1.5 {
		t.Fatalf("toast duration = %f", g.toastTime)
	}
}
