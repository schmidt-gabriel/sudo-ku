// Package sudoku implements the rules, the solver and the puzzle generator.
// It has no UI or I/O dependencies, so any front end can reuse it.
package sudoku

import (
	"fmt"
	"strings"
)

// Grid is a 9x9 board in row-major order: cell i is at row i/9, column i%9.
// 0 means empty.
type Grid [81]uint8

// Row returns the row (0-8) of cell i.
func Row(i int) int { return i / 9 }

// Col returns the column (0-8) of cell i.
func Col(i int) int { return i % 9 }

// Box returns the 3x3 box (0-8, row-major) of cell i.
func Box(i int) int { return (i/27)*3 + (i%9)/3 }

var peers [81][]int

func init() {
	for i := range 81 {
		for j := range 81 {
			if i != j && (Row(i) == Row(j) || Col(i) == Col(j) || Box(i) == Box(j)) {
				peers[i] = append(peers[i], j)
			}
		}
	}
}

// Peers returns the 20 cells that share a row, column or box with cell i.
// Callers must not modify the returned slice.
func Peers(i int) []int { return peers[i] }

// Conflicts returns, in ascending order, the filled cells whose value also
// appears in one of their peers.
func Conflicts(g Grid) []int {
	var out []int
	for i, v := range g {
		if v == 0 {
			continue
		}
		for _, p := range peers[i] {
			if g[p] == v {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

// IsComplete reports whether every cell is filled and no rule is broken.
func IsComplete(g Grid) bool {
	for _, v := range g {
		if v == 0 {
			return false
		}
	}
	return len(Conflicts(g)) == 0
}

// Parse reads a grid of 81 cells: digits 1-9, and '0' or '.' for empty.
// Whitespace is ignored.
func Parse(s string) (Grid, error) {
	var g Grid
	n := 0
	for _, r := range s {
		switch {
		case r == ' ' || r == '\n' || r == '\t' || r == '\r':
			continue
		case n == 81:
			return Grid{}, fmt.Errorf("sudoku: more than 81 cells")
		case r == '.' || r == '0':
			n++
		case r >= '1' && r <= '9':
			g[n] = uint8(r - '0')
			n++
		default:
			return Grid{}, fmt.Errorf("sudoku: invalid character %q", r)
		}
	}
	if n != 81 {
		return Grid{}, fmt.Errorf("sudoku: got %d cells, want 81", n)
	}
	return g, nil
}

// MustParse is like Parse but panics on error. Use it for literals.
func MustParse(s string) Grid {
	g, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return g
}

// String returns the grid as 81 characters, '.' for empty cells.
func (g Grid) String() string {
	var b strings.Builder
	for _, v := range g {
		if v == 0 {
			b.WriteByte('.')
		} else {
			b.WriteByte('0' + v)
		}
	}
	return b.String()
}
