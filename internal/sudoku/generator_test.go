package sudoku

import (
	"math/rand/v2"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	for _, d := range Difficulties {
		for seed := uint64(1); seed <= 3; seed++ {
			p, s := Generate(d, rand.New(rand.NewPCG(seed, 0)))
			if !IsComplete(s) {
				t.Fatalf("%v seed %d: solution is not complete", d, seed)
			}
			for i, v := range p {
				if v != 0 && v != s[i] {
					t.Fatalf("%v seed %d: given at %d differs from the solution", d, seed, i)
				}
			}
			if n := CountSolutions(p, 2); n != 1 {
				t.Fatalf("%v seed %d: %d solutions, want 1", d, seed, n)
			}
			n := countGivens(p)
			if d == Expert {
				if n >= clueTarget[Hard] {
					t.Errorf("Expert seed %d: %d givens, want fewer than %d", seed, n, clueTarget[Hard])
				}
			} else if n != clueTarget[d] {
				t.Errorf("%v seed %d: %d givens, want %d", d, seed, n, clueTarget[d])
			}
		}
	}
}

func TestGenerateExpertIsQuick(t *testing.T) {
	if testing.Short() {
		t.Skip("timing test")
	}
	for seed := uint64(1); seed <= 5; seed++ {
		start := time.Now()
		Generate(Expert, rand.New(rand.NewPCG(seed, 0)))
		if d := time.Since(start); d > 2*time.Second {
			t.Errorf("Expert seed %d took %v; starting a game should feel instant", seed, d)
		}
	}
}

func TestGenerateIsDeterministic(t *testing.T) {
	a, _ := Generate(Medium, rand.New(rand.NewPCG(7, 7)))
	b, _ := Generate(Medium, rand.New(rand.NewPCG(7, 7)))
	if a != b {
		t.Error("the same seed produced different puzzles")
	}
}

func TestDifficulty(t *testing.T) {
	if Easy.String() != "Easy" || Expert.String() != "Expert" || Difficulty(9).String() != "Unknown" {
		t.Error("unexpected Difficulty.String output")
	}
	if !Hard.Valid() || Difficulty(-1).Valid() || Difficulty(4).Valid() {
		t.Error("unexpected Difficulty.Valid output")
	}
}
