package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

// validNotes has bits 1 to 9 set; any other bit in a note mask is corrupt.
const validNotes uint16 = 0b11_1111_1110

// savedGame is the JSON form of a Game. The undo history is not saved.
type savedGame struct {
	Givens     string            `json:"givens"`
	Values     string            `json:"values"`
	Solution   string            `json:"solution"`
	Notes      [81]uint16        `json:"notes"`
	Difficulty sudoku.Difficulty `json:"difficulty"`
	ElapsedMs  int64             `json:"elapsedMs"`
}

// MarshalJSON implements json.Marshaler.
func (g *Game) MarshalJSON() ([]byte, error) {
	return json.Marshal(savedGame{
		Givens:     g.Givens.String(),
		Values:     g.Values.String(),
		Solution:   g.Solution.String(),
		Notes:      g.Notes,
		Difficulty: g.Difficulty,
		ElapsedMs:  g.Elapsed.Milliseconds(),
	})
}

// UnmarshalJSON implements json.Unmarshaler. It rejects data that could not
// come from a real game, so a corrupt save never reaches the UI.
func (g *Game) UnmarshalJSON(b []byte) error {
	var s savedGame
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	givens, err := sudoku.Parse(s.Givens)
	if err != nil {
		return fmt.Errorf("givens: %w", err)
	}
	values, err := sudoku.Parse(s.Values)
	if err != nil {
		return fmt.Errorf("values: %w", err)
	}
	solution, err := sudoku.Parse(s.Solution)
	if err != nil {
		return fmt.Errorf("solution: %w", err)
	}
	if !sudoku.IsComplete(solution) {
		return errors.New("solution is not a solved grid")
	}
	for i := range 81 {
		if givens[i] != 0 && (values[i] != givens[i] || solution[i] != givens[i]) {
			return fmt.Errorf("cell %d does not match its given", i)
		}
		if s.Notes[i]&^validNotes != 0 {
			return fmt.Errorf("cell %d has invalid notes", i)
		}
	}
	if !s.Difficulty.Valid() {
		return fmt.Errorf("unknown difficulty %d", s.Difficulty)
	}
	if s.ElapsedMs < 0 {
		return errors.New("negative elapsed time")
	}
	*g = Game{
		Givens:     givens,
		Values:     values,
		Notes:      s.Notes,
		Solution:   solution,
		Difficulty: s.Difficulty,
		Elapsed:    time.Duration(s.ElapsedMs) * time.Millisecond,
	}
	return nil
}
