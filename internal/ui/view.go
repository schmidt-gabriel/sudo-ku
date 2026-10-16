package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

// The game screen is a header, the board, a shortcut line and a status line.
const (
	minWidth  = boardW
	minHeight = boardH + 4
)

func (m Model) View() tea.View {
	v := tea.NewView(zone.Scan(m.render()))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

// tooSmall reports whether the window is too small to draw the game.
// Width 0 means no size has arrived yet.
func (m Model) tooSmall() bool {
	return m.width > 0 && (m.width < minWidth || m.height < minHeight)
}

func (m Model) render() string {
	if m.tooSmall() {
		return fmt.Sprintf(m.t.TooSmall, minWidth, minHeight, m.width, m.height)
	}
	st := newStyles(m.dark)
	if m.showHelp {
		return st.title.Render(m.t.Help[0]) + "\n" + strings.Join(m.t.Help[1:], "\n")
	}
	if m.confirmQuit {
		opts := []string{m.t.SaveAndQuit, m.t.DiscardAndQuit, m.t.Cancel}
		return m.renderList(st, m.t.QuitTitle, opts, "quit", m.quitIndex, nil, m.t.QuitHelp)
	}
	switch m.mode {
	case modeDifficulty:
		var opts []string
		for _, d := range sudoku.Difficulties {
			opts = append(opts, m.t.Difficulties[d])
		}
		return m.renderList(st, m.t.ChooseDifficulty, opts, "diff", m.menuIndex, nil, m.t.DifficultyHelp)
	case modePlaying, modeTutorial:
		return m.renderGame(st)
	case modeWon:
		return m.renderWon(st)
	}
	return m.renderMenu(st)
}

func (m Model) renderMenu(st styles) string {
	labels := map[menuItem]string{
		itemContinue: m.t.Continue,
		itemNewGame:  m.t.NewGame,
		itemTutorial: m.t.Tutorial,
		itemLanguage: m.t.Language,
		itemQuit:     m.t.Quit,
	}
	var opts []string
	for _, item := range m.menuItems() {
		opts = append(opts, labels[item])
	}
	return m.renderList(st, m.t.Title, opts, "menu", m.menuIndex, m.recentGames(st), m.t.MenuHelp)
}

// recentGamesShown is how many finished games the menu lists.
const recentGamesShown = 5

// recentGames lists the latest finished games: date, difficulty and time.
func (m Model) recentGames(st styles) []string {
	if len(m.history) == 0 {
		return nil
	}
	lines := []string{"", st.title.Render(m.t.RecentGames)}
	for _, r := range m.history[:min(len(m.history), recentGamesShown)] {
		name := r.Difficulty.String()
		if r.Difficulty.Valid() {
			name = m.t.Difficulties[r.Difficulty]
		}
		lines = append(lines, st.help.Render(fmt.Sprintf("  %s  %-8s %s",
			r.FinishedAt.Local().Format(m.t.DateFormat), name, formatDuration(r.Elapsed))))
	}
	return lines
}

// renderList draws a title over clickable options with zone IDs
// "<prefix>-0", "<prefix>-1", and so on, then any extra lines and a help line.
func (m Model) renderList(st styles, title string, opts []string, prefix string, selected int, extra []string, help string) string {
	lines := []string{st.title.Render(title), ""}
	for n, opt := range opts {
		line := st.item.Render("  " + opt)
		if n == selected {
			line = st.itemOn.Render("> " + opt)
		}
		lines = append(lines, zone.Mark(fmt.Sprintf("%s-%d", prefix, n), line))
	}
	lines = append(lines, extra...)
	lines = append(lines, "", st.help.Render(help), st.status.Render(m.status))
	return strings.Join(lines, "\n")
}

func (m Model) renderGame(st styles) string {
	return strings.Join([]string{
		m.header(st),
		zone.Mark("board", m.renderBoard(st)),
		m.footer(st),
		st.status.Render(m.status),
	}, "\n")
}

func (m Model) header(st styles) string {
	if m.mode == modeTutorial {
		return st.title.Render(m.t.Tutorial) + st.help.Render(m.t.TutorialSkip)
	}
	modeName := m.t.ModeValue
	if m.notesMode {
		modeName = m.t.ModeNotes
	}
	left := st.title.Render(m.difficultyName()) +
		st.help.Render(" · "+formatDuration(m.game.Elapsed)+m.t.ModeLabel) +
		st.item.Render(modeName)
	buttons := zone.Mark("btn-save", st.item.Render("["+m.t.Save+"]")) + " " +
		zone.Mark("btn-menu", st.item.Render("["+m.t.Menu+"]"))
	gap := max(boardW-lipgloss.Width(left)-lipgloss.Width(buttons), 1)
	return left + strings.Repeat(" ", gap) + buttons
}

// footer shows the shortcuts, or in the tutorial the current step wrapped
// to at most two lines.
func (m Model) footer(st styles) string {
	if m.mode == modeTutorial {
		text := fmt.Sprintf("(%d/%d) %s", m.tutorialStep+1, len(tutorialSteps), m.t.TutorialSteps[m.tutorialStep])
		return st.item.Width(minWidth).Render(text)
	}
	return st.help.Render(m.t.GameHelp)
}

func (m Model) renderWon(st styles) string {
	return strings.Join([]string{
		st.title.Render(m.t.Won),
		"",
		fmt.Sprintf(m.t.WonDetail, m.difficultyName(), formatDuration(m.game.Elapsed)),
		"",
		zone.Mark("won-new", m.button(st, m.t.NewGame, m.menuIndex == 0)) + "  " +
			zone.Mark("won-menu", m.button(st, m.t.Menu, m.menuIndex == 1)),
		"",
		st.help.Render(m.t.WonHelp),
		st.status.Render(m.status),
	}, "\n")
}

// button draws a bracketed button, highlighted when selected.
func (m Model) button(st styles, label string, selected bool) string {
	if selected {
		return st.itemOn.Render("> [" + label + "]")
	}
	return st.item.Render("  [" + label + "]")
}

// difficultyName names the current game's difficulty in the player's language.
func (m Model) difficultyName() string {
	if !m.game.Difficulty.Valid() {
		return m.game.Difficulty.String()
	}
	return m.t.Difficulties[m.game.Difficulty]
}

// formatDuration shows 04:12, or 1:02:03 past an hour.
func formatDuration(d time.Duration) string {
	s := int(d.Seconds())
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}
