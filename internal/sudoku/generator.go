package sudoku

import "math/rand/v2"

// Difficulty is the target difficulty of a generated puzzle.
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
	Expert
)

// Difficulties lists every difficulty, easiest first.
var Difficulties = []Difficulty{Easy, Medium, Hard, Expert}

func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	case Expert:
		return "Expert"
	}
	return "Unknown"
}

// Valid reports whether d is one of the defined difficulties.
func (d Difficulty) Valid() bool { return d >= Easy && d <= Expert }

// clueTarget is how many givens a puzzle keeps. Expert has no target: the
// generator removes as many as it can.
var clueTarget = map[Difficulty]int{Easy: 40, Medium: 32, Hard: 27, Expert: 0}

// maxAttempts bounds how many removal passes Generate tries.
const maxAttempts = 10

// Generate returns a puzzle with exactly one solution, and that solution.
// Difficulty is approximated by the number of givens.
func Generate(d Difficulty, rng *rand.Rand) (puzzle, solution Grid) {
	best := 82
	for range maxAttempts {
		p, s := generateOnce(clueTarget[d], rng)
		if n := countGivens(p); n < best {
			puzzle, solution, best = p, s, n
		}
		if best <= clueTarget[d] || (d == Expert && best < clueTarget[Hard]) {
			break
		}
	}
	return puzzle, solution
}

// generateOnce fills a random grid, then empties cells in random order,
// keeping each removal only while the solution stays unique.
func generateOnce(target int, rng *rand.Rand) (puzzle, solution Grid) {
	s, _ := newSolver(Grid{})
	s.search(rng, func() bool { solution = s.g; return true })
	puzzle = solution
	givens := 81
	for _, i := range rng.Perm(81) {
		if givens <= target {
			break
		}
		v := puzzle[i]
		puzzle[i] = 0
		if CountSolutions(puzzle, 2) != 1 {
			puzzle[i] = v
			continue
		}
		givens--
	}
	return puzzle, solution
}

func countGivens(g Grid) int {
	n := 0
	for _, v := range g {
		if v != 0 {
			n++
		}
	}
	return n
}
