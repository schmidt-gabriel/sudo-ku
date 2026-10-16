package sudoku

import (
	"strings"
	"testing"
)

// seventeenClue is a valid puzzle with the minimum possible number of givens.
const seventeenClue = "000000010400000000020000000000050407008000300001090000300400200050100000000806000"

func TestSolveClassic(t *testing.T) {
	got, ok := Solve(MustParse(classicPuzzle))
	if !ok {
		t.Fatal("no solution found")
	}
	if want := MustParse(classicSolution); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestSolveSeventeenClues(t *testing.T) {
	p := MustParse(seventeenClue)
	got, ok := Solve(p)
	if !ok {
		t.Fatal("no solution found")
	}
	if !IsComplete(got) {
		t.Fatalf("solution is not complete: %s", got)
	}
	for i, v := range p {
		if v != 0 && got[i] != v {
			t.Fatalf("given at cell %d changed from %d to %d", i, v, got[i])
		}
	}
	if n := CountSolutions(p, 2); n != 1 {
		t.Errorf("CountSolutions = %d, want 1", n)
	}
}

func TestSolveRejectsBrokenGrids(t *testing.T) {
	var dup Grid
	dup[0], dup[1] = 5, 5
	if _, ok := Solve(dup); ok {
		t.Error("grid with a duplicate was solved")
	}
	// Cell 8 needs a 9, but column 8 already has one.
	dead := MustParse("123456780" + "000000009" + strings.Repeat("0", 63))
	if _, ok := Solve(dead); ok {
		t.Error("dead-end grid was solved")
	}
	if n := CountSolutions(dead, 2); n != 0 {
		t.Errorf("CountSolutions(dead end) = %d, want 0", n)
	}
	var tooBig Grid
	tooBig[0] = 10
	if _, ok := Solve(tooBig); ok {
		t.Error("value 10 was accepted")
	}
}

func TestCountSolutions(t *testing.T) {
	if n := CountSolutions(Grid{}, 2); n != 2 {
		t.Errorf("empty grid: %d, want 2 (the limit)", n)
	}
	if n := CountSolutions(MustParse(classicPuzzle), 2); n != 1 {
		t.Errorf("classic puzzle: %d, want 1", n)
	}
	if n := CountSolutions(MustParse(classicSolution), 5); n != 1 {
		t.Errorf("solved grid: %d, want 1", n)
	}
}
