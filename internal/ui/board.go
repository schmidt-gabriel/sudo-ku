package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

const (
	cellW  = 9           // characters per cell: notes "1 2 3 4 5"
	cellH  = 2           // lines per cell: notes 1-5, then 6-9
	boxW   = 3*cellW + 3 // three cells, two thin lines, one thick border
	boxH   = 3*cellH + 3 // three cells, two thin lines, one thick border
	boardW = 3*boxW + 1  // 91
	boardH = 3*boxH + 1  // 28
)

// cellAt maps a position inside the board, where (0,0) is the top-left
// corner of the border, to a cell. It reports false on borders and lines.
func cellAt(x, y int) (int, bool) {
	if x < 1 || y < 1 || x >= boardW-1 || y >= boardH-1 {
		return 0, false
	}
	bx, inX := (x-1)/boxW, (x-1)%boxW
	by, inY := (y-1)/boxH, (y-1)%boxH
	if inX%(cellW+1) == cellW || inY%(cellH+1) == cellH {
		return 0, false
	}
	return (by*3+inY/(cellH+1))*9 + bx*3 + inX/(cellW+1), true
}

func (m Model) renderBoard(st styles) string {
	conflicts := map[int]bool{}
	for _, i := range m.game.Conflicts() {
		conflicts[i] = true
	}
	lines := []string{thickRule(st, "┏", "┯", "┳", "┓")}
	for row := range 9 {
		for sub := range cellH {
			lines = append(lines, m.cellRow(st, row, sub, conflicts))
		}
		switch {
		case row == 8:
			lines = append(lines, thickRule(st, "┗", "┷", "┻", "┛"))
		case row%3 == 2:
			lines = append(lines, thickRule(st, "┣", "┿", "╋", "┫"))
		default:
			lines = append(lines, m.thinRule(st, row))
		}
	}
	return strings.Join(lines, "\n")
}

// thickRule draws a thick horizontal border. thin is the junction with the
// thin lines between cells, thick the junction with the box borders.
func thickRule(st styles, left, thin, thick, right string) string {
	seg := strings.Repeat("━", cellW)
	box := seg + thin + seg + thin + seg
	return st.frame.Render(left + box + thick + box + thick + box + right)
}

// thinRule draws the thin line under row. Lines are never painted, so every
// highlighted cell sits inside the same even margin.
func (m Model) thinRule(st styles, row int) string {
	var b strings.Builder
	b.WriteString(st.frame.Render("┠"))
	for col := range 9 {
		b.WriteString(st.border.Render(strings.Repeat("─", cellW)))
		switch {
		case col == 8:
			b.WriteString(st.frame.Render("┨"))
		case col%3 == 2:
			b.WriteString(st.frame.Render("╂"))
		default:
			b.WriteString(st.border.Render("┼"))
		}
	}
	return b.String()
}

// cellRow draws line sub (0-2) of one board row.
func (m Model) cellRow(st styles, row, sub int, conflicts map[int]bool) string {
	var b strings.Builder
	b.WriteString(st.frame.Render("┃"))
	for col := range 9 {
		i := row*9 + col
		b.WriteString(m.renderCell(st, i, sub, conflicts[i]))
		if col%3 == 2 {
			b.WriteString(st.frame.Render("┃"))
		} else {
			b.WriteString(st.border.Render("│"))
		}
	}
	return b.String()
}

// renderCell draws line sub of cell i. An underline marks only the digit,
// never the padding around it.
func (m Model) renderCell(st styles, i, sub int, conflict bool) string {
	style := m.cellStyle(st, i, conflict)
	line := m.cellLine(i, sub)
	if !style.GetUnderline() {
		return style.Render(line)
	}
	plain := style.Underline(false)
	if sub != 0 {
		return plain.Render(line)
	}
	pad := strings.Repeat(" ", (cellW-1)/2)
	return plain.Render(pad) + style.Render(strings.TrimSpace(line)) + plain.Render(pad)
}

// cellLine returns line sub (0 or 1) of cell i, exactly cellW characters
// wide. A filled cell shows its digit centered on the first line. An empty
// cell shows notes 1-5 on the first line and 6-9, shifted half a step, on
// the second, each digit in a fixed spot.
func (m Model) cellLine(i, sub int) string {
	if v := m.game.Values[i]; v != 0 {
		if sub == 0 {
			pad := strings.Repeat(" ", (cellW-1)/2)
			return pad + string(rune('0'+v)) + pad
		}
		return strings.Repeat(" ", cellW)
	}
	first, last := uint8(1), uint8(5)
	if sub == 1 {
		first, last = 6, 9
	}
	var notes []string
	for d := first; d <= last; d++ {
		n := " "
		if m.game.HasNote(i, d) {
			n = string(rune('0' + d))
		}
		notes = append(notes, n)
	}
	line := strings.Join(notes, " ")
	if sub == 1 {
		line = " " + line + " "
	}
	return line
}

// cellBg is the background of cell i: strong for the selected cell, yellow
// for the tutorial's target, soft for the selection's row, column and box,
// and none otherwise.
func (m Model) cellBg(st styles, i int) color.Color {
	sel := m.cursor
	switch {
	case i == sel:
		return st.selectedBg
	case m.mode == modeTutorial && m.tutorialStep <= 2 && i == tutorialTarget:
		return st.targetBg
	case sudoku.Row(i) == sudoku.Row(sel) || sudoku.Col(i) == sudoku.Col(sel) || sudoku.Box(i) == sudoku.Box(sel):
		return st.peerBg
	}
	return nil
}

// paint returns s with background bg, or s unchanged when bg is nil.
func paint(s lipgloss.Style, bg color.Color) lipgloss.Style {
	if bg == nil {
		return s
	}
	return s.Background(bg)
}

// cellStyle colors a cell's text: conflicts red, clues bold white, the
// player's digits blue, notes grey. Digits equal to the selected one are
// underlined and bold but keep their color, so the player can always tell
// which digits they may change, on the selection too.
func (m Model) cellStyle(st styles, i int, conflict bool) lipgloss.Style {
	g, sel := m.game, m.cursor
	s := paint(lipgloss.NewStyle(), m.cellBg(st, i))
	if i != sel && g.Values[i] != 0 && g.Values[i] == g.Values[sel] {
		s = s.Underline(true).Bold(true)
	}
	switch {
	case conflict:
		return s.Foreground(st.conflict).Bold(true)
	case g.IsGiven(i):
		return s.Foreground(st.text).Bold(true)
	case g.Values[i] != 0:
		return s.Foreground(st.accent).Bold(s.GetBold() || i == sel)
	case i == sel:
		return s.Foreground(st.text) // notes stay readable on the selection
	}
	return s.Foreground(st.muted)
}
