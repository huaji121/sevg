// Package i18n contains the game's UI and content translations.
package i18n

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"sevg/assets"
	"sevg/internal/core"
	"sevg/internal/world"
)

const (
	ZH = "zh-CN"
	EN = "en-US"
)

var current = ZH

// translations is loaded from assets/lang/*.json at startup.
var translations map[string]map[string]string

func init() {
	translations = make(map[string]map[string]string, 2)
	for _, lang := range []string{ZH, EN} {
		path := "lang/" + lang + ".json"
		data, err := assets.Lang.ReadFile(path)
		if err != nil {
			panic(fmt.Sprintf("read %s: %v", path, err))
		}
		var table map[string]string
		if err := json.Unmarshal(data, &table); err != nil {
			panic(fmt.Sprintf("parse %s: %v", path, err))
		}
		translations[lang] = table
	}
}

func Set(lang string) bool {
	normalized := normalize(lang)
	if _, ok := translations[normalized]; !ok {
		return false
	}
	current = normalized
	return true
}

func normalize(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "zh", "zh-cn", "cn", "chinese", "中文":
		return ZH
	case "en", "en-us", "english":
		return EN
	default:
		return lang
	}
}

func Current() string { return current }
func IsChinese() bool { return current == ZH }
func Next() string {
	if current == ZH {
		return EN
	}
	return ZH
}
func LocaleName(l string) string {
	if normalize(l) == ZH {
		return "中文"
	}
	return "English"
}

func T(key string, args ...any) string {
	s := translations[current][key]
	if s == "" {
		s = translations[EN][key]
	}
	if s == "" {
		s = key
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

func keySlug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == ' ' || r == '-' || r == '_':
			b.WriteByte('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func lookupName(key, fallback string) string {
	if s := translations[current][key]; s != "" {
		return s
	}
	if s := translations[EN][key]; s != "" {
		return s
	}
	return fallback
}

func BlockName(b core.BlockID) string {
	return lookupName("block."+keySlug(b.Name()), b.Name())
}

func ItemName(i core.ItemID) string {
	return lookupName("item."+keySlug(i.Name()), i.Name())
}

func BiomeName(b world.Biome) string {
	return lookupName("biome."+keySlug(b.Name()), b.Name())
}

// AllRunes returns every rune used by any translation, for font atlas packing.
func AllRunes() []rune {
	set := make(map[rune]bool)
	for r := rune(32); r < 127; r++ {
		set[r] = true
	}
	for _, lang := range translations {
		for _, s := range lang {
			for _, r := range s {
				set[r] = true
			}
		}
	}
	for _, r := range "中文英语语言：，。！？（）“”‘’、·" {
		set[r] = true
	}
	out := make([]rune, 0, len(set))
	for r := range set {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// ChatRunes returns a broad set of characters that can be entered into the
// chat. They are baked into the UI atlas so arbitrary Chinese chat text does
// not render as question marks.
func ChatRunes() []rune {
	out := make([]rune, 0, 0x9FFF-0x4E00+0x100)
	for r := rune(0x4E00); r <= 0x9FFF; r++ {
		out = append(out, r)
	}
	for r := rune(0x3000); r <= 0x303F; r++ {
		out = append(out, r)
	}
	for r := rune(0xFF01); r <= 0xFF5E; r++ {
		out = append(out, r)
	}
	return out
}
