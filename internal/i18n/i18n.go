// Package i18n contains the game's UI and content translations.
package i18n

import (
	"fmt"
	"sort"
	"strings"

	"sevg/internal/core"
	"sevg/internal/world"
)

const (
	ZH = "zh-CN"
	EN = "en-US"
)

var current = ZH

var translations = map[string]map[string]string{
	ZH: {
		"window.title":          "SEVG - 简单且可扩展的体素游戏",
		"chat.welcome":          "欢迎来到 SEVG！按 T 打开聊天",
		"chat.mode_survival":    "生存模式",
		"chat.mode_creative":    "创造模式",
		"chat.system":           "系统",
		"chat.gamemode_changed": "游戏模式：%s",
		"chat.gamemode_usage":   "用法：/gamemode 0|1（0 生存，1 创造）",
		"chat.command_help":     "指令：/gamemode 0|1",
		"chat.command_unknown":  "未知指令：%s",
		"chat.you":              "你",
		"chat.placeholder":      "输入消息…",
		"chat.hint":             "Enter 发送 · Esc 关闭 · 滚轮翻阅",
		"common.health":         "生命",
		"common.inventory":      "背包",
		"common.crafting":       "合成",
		"common.continue":       "继续",
		"common.close":          "关闭",
		"common.save":           "保存",
		"common.quit":           "退出",
		"common.on":             "开",
		"common.off":            "关",
		"common.none":           "-",
		"common.unknown":        "未知",

		"menu.subtitle":        "简单且可扩展的体素游戏",
		"menu.tagline":         "Go + SDL3 GPU · 程序化世界 · 区块 · 实体 · 物品 · GUI",
		"menu.new":             "新世界",
		"menu.continue":        "继续游戏",
		"menu.quit":            "退出游戏",
		"menu.enter_hint":      "ENTER：新世界    ESC：退出    L：切换语言",
		"menu.controls":        "操作说明",
		"menu.ctrl_move":       "WASD ........ 移动        SPACE ... 跳跃",
		"menu.ctrl_sprint":     "SHIFT 疾跑    SPACE x2 飞行",
		"menu.ctrl_mine":       "左键 ........ 挖掘/攻击   右键 .... 放置",
		"menu.ctrl_pick":       "中键 ........ 选取方块    1-9 ..... 快捷栏",
		"menu.ctrl_inv":        "E ........... 背包        F3 ...... 调试",
		"menu.ctrl_pause":      "ESC ......... 暂停        F5 ...... 保存",
		"menu.render_distance": "渲染距离 %d",
		"menu.language":        "语言：%s",
		"menu.language_hint":   "L 切换语言",

		"pause.title":        "已暂停",
		"pause.resume":       "继续游戏",
		"pause.save":         "保存世界",
		"pause.title_screen": "返回标题",
		"pause.quit":         "退出游戏",
		"pause.hint":         "ESC：继续    L：切换语言",

		"inventory.title":    "背包",
		"inventory.hint":     "左键移动 - 右键拆分 - E 关闭",
		"inventory.crafting": "合成",
		"inventory.handmade": "手工合成 - 无需工作台",
		"inventory.tip1":     "提示：原木 -> 木板 -> 木棍",
		"inventory.tip2":     "开采煤炭和铁矿石来制作工具",

		"death.title":        "你死了",
		"death.subtitle":     "世界不会怜悯你，但你的建筑还在。",
		"death.respawn":      "重生",
		"death.title_screen": "返回标题",

		"hud.health":      "生命",
		"hud.chunks":      "区块 %d/%d",
		"hud.generating":  "正在生成地形",
		"hud.quick_start": "SEVG 快速上手",
		"hud.help1":       "WASD 移动 - SPACE 跳跃 - SHIFT 疾跑",
		"hud.help2":       "左键 挖掘/攻击 - 右键 放置/进食",
		"hud.help3":       "中键 选取方块 - 滚轮 切换 - 1-9 快捷栏",
		"hud.help4":       "E 背包 - T 聊天 - / 指令 - F3 调试",
		"hud.help5":       "破坏方块收集物品，开始建造吧！",
		"hud.help6":       "H 隐藏帮助",
		"hud.help7":       "创造模式：双击 SPACE 飞行",
		"hud.debug_title": "SEVG 调试 (F3)",
		"hud.fps":         "帧率 %.0f（%.1f ms）",
		"hud.pos":         "坐标 %.1f %.1f %.1f",
		"hud.chunk":       "区块 %d %d",
		"hud.biome":       "群系 %s",
		"hud.seed":        "种子 %d",
		"hud.time":        "时间 %s",
		"hud.chunk_stats": "区块 %d 已网格化 %d",
		"hud.mob_stats":   "生物 %d 掉落物 %d",
		"hud.target":      "目标 %s",
		"hud.mode":        "模式：%s",
		"hud.fly":         "飞行 %s",
		"hud.target_none": "-",
		"hud.target_pos":  "%d %d %d",

		"toast.saved":               "世界已保存",
		"toast.autosaved":           "已自动保存",
		"toast.save_failed":         "保存失败：%s",
		"toast.loaded":              "世界已加载",
		"toast.new_world":           "新世界 - 种子 %d",
		"toast.could_not_load_save": "无法读取存档：%s",
		"toast.could_not_load":      "无法加载：%s",
		"toast.could_not_start":     "无法开始：%s",
		"toast.flying":              "飞行：%s",
		"toast.respawned":           "已重生",
		"toast.full_health":         "生命值已满",
		"toast.ate":                 "吃掉了 %s",
		"toast.no_room":             "空间不足",
		"toast.no_item":             "背包里没有 %s",
		"toast.missing":             "缺少 %s",
		"toast.crafted":             "合成了 %s",
		"toast.language":            "语言已切换为 %s",
	},
	EN: {
		"window.title":          "SEVG - Simple and Extendable Voxel Game",
		"chat.welcome":          "Welcome to SEVG! Press T to chat",
		"chat.mode_survival":    "Survival",
		"chat.mode_creative":    "Creative",
		"chat.system":           "SYSTEM",
		"chat.gamemode_changed": "Game mode: %s",
		"chat.gamemode_usage":   "Usage: /gamemode 0|1 (0 survival, 1 creative)",
		"chat.command_help":     "Commands: /gamemode 0|1",
		"chat.command_unknown":  "Unknown command: %s",
		"chat.you":              "You",
		"chat.placeholder":      "Type a message...",
		"chat.hint":             "Enter send · Esc close · Wheel scroll",
		"common.health":         "HEALTH",
		"common.inventory":      "INVENTORY",
		"common.crafting":       "CRAFTING",
		"common.continue":       "CONTINUE",
		"common.close":          "CLOSE",
		"common.save":           "SAVE",
		"common.quit":           "QUIT",
		"common.on":             "ON",
		"common.off":            "OFF",
		"common.none":           "-",
		"common.unknown":        "UNKNOWN",

		"menu.subtitle":        "SIMPLE AND EXTENDABLE VOXEL GAME",
		"menu.tagline":         "GO + SDL3 GPU · PROCEDURAL WORLD · CHUNKS · ENTITIES · ITEMS · GUI",
		"menu.new":             "NEW WORLD",
		"menu.continue":        "CONTINUE",
		"menu.quit":            "QUIT",
		"menu.enter_hint":      "ENTER: NEW WORLD    ESC: QUIT    L: LANGUAGE",
		"menu.controls":        "CONTROLS",
		"menu.ctrl_move":       "WASD ........ MOVE        SPACE ... JUMP",
		"menu.ctrl_sprint":     "SHIFT SPRINT    SPACE x2 FLY",
		"menu.ctrl_mine":       "LMB ......... MINE/ATTACK RMB ..... PLACE",
		"menu.ctrl_pick":       "MMB ......... PICK BLOCK  1-9 ..... HOTBAR",
		"menu.ctrl_inv":        "E ........... INVENTORY   F3 ...... DEBUG",
		"menu.ctrl_pause":      "ESC ......... PAUSE       F5 ...... SAVE",
		"menu.render_distance": "RENDER DISTANCE %d",
		"menu.language":        "LANGUAGE: %s",
		"menu.language_hint":   "L: SWITCH LANGUAGE",

		"pause.title":        "PAUSED",
		"pause.resume":       "RESUME",
		"pause.save":         "SAVE WORLD",
		"pause.title_screen": "QUIT TO TITLE",
		"pause.quit":         "QUIT GAME",
		"pause.hint":         "ESC: RESUME    L: LANGUAGE",

		"inventory.title":    "INVENTORY",
		"inventory.hint":     "CLICK TO MOVE - RIGHT CLICK TO SPLIT - E TO CLOSE",
		"inventory.crafting": "CRAFTING",
		"inventory.handmade": "HANDMADE - NO TABLE NEEDED",
		"inventory.tip1":     "TIPS: LOG -> PLANKS -> STICKS",
		"inventory.tip2":     "MINE COAL AND IRON FOR TOOLS",

		"death.title":        "YOU DIED",
		"death.subtitle":     "THE WORLD IS UNFORGIVING. YOUR BUILDINGS ARE STILL THERE.",
		"death.respawn":      "RESPAWN",
		"death.title_screen": "TITLE SCREEN",

		"hud.health":      "HEALTH",
		"hud.chunks":      "CHUNKS %d/%d",
		"hud.generating":  "GENERATING TERRAIN",
		"hud.quick_start": "SEVG QUICK START",
		"hud.help1":       "WASD MOVE - SPACE JUMP - SHIFT SPRINT",
		"hud.help2":       "LMB MINE / ATTACK - RMB PLACE / EAT",
		"hud.help3":       "MMB PICK BLOCK - WHEEL SELECT - 1-9 HOTBAR",
		"hud.help4":       "E INVENTORY - T CHAT - / COMMANDS - F3 DEBUG",
		"hud.help5":       "BREAK BLOCKS TO COLLECT ITEMS AND BUILD!",
		"hud.help6":       "H TO HIDE HELP",
		"hud.help7":       "CREATIVE: DOUBLE-TAP SPACE TO FLY",
		"hud.debug_title": "SEVG DEBUG (F3)",
		"hud.fps":         "FPS %.0f (%.1f ms)",
		"hud.pos":         "POS %.1f %.1f %.1f",
		"hud.chunk":       "CHUNK %d %d",
		"hud.biome":       "BIOME %s",
		"hud.seed":        "SEED %d",
		"hud.time":        "TIME %s",
		"hud.chunk_stats": "CHUNKS %d  MESHED %d",
		"hud.mob_stats":   "MOBS %d  DROPS %d",
		"hud.target":      "TARGET %s",
		"hud.mode":        "MODE: %s",
		"hud.fly":         "FLY %s",
		"hud.target_none": "-",
		"hud.target_pos":  "%d %d %d",

		"toast.saved":               "WORLD SAVED",
		"toast.autosaved":           "AUTO-SAVED",
		"toast.save_failed":         "SAVE FAILED: %s",
		"toast.loaded":              "WORLD LOADED",
		"toast.new_world":           "NEW WORLD - SEED %d",
		"toast.could_not_load_save": "COULD NOT LOAD SAVE: %s",
		"toast.could_not_load":      "COULD NOT LOAD: %s",
		"toast.could_not_start":     "COULD NOT START: %s",
		"toast.flying":              "FLYING %s",
		"toast.respawned":           "RESPAWNED",
		"toast.full_health":         "ALREADY AT FULL HEALTH",
		"toast.ate":                 "ATE %s",
		"toast.no_room":             "NOT ENOUGH ROOM",
		"toast.no_item":             "NO %s IN INVENTORY",
		"toast.missing":             "MISSING %s",
		"toast.crafted":             "CRAFTED %s",
		"toast.language":            "LANGUAGE: %s",
	},
}

