package store

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

const historyFile = "history.json"

// maxResults bounds how many finished games the history keeps.
const maxResults = 100

// Result is one finished game.
type Result struct {
	Difficulty sudoku.Difficulty `json:"difficulty"`
	Elapsed    time.Duration     `json:"elapsed"`
	FinishedAt time.Time         `json:"finishedAt"`
}

// LoadHistory returns finished games, newest first, or nil if there are none.
func (s *Store) LoadHistory() ([]Result, error) {
	b, err := s.read(historyFile)
	if b == nil || err != nil {
		return nil, err
	}
	var h []Result
	if err := json.Unmarshal(b, &h); err != nil {
		return nil, fmt.Errorf("store: %s: %w", historyFile, err)
	}
	return h, nil
}

// AddResult records a finished game at the top of the history. A corrupt
// history file is replaced rather than blocking the new result.
func (s *Store) AddResult(r Result) error {
	h, _ := s.LoadHistory()
	h = append([]Result{r}, h...)
	if len(h) > maxResults {
		h = h[:maxResults]
	}
	b, err := json.Marshal(h)
	if err != nil {
		return err
	}
	return s.write(historyFile, b)
}
