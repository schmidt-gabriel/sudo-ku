package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

const (
	cellW  = 7           // characters per cell, " 1 2 3 ", close to square
	cellH  = 3           // lines per cell, room for three rows of notes
	boxW   = 3*cellW + 3 // three cells, two thin lines, one thick border
	boxH   = 3*cellH + 3 // three cells, two thin lines, one thick border
	boardW = 3*boxW + 1  // 73
	boardH = 3*boxH + 1  // 37
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
	return st.border.Render(left + box + thick + box + thick + box + right)
}

// thinRule draws the thin line under row, painting the parts that sit
// between two highlighted cells so highlights stay solid.
func (m Model) thinRule(st styles, row int) string {
	var b strings.Builder
	b.WriteString(st.border.Render("┠"))
	for col := range 9 {
		above, below := row*9+col, (row+1)*9+col
		b.WriteString(paint(st.border, m.separatorBg(st, above, below)).Render(strings.Repeat("─", cellW)))
		switch {
		case col == 8:
			b.WriteString(st.border.Render("┨"))
		case col%3 == 2:
			b.WriteString(st.border.Render("╂"))
		default:
			b.WriteString(paint(st.border, m.crossBg(st, above, above+1, below, below+1)).Render("┼"))
		}
	}
	return b.String()
}

// cellRow draws line sub (0-2) of one board row.
func (m Model) cellRow(st styles, row, sub int, conflicts map[int]bool) string {
	var b strings.Builder
	b.WriteString(st.border.Render("┃"))
	for col := range 9 {
		i := row*9 + col
		b.WriteString(m.cellStyle(st, i, conflicts[i]).Render(m.cellLine(i, sub)))
		if col%3 == 2 {
			b.WriteString(st.border.Render("┃"))
		} else {
			b.WriteString(paint(st.border, m.separatorBg(st, i, i+1)).Render("│"))
		}
	}
	return b.String()
}

// cellLine returns line sub (0-2) of cell i, exactly cellW characters wide.
// A filled cell shows its digit in the middle; an empty one shows its notes.
func (m Model) cellLine(i, sub int) string {
	if v := m.game.Values[i]; v != 0 {
		if sub == 1 {
			return "   " + string(rune('0'+v)) + "   "
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
	return " " + strings.Join(notes, " ") + " "
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

// separatorBg paints the line between two neighboring cells when both are
// highlighted, so a highlighted row, column or box reads as one block.
func (m Model) separatorBg(st styles, a, b int) color.Color {
	if m.cellBg(st, a) != nil && m.cellBg(st, b) != nil {
		return st.peerBg
	}
	return nil
}

// crossBg paints a line crossing when all four cells around it are highlighted.
func (m Model) crossBg(st styles, cells ...int) color.Color {
	for _, i := range cells {
		if m.cellBg(st, i) == nil {
			return nil
		}
	}
	return st.peerBg
}

// paint returns s with background bg, or s unchanged when bg is nil.
func paint(s lipgloss.Style, bg color.Color) lipgloss.Style {
	if bg == nil {
		return s
	}
	return s.Background(bg)
}

// cellStyle colors a cell: conflicts red, digits equal to the selected one
// in bold amber, givens bold, the player's digits blue, notes grey.
func (m Model) cellStyle(st styles, i int, conflict bool) lipgloss.Style {
	g, sel := m.game, m.cursor
	s := paint(lipgloss.NewStyle(), m.cellBg(st, i))
	switch {
	case conflict:
		return s.Foreground(st.conflict).Bold(true)
	case i != sel && g.Values[i] != 0 && g.Values[i] == g.Values[sel]:
		return s.Foreground(st.same).Bold(true)
	case g.IsGiven(i):
		return s.Foreground(st.text).Bold(true)
	case g.Values[i] != 0:
		return s.Foreground(st.accent)
	case i == sel:
		return s.Foreground(st.text) // notes stay readable on the strong background
	}
	return s.Foreground(st.muted)
}
