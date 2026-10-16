package ui

import (
	"math/bits"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

var tutorialSolution = sudoku.MustParse("534678912672195348198342567859761423426853791713924856961537284287419635345286179")

// tutorialBlanks are the empty cells of the tutorial puzzle. Each is in a
// different row, so each has exactly one digit that fits.
var tutorialBlanks = []int{2, 15, 30, 40, 58, 75}

// tutorialTarget is the cell the tutorial asks the player to fill first.
const tutorialTarget = 40

// tutorialStep is one step of the tutorial. Its text is texts.TutorialSteps
// at the same index.
type tutorialStep struct {
	// done reports whether the player finished the step. Nil means the
	// step waits for enter.
	done func(m *Model) bool
}

var tutorialSteps = []tutorialStep{
	{},
	{done: func(m *Model) bool { return m.cursor == tutorialTarget }},
	{done: func(m *Model) bool {
		return m.game.Values[tutorialTarget] == tutorialSolution[tutorialTarget]
	}},
	{done: func(m *Model) bool {
		for i, notes := range m.game.Notes {
			if m.game.Values[i] == 0 && bits.OnesCount16(notes) >= 2 {
				return true
			}
		}
		return false
	}},
	{done: func(m *Model) bool {
		if len(m.game.Conflicts()) > 0 {
			m.sawConflict = true
		}
		return m.sawConflict && len(m.game.Conflicts()) == 0
	}},
	{done: func(m *Model) bool { return m.game.Won() }},
}

func (m *Model) startTutorial() {
	puzzle := tutorialSolution
	for _, i := range tutorialBlanks {
		puzzle[i] = 0
	}
	m.game = game.FromPuzzle(puzzle, tutorialSolution, sudoku.Easy)
	m.mode, m.cursor, m.notesMode, m.showHelp = modeTutorial, 0, false, false
	m.tutorialStep, m.sawConflict = 0, false
}

// advanceTutorial skips past every step the player has already completed,
// and finishes the tutorial after the last one.
func (m *Model) advanceTutorial() {
	if m.mode != modeTutorial {
		return
	}
	for m.tutorialStep < len(tutorialSteps) {
		done := tutorialSteps[m.tutorialStep].done
		if done == nil || !done(m) {
			return
		}
		m.tutorialStep++
	}
	m.finishTutorial(m.t.TutorialDone)
}

// finishTutorial records that the tutorial was seen and opens the menu.
func (m *Model) finishTutorial(status string) {
	m.settings.TutorialDone = true
	if err := m.store.SaveSettings(m.settings); err != nil {
		status = m.t.SaveFailed
	}
	m.toMenu()
	m.status = status
}
