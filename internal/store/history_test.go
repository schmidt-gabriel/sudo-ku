package store

import (
	"testing"
	"time"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

func TestHistoryMissingIsEmpty(t *testing.T) {
	h, err := newStore(t).LoadHistory()
	if h != nil || err != nil {
		t.Errorf("LoadHistory() = %v, %v; want nil, nil", h, err)
	}
}

func TestHistoryNewestFirstAndCapped(t *testing.T) {
	s := newStore(t)
	start := time.Date(2026, 10, 16, 14, 0, 0, 0, time.UTC)
	for n := range maxResults + 5 {
		r := Result{Difficulty: sudoku.Hard, Elapsed: time.Duration(n) * time.Second, FinishedAt: start.Add(time.Duration(n) * time.Minute)}
		if err := s.AddResult(r); err != nil {
			t.Fatal(err)
		}
	}
	h, err := s.LoadHistory()
	if err != nil {
		t.Fatal(err)
	}
	if len(h) != maxResults {
		t.Fatalf("history has %d results, want %d", len(h), maxResults)
	}
	newest := h[0]
	if newest.Elapsed != time.Duration(maxResults+4)*time.Second || newest.Difficulty != sudoku.Hard ||
		!newest.FinishedAt.Equal(start.Add(time.Duration(maxResults+4)*time.Minute)) {
		t.Errorf("newest result = %+v", newest)
	}
	if h[len(h)-1].Elapsed != 5*time.Second {
		t.Errorf("oldest kept result = %v, want the 6th one added", h[len(h)-1].Elapsed)
	}
}

func TestCorruptHistoryStartsOver(t *testing.T) {
	s := newStore(t)
	if err := s.write(historyFile, []byte("{oops")); err != nil {
		t.Fatal(err)
	}
	if h, err := s.LoadHistory(); h != nil || err == nil {
		t.Errorf("corrupt history: %v, %v; want nil and an error", h, err)
	}
	if err := s.AddResult(Result{Difficulty: sudoku.Easy}); err != nil {
		t.Fatal(err)
	}
	if h, err := s.LoadHistory(); err != nil || len(h) != 1 {
		t.Errorf("after AddResult: %v, %v; want one result", h, err)
	}
}
