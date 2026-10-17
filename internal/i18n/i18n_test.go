package i18n

import (
	"strings"
	"testing"

	"sevg/internal/core"
)

func TestDefaultChinese(t *testing.T) {
	defer Set(ZH)
	Set(ZH)
	if got := T("menu.new"); got != "新世界" {
		t.Fatalf("menu.new = %q", got)
	}
	if got := ItemName(core.ItemPlanks); got != "橡木木板" {
		t.Fatalf("plank name = %q", got)
	}
}

func TestEnglish(t *testing.T) {
	defer Set(ZH)
	Set(EN)
	if got := T("menu.new"); got != "NEW WORLD" {
		t.Fatalf("menu.new = %q", got)
	}
	if got := ItemName(core.ItemPlanks); got != "Oak Planks" {
		t.Fatalf("plank name = %q", got)
	}
	if Next() != ZH {
		t.Fatal("Next should return Chinese from English")
	}
}

func TestAllRunesContainsChinese(t *testing.T) {
	runes := AllRunes()
	set := make(map[rune]bool)
	for _, r := range runes {
		set[r] = true
	}
	for _, r := range "新世界中文草方块" {
		if !set[r] {
			t.Fatalf("missing rune %q", r)
		}
	}
	if len(runes) < 200 {
		t.Fatalf("only %d runes", len(runes))
	}
}

func TestAllTranslationKeysExist(t *testing.T) {
	for key, zh := range translations[ZH] {
		if strings.TrimSpace(zh) == "" {
			t.Fatalf("empty zh translation for %s", key)
		}
		if strings.TrimSpace(translations[EN][key]) == "" {
			t.Fatalf("missing en translation for %s", key)
		}
	}
}
func TestContentNamesComplete(t *testing.T) {
	for id := core.BlockID(0); id < core.BlockCount; id++ {
		if zhBlockNames[id] == "" {
			t.Fatalf("missing Chinese block name for %d", id)
		}
	}
	for id := core.ItemID(0); id < core.ItemCount; id++ {
		if zhItemNames[id] == "" {
			t.Fatalf("missing Chinese item name for %d", id)
		}
	}
}
func TestChatRunesIncludeChinese(t *testing.T) {
	rs := ChatRunes()
	if len(rs) < 20000 {
		t.Fatalf("chat rune set too small: %d", len(rs))
	}
	found := false
	for _, r := range rs {
		if r == '你' {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("chat rune set missing common Chinese characters")
	}
}
