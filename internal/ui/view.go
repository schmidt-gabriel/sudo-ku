package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

// The game screen is a header, the board, the number pad, a shortcut line
// and a status line.
const (
	minWidth  = 60
	minHeight = boardH + 5
)

func (m Model) View() tea.View {
	v := tea.NewView(zone.Scan(m.render()))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) render() string {
	if m.width > 0 && (m.width < minWidth || m.height < minHeight) {
		return fmt.Sprintf(textTooSmall, minWidth, minHeight, m.width, m.height)
	}
	st := newStyles(m.dark)
	if m.showHelp {
		return st.title.Render(helpLines[0]) + "\n" + strings.Join(helpLines[1:], "\n")
	}
	switch m.mode {
	case modeDifficulty:
		var opts []string
		for _, d := range sudoku.Difficulties {
			opts = append(opts, d.String())
		}
		return m.renderList(st, textChooseDifficulty, opts, "diff")
	case modePlaying, modeTutorial:
		return m.renderGame(st)
	case modeWon:
		return m.renderWon(st)
	}
	return m.renderMenu(st)
}

func (m Model) renderMenu(st styles) string {
	labels := map[menuItem]string{
		itemContinue: textContinue,
		itemNewGame:  textNewGame,
		itemTutorial: textTutorial,
		itemQuit:     textQuit,
	}
	var opts []string
	for _, item := range m.menuItems() {
		opts = append(opts, labels[item])
	}
	return m.renderList(st, textTitle, opts, "menu")
}

// renderList draws a title over clickable options with zone IDs
// "<prefix>-0", "<prefix>-1", and so on.
func (m Model) renderList(st styles, title string, opts []string, prefix string) string {
	lines := []string{st.title.Render(title), ""}
	for n, opt := range opts {
		line := st.item.Render("  " + opt)
		if n == m.menuIndex {
			line = st.itemOn.Render("> " + opt)
		}
		lines = append(lines, zone.Mark(fmt.Sprintf("%s-%d", prefix, n), line))
	}
	lines = append(lines, "", st.help.Render(textMenuHelp), st.status.Render(m.status))
	return strings.Join(lines, "\n")
}

func (m Model) renderGame(st styles) string {
	return strings.Join([]string{
		m.header(st),
		zone.Mark("board", m.renderBoard(st)),
		m.renderPad(st),
		m.footer(st),
		st.status.Render(m.status),
	}, "\n")
}

func (m Model) header(st styles) string {
	if m.mode == modeTutorial {
		return st.title.Render(textTutorial) + st.help.Render(textTutorialSkip)
	}
	modeName := textModeValue
	if m.notesMode {
		modeName = textModeNotes
	}
	return st.title.Render(m.game.Difficulty.String()) +
		st.help.Render(" · "+formatDuration(m.game.Elapsed)+textModeLabel) +
		st.item.Render(modeName)
}

// footer shows the shortcuts, or in the tutorial the current step wrapped
// to at most two lines.
func (m Model) footer(st styles) string {
	if m.mode == modeTutorial {
		text := fmt.Sprintf("(%d/%d) %s", m.tutorialStep+1, len(tutorialSteps), tutorialSteps[m.tutorialStep].text)
		return st.item.Width(minWidth).Render(text)
	}
	return st.help.Render(textGameHelp)
}

func (m Model) renderPad(st styles) string {
	var parts []string
	for d := 1; d <= 9; d++ {
		parts = append(parts, zone.Mark(fmt.Sprintf("num-%d", d), st.item.Render(fmt.Sprintf("[%d]", d))))
	}
	notes := st.item
	if m.notesMode {
		notes = st.itemOn
	}
	parts = append(parts,
		zone.Mark("btn-erase", st.item.Render("["+textErase+"]")),
		zone.Mark("btn-notes", notes.Render("["+textNotes+"]")),
		zone.Mark("btn-undo", st.item.Render("["+textUndo+"]")),
	)
	return strings.Join(parts, " ")
}

func (m Model) renderWon(st styles) string {
	return strings.Join([]string{
		st.title.Render(textWon),
		"",
		fmt.Sprintf(textWonDetail, m.game.Difficulty, formatDuration(m.game.Elapsed)),
		"",
		zone.Mark("won-new", st.itemOn.Render("["+textNewGame+"]")) + "  " +
			zone.Mark("won-menu", st.item.Render("["+textMenu+"]")),
		"",
		st.help.Render(textWonHelp),
		st.status.Render(m.status),
	}, "\n")
}

// formatDuration shows 04:12, or 1:02:03 past an hour.
func formatDuration(d time.Duration) string {
	s := int(d.Seconds())
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}
