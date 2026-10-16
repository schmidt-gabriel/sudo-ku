package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

func init() { zone.NewGlobal() }

// hit returns the zone ID under a click, or "" if none. Board clicks become
// "cell-<i>". Only zones drawn in the current mode are checked, so zones
// left over from a previous screen never match.
func (m Model) hit(msg tea.MouseClickMsg) string {
	switch m.mode {
	case modePlaying, modeTutorial:
		if m.confirmQuit {
			return firstHit(msg, listZones("quit", quitCancel+1))
		}
		if m.mode == modePlaying {
			if id := firstHit(msg, []string{"btn-save", "btn-quit"}); id != "" {
				return id
			}
		}
		if z := zone.Get("board"); z != nil && z.InBounds(msg) {
			if i, ok := cellAt(z.Pos(msg)); ok {
				return fmt.Sprintf("cell-%d", i)
			}
		}
		return ""
	case modeMenu:
		return firstHit(msg, listZones("menu", len(m.menuItems())))
	case modeDifficulty:
		return firstHit(msg, listZones("diff", len(sudoku.Difficulties)))
	case modeWon:
		return firstHit(msg, []string{"won-new", "won-menu"})
	}
	return ""
}

func listZones(prefix string, n int) []string {
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("%s-%d", prefix, i)
	}
	return ids
}

func firstHit(msg tea.MouseClickMsg, ids []string) string {
	for _, id := range ids {
		if z := zone.Get(id); z != nil && z.InBounds(msg) {
			return id
		}
	}
	return ""
}

// handleClick runs the action for a zone ID, or "" for a click outside every
// zone. Clicks do exactly what the matching keys do.
func (m Model) handleClick(id string) (Model, tea.Cmd) {
	if m.tooSmall() {
		return m, nil
	}
	m.status = ""
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	name, arg, _ := strings.Cut(id, "-")
	n, _ := strconv.Atoi(arg)
	if m.confirmQuit && name != "quit" {
		return m, nil // only the dialog's choices respond while it is open
	}
	switch name {
	case "quit":
		return m.chooseQuit(n)
	case "btn":
		switch {
		case m.mode != modePlaying:
		case arg == "save":
			m.saveNow()
		case arg == "quit":
			m.confirmQuit, m.quitIndex = true, quitSave
		}
	case "cell":
		m.cursor = n
	case "menu":
		if items := m.menuItems(); n < len(items) {
			m.menuIndex = n
			return m.chooseMenu(items[n])
		}
	case "diff":
		m.newGame(sudoku.Difficulties[n])
	case "won":
		if arg == "new" {
			m.mode, m.menuIndex = modeDifficulty, 0
		} else {
			m.toMenu()
		}
	}
	if m.mode == modeTutorial && tutorialSteps[m.tutorialStep].done == nil {
		m.tutorialStep++ // a click works like enter on "press enter" steps
	}
	m.advanceTutorial()
	return m, nil
}
