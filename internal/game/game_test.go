package game

import (
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

var (
	testPuzzle   = sudoku.MustParse("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	testSolution = sudoku.MustParse("534678912672195348198342567859761423426853791713924856961537284287419635345286179")
)

func newGame() *Game { return FromPuzzle(testPuzzle, testSolution, sudoku.Easy) }

func TestNew(t *testing.T) {
	g := New(sudoku.Medium, rand.New(rand.NewPCG(1, 1)))
	if g.Givens != g.Values || g.Difficulty != sudoku.Medium {
		t.Fatal("new game should start with Values == Givens and the requested difficulty")
	}
	if n := sudoku.CountSolutions(g.Givens, 2); n != 1 {
		t.Errorf("puzzle has %d solutions, want 1", n)
	}
	if !sudoku.IsComplete(g.Solution) {
		t.Error("solution is not complete")
	}
}

func TestSetPlacesValue(t *testing.T) {
	g := newGame()
	g.Set(2, 4)
	if g.Values[2] != 4 {
		t.Fatalf("Values[2] = %d, want 4", g.Values[2])
	}
}

func TestGivensAreReadOnly(t *testing.T) {
	g := newGame()
	g.Set(0, 9)
	g.Clear(0)
	g.ToggleNote(0, 1)
	if g.Values[0] != 5 || g.Notes[0] != 0 {
		t.Errorf("given changed: value %d notes %b", g.Values[0], g.Notes[0])
	}
	if g.Undo() {
		t.Error("no-op edits on a given must not be recorded")
	}
}

func TestSetRemovesNoteFromPeers(t *testing.T) {
	g := newGame()
	// 3 shares the row with 2, 11 the column, 10 the box; 30 is not a peer.
	for _, i := range []int{3, 10, 11, 30} {
		g.ToggleNote(i, 4)
	}
	g.ToggleNote(3, 7)
	g.Set(2, 4)
	for _, i := range []int{3, 10, 11} {
		if g.HasNote(i, 4) {
			t.Errorf("peer %d still has note 4", i)
		}
	}
	if !g.HasNote(30, 4) {
		t.Error("non-peer 30 lost note 4")
	}
	if !g.HasNote(3, 7) {
		t.Error("peer 3 lost an unrelated note")
	}
}

func TestSetClearsOwnNotes(t *testing.T) {
	g := newGame()
	g.ToggleNote(2, 1)
	g.ToggleNote(2, 2)
	g.Set(2, 4)
	if g.Notes[2] != 0 {
		t.Errorf("Notes[2] = %b, want 0", g.Notes[2])
	}
}

func TestToggleNote(t *testing.T) {
	g := newGame()
	g.ToggleNote(2, 3)
	if !g.HasNote(2, 3) {
		t.Fatal("note 3 not added")
	}
	g.ToggleNote(2, 3)
	if g.HasNote(2, 3) {
		t.Fatal("note 3 not removed")
	}
	g.ToggleNote(2, 0)
	g.ToggleNote(2, 10)
	g.Set(2, 4)
	g.ToggleNote(2, 5)
	if g.Notes[2] != 0 {
		t.Errorf("Notes[2] = %b, want 0 (out of range digits and filled cell)", g.Notes[2])
	}
}

func TestClear(t *testing.T) {
	g := newGame()
	g.Set(2, 4)
	g.Clear(2)
	if g.Values[2] != 0 {
		t.Errorf("Values[2] = %d after Clear, want 0", g.Values[2])
	}
	g.ToggleNote(2, 1)
	g.Clear(2)
	if g.Notes[2] != 0 {
		t.Errorf("Notes[2] = %b after Clear, want 0", g.Notes[2])
	}
}

func TestUndoRedo(t *testing.T) {
	g := newGame()
	g.Set(2, 4)
	g.Set(3, 6)
	g.Undo()
	if g.Values[3] != 0 || g.Values[2] != 4 {
		t.Fatalf("after one undo: cells 2,3 = %d,%d, want 4,0", g.Values[2], g.Values[3])
	}
	g.Undo()
	if g.Values[2] != 0 {
		t.Fatalf("after two undos: cell 2 = %d, want 0", g.Values[2])
	}
	if g.Undo() {
		t.Fatal("Undo with empty history returned true")
	}
	g.Redo()
	g.Redo()
	if g.Values[2] != 4 || g.Values[3] != 6 {
		t.Fatalf("after two redos: cells 2,3 = %d,%d, want 4,6", g.Values[2], g.Values[3])
	}
	if g.Redo() {
		t.Fatal("Redo with nothing undone returned true")
	}
}

func TestUndoRestoresPeerNotes(t *testing.T) {
	g := newGame()
	g.ToggleNote(3, 4)
	g.Set(2, 4)
	g.Undo()
	if g.Values[2] != 0 || !g.HasNote(3, 4) {
		t.Error("undo did not restore both the cell and the peer's note")
	}
}

func TestNewChangeClearsRedo(t *testing.T) {
	g := newGame()
	g.Set(2, 4)
	g.Undo()
	g.Set(3, 6)
	if g.Redo() {
		t.Error("Redo succeeded after a new change")
	}
}

func TestSameValueIsNotRecorded(t *testing.T) {
	g := newGame()
	g.Set(2, 4)
	g.Set(2, 4)
	g.Undo()
	if g.Values[2] != 0 || g.Undo() {
		t.Error("setting the same value twice created two history entries")
	}
}

func TestHistoryIsBounded(t *testing.T) {
	g := newGame()
	for range 250 {
		g.ToggleNote(2, 1)
	}
	if len(g.undo) != maxHistory {
		t.Errorf("undo has %d entries, want %d", len(g.undo), maxHistory)
	}
}

func TestWonAndConflicts(t *testing.T) {
	g := newGame()
	g.Set(2, 5) // 5 is already at cell 0, same row
	if got := g.Conflicts(); !slices.Equal(got, []int{0, 2}) {
		t.Errorf("Conflicts = %v, want [0 2]", got)
	}
	g.Clear(2)
	for i, v := range g.Solution {
		g.Set(i, v)
	}
	if !g.Won() {
		t.Error("full correct board is not a win")
	}
}

func TestTick(t *testing.T) {
	g := newGame()
	g.Tick(time.Second)
	g.Tick(2 * time.Second)
	if g.Elapsed != 3*time.Second {
		t.Errorf("Elapsed = %v, want 3s", g.Elapsed)
	}
}

func TestCompleted(t *testing.T) {
	g := newGame()
	if g.Completed(5) {
		t.Fatal("5 is not complete in the starting puzzle")
	}
	var fives []int
	for i, v := range g.Solution {
		if v == 5 && !g.IsGiven(i) {
			fives = append(fives, i)
		}
	}
	for _, i := range fives[:len(fives)-1] {
		g.Set(i, 5)
	}
	if g.Completed(5) {
		t.Fatal("5 is complete with one still missing")
	}
	last := fives[len(fives)-1]
	g.Set(last, 5)
	if !g.Completed(5) {
		t.Fatal("all nine 5s are in place but Completed(5) is false")
	}
	g.Clear(last)
	g.Set(last, 4) // wrong digit where a 5 belongs
	if g.Completed(5) {
		t.Error("5 counted as complete with a wrong digit in one of its cells")
	}
	if g.Completed(0) || g.Completed(10) {
		t.Error("out-of-range digits are never complete")
	}
}
