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

var padZones = func() []string {
	ids := []string{"btn-erase", "btn-notes", "btn-undo"}
	for d := 1; d <= 9; d++ {
		ids = append(ids, fmt.Sprintf("num-%d", d))
	}
	return ids
}()

// hit returns the zone ID under a click, or "" if none. Board clicks become
// "cell-<i>". Only zones drawn in the current mode are checked, so zones
// left over from a previous screen never match.
func (m Model) hit(msg tea.MouseClickMsg) string {
	switch m.mode {
	case modePlaying, modeTutorial:
		if z := zone.Get("board"); z != nil && z.InBounds(msg) {
			if i, ok := cellAt(z.Pos(msg)); ok {
				return fmt.Sprintf("cell-%d", i)
			}
			return ""
		}
		return firstHit(msg, padZones)
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
	switch name {
	case "cell":
		m.cursor = n
	case "num":
		m.input(uint8(n))
	case "btn":
		switch arg {
		case "erase":
			m.erase()
		case "notes":
			m.notesMode = !m.notesMode
		case "undo":
			m.undo()
		}
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
