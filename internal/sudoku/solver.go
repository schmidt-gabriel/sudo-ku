package sudoku

import (
	"math/bits"
	"math/rand/v2"
)

// allDigits has bits 1 to 9 set.
const allDigits uint16 = 0b11_1111_1110

// Solve returns a solution of g, or false if g has none.
func Solve(g Grid) (Grid, bool) {
	s, ok := newSolver(g)
	if !ok {
		return g, false
	}
	var out Grid
	found := s.search(nil, func() bool { out = s.g; return true })
	return out, found
}

// CountSolutions counts the solutions of g, stopping once it reaches limit.
// CountSolutions(g, 2) == 1 means g has exactly one solution.
func CountSolutions(g Grid, limit int) int {
	s, ok := newSolver(g)
	if !ok {
		return 0
	}
	n := 0
	s.search(nil, func() bool { n++; return n >= limit })
	return n
}

// solver is a backtracking search that tracks the digits used in each row,
// column and box as bitmasks.
type solver struct {
	g                 Grid
	rows, cols, boxes [9]uint16
}

// newSolver loads g, reporting false if g already breaks a rule.
func newSolver(g Grid) (*solver, bool) {
	s := &solver{}
	for i, v := range g {
		if v == 0 {
			continue
		}
		if v > 9 || s.candidates(i)&(1<<v) == 0 {
			return nil, false
		}
		s.place(i, v)
	}
	return s, true
}

func (s *solver) candidates(i int) uint16 {
	return allDigits &^ (s.rows[Row(i)] | s.cols[Col(i)] | s.boxes[Box(i)])
}

func (s *solver) place(i int, v uint8) {
	bit := uint16(1) << v
	s.g[i] = v
	s.rows[Row(i)] |= bit
	s.cols[Col(i)] |= bit
	s.boxes[Box(i)] |= bit
}

func (s *solver) unplace(i int, v uint8) {
	mask := ^(uint16(1) << v)
	s.g[i] = 0
	s.rows[Row(i)] &= mask
	s.cols[Col(i)] &= mask
	s.boxes[Box(i)] &= mask
}

// search fills the empty cells, calling found at each solution, and stops as
// soon as found returns true. A non-nil rng shuffles the order digits are tried.
func (s *solver) search(rng *rand.Rand, found func() bool) bool {
	best, bestN := -1, 10
	for i, v := range s.g {
		if v != 0 {
			continue
		}
		n := bits.OnesCount16(s.candidates(i))
		if n == 0 {
			return false
		}
		if n < bestN {
			best, bestN = i, n
		}
	}
	if best == -1 {
		return found()
	}
	digits := digitsIn(s.candidates(best))
	if rng != nil {
		rng.Shuffle(len(digits), func(a, b int) { digits[a], digits[b] = digits[b], digits[a] })
	}
	for _, d := range digits {
		s.place(best, d)
		if s.search(rng, found) {
			return true
		}
		s.unplace(best, d)
	}
	return false
}

func digitsIn(mask uint16) []uint8 {
	var ds []uint8
	for d := uint8(1); d <= 9; d++ {
		if mask&(1<<d) != 0 {
			ds = append(ds, d)
		}
	}
	return ds
}
