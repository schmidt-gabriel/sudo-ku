package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

func TestCellAt(t *testing.T) {
	cases := []struct {
		x, y int
		cell int
		ok   bool
	}{
		{1, 1, 0, true},   // top-left cell
		{9, 2, 0, true},   // bottom-right character of cell 0
		{11, 1, 1, true},  // after the thin vertical line
		{31, 1, 3, true},  // first cell of the second box
		{1, 4, 9, true},   // second row, below a thin line
		{1, 10, 27, true}, // first row of the second band
		{89, 26, 80, true},
		{0, 1, 0, false},  // left border
		{10, 1, 0, false}, // thin vertical line
		{30, 1, 0, false}, // thick border between boxes
		{1, 3, 0, false},  // thin line between rows
		{1, 9, 0, false},  // thick border between bands
		{90, 1, 0, false}, // right border
		{1, 27, 0, false}, // bottom border
		{-1, -1, 0, false},
	}
	for _, c := range cases {
		cell, ok := cellAt(c.x, c.y)
		if ok != c.ok || (ok && cell != c.cell) {
			t.Errorf("cellAt(%d,%d) = %d,%v; want %d,%v", c.x, c.y, cell, ok, c.cell, c.ok)
		}
	}
}

func TestBoardSize(t *testing.T) {
	m := playing(tutorialSeen())
	m.game.ToggleNote(2, 1)
	lines := strings.Split(m.renderBoard(newStyles(true)), "\n")
	if len(lines) != boardH {
		t.Fatalf("board has %d lines, want %d", len(lines), boardH)
	}
	for n, line := range lines {
		if w := lipgloss.Width(line); w != boardW {
			t.Fatalf("line %d is %d wide, want %d", n, w, boardW)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0:                     "00:00",
		65 * time.Second:      "01:05",
		3723 * time.Second:    "1:02:03",
		59*time.Minute + 59e9: "59:59",
	} {
		if got := formatDuration(d); got != want {
			t.Errorf("formatDuration(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestRenderEveryScreen(t *testing.T) {
	m := playing(tutorialSeen())
	m.width, m.height = 100, 50
	screens := map[string]func(Model) Model{
		"New game":    func(m Model) Model { m.toMenu(); return m },
		"Expert":      func(m Model) Model { m.mode = modeDifficulty; return m },
		"Mode: ":      func(m Model) Model { return m },
		"Solved!":     func(m Model) Model { m.mode = modeWon; return m },
		"How to play": func(m Model) Model { m.showHelp = true; return m },
		"Make the terminal": func(m Model) Model {
			m.width, m.height = 40, 20
			return m
		},
	}
	for want, setup := range screens {
		if got := setup(m).View().Content; !strings.Contains(got, want) {
			t.Errorf("screen should contain %q:\n%s", want, got)
		}
	}
}

func TestResizeBackRestoresGame(t *testing.T) {
	m := press(playing(tutorialSeen()), "4")
	next, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	m = next.(Model)
	if !strings.Contains(m.View().Content, "Make the terminal") {
		t.Fatal("small window did not show the size message")
	}
	next, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = next.(Model)
	if !strings.Contains(m.View().Content, "Mode: ") || m.game.Values[2] != 4 {
		t.Error("growing the window did not bring the game back intact")
	}
}

func TestClicks(t *testing.T) {
	fs := tutorialSeen()
	m := playing(fs)
	click := func(id string) {
		t.Helper()
		m, _ = m.handleClick(id)
	}
	click("cell-3")
	if m.cursor != 3 {
		t.Fatalf("cursor = %d, want 3", m.cursor)
	}
	m = press(m, "6")
	m.toMenu()
	click("menu-0") // Continue
	if m.mode != modePlaying {
		t.Fatalf("menu click: mode %v, want playing", m.mode)
	}
	m.mode = modeDifficulty
	click("diff-2")
	if m.mode != modePlaying || m.game.Difficulty != sudoku.Hard {
		t.Fatal("difficulty click did not start a Hard game")
	}
	m.mode = modeWon
	click("won-menu")
	if m.mode != modeMenu {
		t.Fatalf("won-menu: mode %v, want menu", m.mode)
	}
	m.showHelp = true
	click("menu-0")
	if m.showHelp || m.mode != modeMenu {
		t.Error("a click while help is open should only close help")
	}
}

func TestNotesRenderInCell(t *testing.T) {
	m := playing(tutorialSeen())
	m.game = game.FromPuzzle(testPuzzle, testSolution, sudoku.Easy)
	m.game.ToggleNote(2, 1)
	m.game.ToggleNote(2, 9)
	if got := m.cellLine(2, 0); got != "1        " {
		t.Errorf("notes line 0 = %q, want %q", got, "1        ")
	}
	if got := m.cellLine(2, 1); got != "       9 " {
		t.Errorf("notes line 1 = %q, want %q", got, "       9 ")
	}
	m.game.ToggleNote(3, 5)
	m.game.ToggleNote(3, 6)
	if got := m.cellLine(3, 0) + "|" + m.cellLine(3, 1); got != "        5| 6       " {
		t.Errorf("notes 5 and 6 = %q, want %q", got, "        5| 6       ")
	}
	if got := m.cellLine(0, 0) + "|" + m.cellLine(0, 1); got != "    5    |         " {
		t.Errorf("value lines = %q, want %q", got, "    5    |         ")
	}
}

func TestTooSmallFreezesTheGame(t *testing.T) {
	m := playing(tutorialSeen())
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	m = press(m, "4", "x", "u")
	m, _ = m.handleClick("cell-3")
	if m.cursor != 2 {
		t.Errorf("a click on the hidden board moved the cursor to %d", m.cursor)
	}
	next, _ = m.Update(tickMsg{})
	m = next.(Model)
	if m.game.Values[2] != 0 || m.game.Elapsed != 0 {
		t.Errorf("hidden board changed: cell 2 = %d, elapsed %v", m.game.Values[2], m.game.Elapsed)
	}
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Fatal("q did nothing while the window was too small")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("q should still quit while the window is too small")
	}
}

func TestClickAnywhereClosesHelp(t *testing.T) {
	m := playing(tutorialSeen())
	m.showHelp = true
	next, _ := m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if next.(Model).showHelp {
		t.Error("a click outside any zone did not close help")
	}
}

func TestBoardGeometry(t *testing.T) {
	if cellW != 9 || cellH != 2 || boardW != 91 || boardH != 28 {
		t.Errorf("cells %dx%d board %dx%d, want 9x2 and 91x28", cellW, cellH, boardW, boardH)
	}
	if minWidth > 91 || minHeight > 33 {
		t.Errorf("minimum terminal %dx%d, want it to fit 91x33", minWidth, minHeight)
	}
	if minWidth < boardW {
		t.Errorf("minWidth %d is narrower than the board (%d)", minWidth, boardW)
	}
}

func TestHighlightHierarchy(t *testing.T) {
	st := newStyles(true)
	m := playing(tutorialSeen())
	m.cursor = 0 // a given 5

	if bg := m.cellBg(st, 0); bg != st.selectedBg {
		t.Errorf("selected cell background = %v, want selectedBg", bg)
	}
	for _, peer := range []int{1, 9, 10} { // row, column, box
		if bg := m.cellBg(st, peer); bg != st.peerBg {
			t.Errorf("peer %d background = %v, want peerBg", peer, bg)
		}
	}
	// Cell 14 holds another 5 but is not a peer of cell 0: text only, no background.
	if bg := m.cellBg(st, 14); bg != nil {
		t.Errorf("same-digit cell has background %v, want none", bg)
	}
	if s := m.cellStyle(st, 14, false); s.GetForeground() != st.text || !s.GetUnderline() || !s.GetBold() {
		t.Error("a clue with the selected digit should stay white, underlined and bold")
	}
	m.game.Set(40, 5) // a player's 5, not a peer of cell 0
	if s := m.cellStyle(st, 40, false); s.GetForeground() != st.accent || !s.GetUnderline() {
		t.Error("a player's digit equal to the selected one should stay blue and be underlined")
	}
	if bg := m.cellBg(st, 40); bg != nil {
		t.Errorf("unrelated cell has background %v, want none", bg)
	}
}

func TestGridLinesAreNeverPainted(t *testing.T) {
	st := newStyles(true)
	m := playing(tutorialSeen())
	m.cursor = 10 // highlights row 1, column 1 and box 0, which touch many lines
	for row := range 8 {
		if line := m.thinRule(st, row); strings.Contains(line, "48;") {
			t.Errorf("thin line under row %d has a background: %q", row, line)
		}
	}
	thin := st.border.Render("│")
	for row := range 9 {
		if n := strings.Count(m.cellRow(st, row, 0, nil), thin); n != 6 {
			t.Errorf("row %d has %d plain thin separators, want 6", row, n)
		}
	}
}

func TestSelectedCellKeepsItsColors(t *testing.T) {
	for _, dark := range []bool{true, false} {
		st := newStyles(dark)
		m := playing(tutorialSeen()) // cursor on empty cell 2
		fg := func() any { return m.cellStyle(st, m.cursor, false).GetForeground() }
		m.game.ToggleNote(2, 1)
		if got := fg(); got != st.text {
			t.Errorf("dark=%v: notes on the selected cell = %v, want the main text color", dark, got)
		}
		m.game.Set(2, 4)
		if got := fg(); got != st.accent {
			t.Errorf("dark=%v: a player's digit on the selected cell = %v, want blue", dark, got)
		}
		m.cursor = 0
		if got := fg(); got != st.text {
			t.Errorf("dark=%v: a given on the selected cell = %v, want the main text color", dark, got)
		}
		if got := m.cellStyle(st, 0, true).GetForeground(); got != st.conflict {
			t.Errorf("dark=%v: a conflict on the selected cell should stay red", dark)
		}
		if st.selectedBg == st.peerBg || st.selectedBg == st.accent {
			t.Errorf("dark=%v: the selection must differ from peers and from the player-digit blue", dark)
		}
		m.cursor = 2
		if got := m.cellStyle(st, 3, false).GetForeground(); got != st.muted {
			t.Errorf("dark=%v: notes elsewhere should stay muted", dark)
		}
	}
}

func TestBoxBordersStandOut(t *testing.T) {
	for _, dark := range []bool{true, false} {
		st := newStyles(dark)
		if st.frame.GetForeground() == st.border.GetForeground() {
			t.Errorf("dark=%v: box borders use the same color as the thin lines", dark)
		}
		if !st.frame.GetBold() {
			t.Errorf("dark=%v: box borders should be bold", dark)
		}
	}
	m := playing(tutorialSeen())
	st := newStyles(true)
	if got := thickRule(st, "┏", "┯", "┳", "┓"); got != st.frame.Render(ansi.Strip(got)) {
		t.Error("thick rules are not drawn with the frame style")
	}
	if row := m.cellRow(st, 0, 0, nil); !strings.HasPrefix(row, st.frame.Render("┃")) {
		t.Error("the left box border is not drawn with the frame style")
	}
}

func TestNoNumberPad(t *testing.T) {
	m := playing(tutorialSeen())
	m.width, m.height = 100, 50
	for _, tx := range languages {
		m.t = tx
		got := ansi.Strip(m.View().Content)
		if strings.Contains(got, "[1]") || strings.Contains(got, "[9]") {
			t.Errorf("game screen still shows a number pad:\n%s", got)
		}
	}
	if minHeight != boardH+4 {
		t.Errorf("minHeight = %d, want board plus header, shortcuts and status (%d)", minHeight, boardH+4)
	}
}

func TestUnderlineCoversOnlyTheDigit(t *testing.T) {
	st := newStyles(true)
	m := playing(tutorialSeen())
	m.cursor = 0 // a clue 5; cell 14 holds another 5
	style := m.cellStyle(st, 14, false)
	got := m.renderCell(st, 14, 0, false)
	if ansi.Strip(got) != m.cellLine(14, 0) {
		t.Fatalf("renderCell text = %q, want %q", ansi.Strip(got), m.cellLine(14, 0))
	}
	if !strings.Contains(got, style.Render("5")) || got == style.Render(m.cellLine(14, 0)) {
		t.Errorf("only the digit should be underlined: %q", got)
	}
	if line, want := m.renderCell(st, 14, 1, false), style.Underline(false).Render(m.cellLine(14, 1)); line != want {
		t.Errorf("the empty second line should not be underlined: %q", line)
	}
}