var zhBlockNames = map[core.BlockID]string{
	core.Air: "空气", core.Grass: "草方块", core.Dirt: "泥土", core.Stone: "石头",
	core.Cobble: "圆石", core.Sand: "沙子", core.Water: "水", core.Log: "橡木原木",
	core.Leaves: "树叶", core.Planks: "橡木木板", core.CoalOre: "煤矿石", core.IronOre: "铁矿石",
	core.Snow: "雪", core.SnowyGrass: "积雪草方块", core.Cactus: "仙人掌", core.Glass: "玻璃",
	core.Bedrock: "基岩", core.Gravel: "沙砾", core.Flower: "花", core.TallGrass: "高草",
	core.Brick: "砖块", core.StoneBrick: "石砖", core.Ice: "冰", core.Obsidian: "黑曜石",
	core.Wool: "羊毛",
}

var zhItemNames = map[core.ItemID]string{
	core.ItemNone: "空", core.ItemGrass: "草方块", core.ItemDirt: "泥土", core.ItemStone: "石头",
	core.ItemCobble: "圆石", core.ItemSand: "沙子", core.ItemLog: "橡木原木", core.ItemLeaves: "树叶",
	core.ItemPlanks: "橡木木板", core.ItemCoalOre: "煤矿石", core.ItemIronOre: "铁矿石",
	core.ItemSnow: "雪", core.ItemCactus: "仙人掌", core.ItemGlass: "玻璃", core.ItemGravel: "沙砾",
	core.ItemFlower: "花", core.ItemTallGrass: "高草", core.ItemBrick: "砖块",
	core.ItemStoneBrick: "石砖", core.ItemIce: "冰", core.ItemObsidian: "黑曜石",
	core.ItemStick: "木棍", core.ItemApple: "苹果", core.ItemCoal: "煤炭",
	core.ItemIronIngot: "铁锭", core.ItemWoodSword: "木剑", core.ItemStonePickaxe: "石镐",
	core.ItemIronSword: "铁剑", core.ItemIronPickaxe: "铁镐", core.ItemWool: "羊毛",
}

var zhBiomeNames = map[world.Biome]string{
	world.BiomePlains: "平原", world.BiomeForest: "森林", world.BiomeDesert: "沙漠",
	world.BiomeSnow: "雪原", world.BiomeMountains: "山地", world.BiomeOcean: "海洋",
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

func BlockName(b core.BlockID) string {
	if current == ZH {
		if s, ok := zhBlockNames[b]; ok {
			return s
		}
	}
	return b.Name()
}

func ItemName(i core.ItemID) string {
	if current == ZH {
		if s, ok := zhItemNames[i]; ok {
			return s
		}
	}
	return i.Name()
}

func BiomeName(b world.Biome) string {
	if current == ZH {
		if s, ok := zhBiomeNames[b]; ok {
			return s
		}
	}
	return b.Name()
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
	for _, s := range zhBlockNames {
		for _, r := range s {
			set[r] = true
		}
	}
	for _, s := range zhItemNames {
		for _, r := range s {
			set[r] = true
		}
	}
	for _, s := range zhBiomeNames {
		for _, r := range s {
			set[r] = true
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
