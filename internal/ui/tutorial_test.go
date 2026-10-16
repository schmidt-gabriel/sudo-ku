package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

func TestTutorialStartsOnFirstRun(t *testing.T) {
	m := newModel(&fakeStore{})
	if m.mode != modeTutorial || m.tutorialStep != 0 {
		t.Fatalf("mode %v step %d, want the tutorial's first step", m.mode, m.tutorialStep)
	}
}

func TestTutorialPuzzle(t *testing.T) {
	m := newModel(&fakeStore{})
	if n := sudoku.CountSolutions(m.game.Givens, 2); n != 1 {
		t.Fatalf("tutorial puzzle has %d solutions, want 1", n)
	}
	rows := map[int]bool{}
	for _, i := range tutorialBlanks {
		if m.game.Values[i] != 0 || rows[sudoku.Row(i)] {
			t.Fatalf("blank %d is filled or shares a row with another blank", i)
		}
		rows[sudoku.Row(i)] = true
	}
	for code, tx := range languages {
		if !strings.Contains(tx.TutorialSteps[2], string(rune('0'+tutorialSolution[tutorialTarget]))) {
			t.Errorf("%s: the place step names the wrong digit", code)
		}
	}
}

func TestTutorialWalkthrough(t *testing.T) {
	fs := &fakeStore{}
	m := newModel(fs)
	click := func(id string) { m, _ = m.handleClick(id) }
	wantStep := func(step int) {
		t.Helper()
		if m.tutorialStep != step {
			t.Fatalf("tutorial at step %d, want %d", m.tutorialStep, step)
		}
	}

	m = press(m, "enter")
	wantStep(1)
	m = press(m, "down", "down", "down", "down", "right", "right", "right", "right")
	wantStep(2)
	m = press(m, "5")
	wantStep(3)
	click("cell-2")
	m = press(m, "n", "1", "4")
	wantStep(4)
	m = press(m, "n")
	click("cell-15")
	m = press(m, "6") // row 1 already has a 6
	if !m.sawConflict {
		t.Fatal("conflict not noticed")
	}
	m = press(m, "u")
	wantStep(5)

	for _, i := range tutorialBlanks {
		if i == tutorialTarget {
			continue
		}
		click(fmt.Sprintf("cell-%d", i))
		click(fmt.Sprintf("num-%d", tutorialSolution[i]))
	}
	if m.mode != modeMenu || m.status != english.TutorialDone {
		t.Fatalf("mode %v status %q, want the menu with %q", m.mode, m.status, english.TutorialDone)
	}
	if !fs.settings.TutorialDone {
		t.Error("TutorialDone not saved")
	}
	if fs.saves != 0 {
		t.Errorf("the tutorial saved the game %d times, want 0", fs.saves)
	}
}

func TestTutorialSkip(t *testing.T) {
	fs := &fakeStore{}
	m := press(newModel(fs), "esc")
	if m.mode != modeMenu || !fs.settings.TutorialDone {
		t.Errorf("esc: mode %v done %v, want the menu with the tutorial marked done", m.mode, fs.settings.TutorialDone)
	}
}

func TestTutorialFromMenu(t *testing.T) {
	m := newModel(tutorialSeen())
	m = press(m, "down", "enter") // New game, Tutorial, Quit
	if m.mode != modeTutorial {
		t.Errorf("mode = %v, want tutorial", m.mode)
	}
}

func TestTutorialDoesNotTickOrQuitAsDone(t *testing.T) {
	fs := &fakeStore{}
	m := newModel(fs)
	next, _ := m.Update(tickMsg{})
	if next.(Model).game.Elapsed != 0 {
		t.Error("timer ran during the tutorial")
	}
	_, cmd := m.Update(key("q"))
	if _, ok := cmd().(tea.QuitMsg); !ok || fs.settings.TutorialDone {
		t.Error("q should quit without marking the tutorial done")
	}
}

func TestTutorialClickStarts(t *testing.T) {
	m := newModel(&fakeStore{})
	next, _ := m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if got := next.(Model).tutorialStep; got != 1 {
		t.Errorf("after a click the tutorial is at step %d, want 1", got)
	}
}
