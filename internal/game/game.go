// Package game holds the state of one sudoku game: moves, notes, undo
// history and elapsed time. It has no UI or disk dependencies.
package game

import (
	"math/rand/v2"
	"time"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

// maxHistory bounds the undo stack.
const maxHistory = 200

// Game is one game in progress. Notes[i] has bit v set when digit v is
// pencilled into cell i.
type Game struct {
	Givens     sudoku.Grid
	Values     sudoku.Grid
	Notes      [81]uint16
	Solution   sudoku.Grid
	Difficulty sudoku.Difficulty
	Elapsed    time.Duration

	undo, redo []snapshot
}

type snapshot struct {
	values sudoku.Grid
	notes  [81]uint16
}

// New starts a game with a freshly generated puzzle.
func New(d sudoku.Difficulty, rng *rand.Rand) *Game {
	puzzle, solution := sudoku.Generate(d, rng)
	return FromPuzzle(puzzle, solution, d)
}

// FromPuzzle starts a game from a known puzzle and its solution.
func FromPuzzle(puzzle, solution sudoku.Grid, d sudoku.Difficulty) *Game {
	return &Game{Givens: puzzle, Values: puzzle, Solution: solution, Difficulty: d}
}

// IsGiven reports whether cell i is one of the puzzle's read-only clues.
func (g *Game) IsGiven(i int) bool { return g.Givens[i] != 0 }

// HasNote reports whether digit v is pencilled into cell i.
func (g *Game) HasNote(i int, v uint8) bool { return g.Notes[i]&(1<<v) != 0 }

// Set places v (1-9) in cell i, clears the cell's notes and removes v from
// the notes of every peer.
func (g *Game) Set(i int, v uint8) {
	if g.IsGiven(i) || v < 1 || v > 9 || g.Values[i] == v {
		return
	}
	g.record()
	g.Values[i] = v
	g.Notes[i] = 0
	for _, p := range sudoku.Peers(i) {
		g.Notes[p] &^= 1 << v
	}
}

// Clear empties cell i, removing both its value and its notes.
func (g *Game) Clear(i int) {
	if g.IsGiven(i) || (g.Values[i] == 0 && g.Notes[i] == 0) {
		return
	}
	g.record()
	g.Values[i] = 0
	g.Notes[i] = 0
}

// ToggleNote adds or removes note v in cell i. Filled cells take no notes.
func (g *Game) ToggleNote(i int, v uint8) {
	if g.Values[i] != 0 || v < 1 || v > 9 {
		return
	}
	g.record()
	g.Notes[i] ^= 1 << v
}

// Undo reverts the last change, reporting false when there is none.
func (g *Game) Undo() bool {
	if len(g.undo) == 0 {
		return false
	}
	g.redo = append(g.redo, g.snapshot())
	g.restore(g.undo[len(g.undo)-1])
	g.undo = g.undo[:len(g.undo)-1]
	return true
}

// Redo reapplies the last undone change, reporting false when there is none.
func (g *Game) Redo() bool {
	if len(g.redo) == 0 {
		return false
	}
	g.undo = append(g.undo, g.snapshot())
	g.restore(g.redo[len(g.redo)-1])
	g.redo = g.redo[:len(g.redo)-1]
	return true
}

// Conflicts returns the filled cells that break a rule.
func (g *Game) Conflicts() []int { return sudoku.Conflicts(g.Values) }

// Won reports whether the board is full and breaks no rule.
func (g *Game) Won() bool { return sudoku.IsComplete(g.Values) }

// Tick adds d to the elapsed time.
func (g *Game) Tick(d time.Duration) { g.Elapsed += d }

// record saves the current state for Undo and forgets undone changes.
func (g *Game) record() {
	g.undo = append(g.undo, g.snapshot())
	if len(g.undo) > maxHistory {
		g.undo = g.undo[1:]
	}
	g.redo = nil
}

func (g *Game) snapshot() snapshot { return snapshot{values: g.Values, notes: g.Notes} }

func (g *Game) restore(s snapshot) { g.Values, g.Notes = s.values, s.notes }
