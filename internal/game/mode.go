package game

import (
	"fmt"
	"strings"

	"sevg/internal/core"

	"sevg/internal/i18n"
)

func (g *Game) modeName() string {
	if g.mode == ModeCreative {
		return i18n.T("chat.mode_creative")
	}
	return i18n.T("chat.mode_survival")
}

func (g *Game) setGameMode(mode GameMode) bool {
	if mode != ModeSurvival && mode != ModeCreative {
		return false
	}
	g.mode = mode
	if g.player != nil {
		g.player.Creative = mode == ModeCreative
		g.player.Flying = false
		g.player.FallingReset()
		if mode == ModeCreative {
			g.player.Health = g.player.MaxHP
		} else if g.player.Health <= 0 {
			g.player.Health = g.player.MaxHP
		}
	}
	g.breakActive = false
	g.breakProgress = 0
	g.breakCool = 0
	g.spaceTapTimer = 0
	g.heldStack = core.Stack{}
	if mode == ModeCreative {
		g.inventoryTab = 1
	} else {
		g.inventoryTab = 0
	}
	return true
}

func (g *Game) systemChat(text string) {
	g.addChat(fmt.Sprintf("[%s] %s", i18n.T("chat.system"), text))
}

func (g *Game) executeCommand(cmd string) {
	fields := strings.Fields(strings.TrimSpace(cmd))
	if len(fields) == 0 {
		g.systemChat(i18n.T("chat.command_help"))
		return
	}
	switch strings.ToLower(fields[0]) {
	case "gamemode", "gm":
		if len(fields) != 2 {
			g.systemChat(i18n.T("chat.gamemode_usage"))
			return
		}
		var mode GameMode
		switch strings.ToLower(fields[1]) {
		case "0", "survival", "s":
			mode = ModeSurvival
		case "1", "creative", "c":
			mode = ModeCreative
		default:
			g.systemChat(i18n.T("chat.gamemode_usage"))
			return
		}
		g.setGameMode(mode)
		g.systemChat(i18n.T("chat.gamemode_changed", g.modeName()))
	case "time":
		if len(fields) != 3 || strings.ToLower(fields[1]) != "set" {
			g.systemChat(i18n.T("chat.time_usage"))
			return
		}
		switch strings.ToLower(fields[2]) {
		case "day":
			g.time = 0.25
		case "noon":
			g.time = 0.5
		case "night":
			g.time = 0.75
		case "midnight":
			g.time = 0
		default:
			g.systemChat(i18n.T("chat.time_usage"))
			return
		}
		g.systemChat(i18n.T("chat.time_set", g.timeString()))
	case "help":
		g.systemChat(i18n.T("chat.command_help"))
	default:
		g.systemChat(i18n.T("chat.command_unknown", fields[0]))
	}
}
