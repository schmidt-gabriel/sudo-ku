package ui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

const (
	cellW  = 5           // characters per cell, room for "1 2 3"
	cellH  = 3           // lines per cell, room for three rows of notes
	boxW   = 3*cellW + 3 // three cells, two thin separators, one thick border
	boxH   = 3*cellH + 1 // three cells, one thick border
	boardW = 3*boxW + 1  // 55
	boardH = 3*boxH + 1  // 31
)

// cellAt maps a position inside the board, where (0,0) is the top-left
// corner of the border, to a cell. It reports false on borders and separators.
func cellAt(x, y int) (int, bool) {
	if x < 1 || y < 1 || x >= boardW-1 || y >= boardH-1 {
		return 0, false
	}
	bx, inX := (x-1)/boxW, (x-1)%boxW
	by, inY := (y-1)/boxH, (y-1)%boxH
	if inX%(cellW+1) == cellW || inY == boxH-1 {
		return 0, false
	}
	return (by*3+inY/cellH)*9 + bx*3 + inX/(cellW+1), true
}

func (m Model) renderBoard(st styles) string {
	conflicts := map[int]bool{}
	for _, i := range m.game.Conflicts() {
		conflicts[i] = true
	}
	rule := func(left, mid, right string) string {
		seg := strings.Repeat("━", boxW-1)
		return st.border.Render(left + seg + mid + seg + mid + seg + right)
	}
	lines := []string{rule("┏", "┳", "┓")}
	for row := range 9 {
		for sub := range cellH {
			var b strings.Builder
			b.WriteString(st.border.Render("┃"))
			for col := range 9 {
				i := row*9 + col
				b.WriteString(m.cellStyle(st, i, conflicts[i]).Render(m.cellLine(i, sub)))
				if col%3 == 2 {
					b.WriteString(st.border.Render("┃"))
				} else {
					b.WriteString(st.border.Render("│"))
				}
			}
			lines = append(lines, b.String())
		}
		if row == 2 || row == 5 {
			lines = append(lines, rule("┣", "╋", "┫"))
		}
	}
	lines = append(lines, rule("┗", "┻", "┛"))
	return strings.Join(lines, "\n")
}

// cellLine returns line sub (0-2) of cell i, exactly cellW characters wide.
// A filled cell shows its digit in the middle; an empty one shows its notes.
func (m Model) cellLine(i, sub int) string {
	if v := m.game.Values[i]; v != 0 {
		if sub == 1 {
			return "  " + string(rune('0'+v)) + "  "
		}
		return strings.Repeat(" ", cellW)
	}
	notes := make([]string, 3)
	for k := range notes {
		d := uint8(sub*3 + k + 1)
		notes[k] = " "
		if m.game.HasNote(i, d) {
			notes[k] = string(rune('0' + d))
		}
	}
	return strings.Join(notes, " ")
}

func (m Model) cellStyle(st styles, i int, conflict bool) lipgloss.Style {
	g, sel := m.game, m.cursor
	s := lipgloss.NewStyle()
	switch {
	case i == sel:
		s = s.Background(st.selectedBg)
	case g.Values[sel] != 0 && g.Values[i] == g.Values[sel]:
		s = s.Background(st.sameBg)
	case sudoku.Row(i) == sudoku.Row(sel) || sudoku.Col(i) == sudoku.Col(sel) || sudoku.Box(i) == sudoku.Box(sel):
		s = s.Background(st.peerBg)
	}
	switch {
	case conflict:
		return s.Foreground(st.conflict).Bold(true)
	case g.IsGiven(i):
		return s.Foreground(st.text).Bold(true)
	case g.Values[i] != 0:
		return s.Foreground(st.accent)
	}
	return s.Foreground(st.muted)
}
