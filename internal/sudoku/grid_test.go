package sudoku

import (
	"slices"
	"strings"
	"testing"
)

const (
	classicPuzzle   = "530070000600195000098000060800060003400803001700020006060000280000419005000080079"
	classicSolution = "534678912672195348198342567859761423426853791713924856961537284287419635345286179"
)

func TestRowColBox(t *testing.T) {
	cases := []struct{ i, row, col, box int }{
		{0, 0, 0, 0}, {8, 0, 8, 2}, {30, 3, 3, 4}, {40, 4, 4, 4}, {72, 8, 0, 6}, {80, 8, 8, 8},
	}
	for _, c := range cases {
		if Row(c.i) != c.row || Col(c.i) != c.col || Box(c.i) != c.box {
			t.Errorf("cell %d: got row=%d col=%d box=%d, want %d %d %d",
				c.i, Row(c.i), Col(c.i), Box(c.i), c.row, c.col, c.box)
		}
	}
}

func TestPeers(t *testing.T) {
	for i := range 81 {
		if n := len(Peers(i)); n != 20 {
			t.Fatalf("Peers(%d) has %d cells, want 20", i, n)
		}
	}
	p := Peers(0)
	for _, want := range []int{1, 8, 9, 10, 20, 72} {
		if !slices.Contains(p, want) {
			t.Errorf("Peers(0) is missing %d", want)
		}
	}
	for _, notWant := range []int{0, 30, 80} {
		if slices.Contains(p, notWant) {
			t.Errorf("Peers(0) contains %d", notWant)
		}
	}
}

func TestConflicts(t *testing.T) {
	var g Grid
	if c := Conflicts(g); len(c) != 0 {
		t.Fatalf("empty grid has conflicts %v", c)
	}
	g[0], g[3] = 5, 5   // same row
	g[40], g[50] = 7, 7 // same box
	g[80] = 5           // not a peer of 0 or 3
	want := []int{0, 3, 40, 50}
	if got := Conflicts(g); !slices.Equal(got, want) {
		t.Errorf("Conflicts = %v, want %v", got, want)
	}
}

func TestIsComplete(t *testing.T) {
	sol := MustParse(classicSolution)
	if !IsComplete(sol) {
		t.Fatal("solved grid reported incomplete")
	}
	if IsComplete(MustParse(classicPuzzle)) {
		t.Error("grid with blanks reported complete")
	}
	bad := sol
	bad[0], bad[1] = bad[1], bad[0]
	if IsComplete(bad) {
		t.Error("grid with swapped cells reported complete")
	}
}

func TestParseAndString(t *testing.T) {
	g, err := Parse(classicPuzzle)
	if err != nil {
		t.Fatal(err)
	}
	if g[0] != 5 || g[1] != 3 || g[2] != 0 {
		t.Errorf("first cells = %v, want [5 3 0]", g[:3])
	}
	if got, want := g.String(), strings.ReplaceAll(classicPuzzle, "0", "."); got != want {
		t.Errorf("String() = %s, want %s", got, want)
	}
	dotted, err := Parse("53..7....\n" + classicPuzzle[9:])
	if err != nil || dotted != g {
		t.Errorf("dots and newlines not parsed: %v", err)
	}
	for _, bad := range []string{"", classicPuzzle[:80], classicPuzzle + "1", "x" + classicPuzzle[1:]} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) succeeded, want an error", bad)
		}
	}
}
