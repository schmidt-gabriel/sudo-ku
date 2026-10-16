package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

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
		{7, 3, 0, true},   // bottom-right character of cell 0
		{9, 1, 1, true},   // after the thin separator
		{25, 1, 3, true},  // first cell of the second box
		{1, 5, 9, true},   // second row, below a thin line
		{1, 13, 27, true}, // first row of the second band
		{71, 35, 80, true},
		{1, 4, 0, false},  // thin line between rows
		{0, 1, 0, false},  // left border
		{8, 1, 0, false},  // thin separator
		{24, 1, 0, false}, // thick border between boxes
		{1, 12, 0, false}, // thick border between bands
		{72, 1, 0, false}, // right border
		{1, 36, 0, false}, // bottom border
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
	click("num-6")
	if m.game.Values[3] != 6 {
		t.Fatal("number pad did not place 6")
	}
	click("btn-erase")
	if m.game.Values[3] != 0 {
		t.Fatal("erase button did not erase")
	}
	click("btn-undo")
	if m.game.Values[3] != 6 {
		t.Fatal("undo button did not undo")
	}
	click("btn-notes")
	if !m.notesMode {
		t.Fatal("notes button did not toggle notes mode")
	}

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
	if got := m.cellLine(2, 0); got != " 1     " {
		t.Errorf("notes line 0 = %q, want %q", got, " 1     ")
	}
	if got := m.cellLine(2, 2); got != "     9 " {
		t.Errorf("notes line 2 = %q, want %q", got, "     9 ")
	}
	if got := m.cellLine(0, 1); got != "   5   " {
		t.Errorf("value line = %q, want %q", got, "   5   ")
	}
}

func TestTooSmallFreezesTheGame(t *testing.T) {
	m := playing(tutorialSeen())
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = next.(Model)
	m = press(m, "4", "x", "u")
	m, _ = m.handleClick("num-7")
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
	if cellW != 7 || cellH != 3 || boardW != 73 || boardH != 37 {
		t.Errorf("cells %dx%d board %dx%d, want 7x3 and 73x37", cellW, cellH, boardW, boardH)
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
	if s := m.cellStyle(st, 14, false); s.GetForeground() != st.same || !s.GetBold() {
		t.Error("same-digit cell should be bold in the same-digit color")
	}
	if bg := m.cellBg(st, 40); bg != nil {
		t.Errorf("unrelated cell has background %v, want none", bg)
	}
}

func TestSeparatorsJoinHighlights(t *testing.T) {
	st := newStyles(true)
	m := playing(tutorialSeen())
	m.cursor = 0
	if bg := m.separatorBg(st, 0, 1); bg != st.peerBg {
		t.Errorf("separator inside the selected row = %v, want peerBg", bg)
	}
	if bg := m.separatorBg(st, 9, 10); bg != st.peerBg {
		t.Errorf("separator inside the selected box = %v, want peerBg", bg)
	}
	if bg := m.separatorBg(st, 0, 9); bg != st.peerBg {
		t.Errorf("line between two cells of the selected column = %v, want peerBg", bg)
	}
	if bg := m.crossBg(st, 0, 1, 9, 10); bg != st.peerBg {
		t.Errorf("crossing inside the selected box = %v, want peerBg", bg)
	}
	if bg := m.crossBg(st, 36, 37, 45, 46); bg != nil {
		t.Errorf("crossing next to plain cells = %v, want none", bg)
	}
	if bg := m.separatorBg(st, 36, 37); bg != nil {
		t.Errorf("separator between a highlighted and a plain cell = %v, want none", bg)
	}
	if bg := m.separatorBg(st, 12, 13); bg != nil {
		t.Errorf("separator outside every highlight = %v, want none", bg)
	}
}

func TestNotesReadableOnSelectedCell(t *testing.T) {
	st := newStyles(true)
	m := playing(tutorialSeen()) // cursor on empty cell 2
	if s := m.cellStyle(st, 2, false); s.GetForeground() != st.text {
		t.Error("notes on the selected cell should use the main text color")
	}
	if s := m.cellStyle(st, 3, false); s.GetForeground() != st.muted {
		t.Error("notes elsewhere should stay muted")
	}
}
