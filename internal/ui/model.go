// Package ui is the terminal front end: a single Bubble Tea model whose mode
// decides what is drawn and how input is handled.
package ui

import (
	"math/rand/v2"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
	"github.com/schmidt-gabriel/sudo-ku/internal/store"
	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

type mode int

const (
	modeMenu mode = iota
	modeDifficulty
	modePlaying
	modeTutorial
	modeWon
)

// Store is the persistence the UI needs; *store.Store implements it.
type Store interface {
	LoadGame() (*game.Game, error)
	SaveGame(*game.Game) error
	DeleteGame() error
	LoadSettings() (store.Settings, error)
	SaveSettings(store.Settings) error
	LoadHistory() ([]store.Result, error)
	AddResult(store.Result) error
}

type menuItem int

const (
	itemContinue menuItem = iota
	itemNewGame
	itemTutorial
	itemLanguage
	itemQuit
)

// Choices in the quit dialog, in display order.
const (
	quitSave = iota
	quitDiscard
	quitCancel
)

type tickMsg struct{}

// Model is the whole UI state.
type Model struct {
	store    Store
	rng      *rand.Rand
	settings store.Settings
	t        *texts           // the player's language
	now      func() time.Time // the clock, replaceable in tests
	history  []store.Result   // finished games, newest first

	mode      mode
	game      *game.Game // the game being played, or the saved game in the menu
	cursor    int        // selected cell, 0-80
	notesMode bool
	showHelp  bool
	menuIndex int
	status    string // one-line message under the board or menu

	confirmQuit bool // the save-or-discard dialog is open
	quitIndex   int  // highlighted choice in that dialog

	tutorialStep int  // index into tutorialSteps
	sawConflict  bool // the tutorial's conflict step saw a conflict

	width, height int
	dark          bool
}

// New loads settings and any saved game, and opens the menu.
func New(st Store, rng *rand.Rand) Model {
	m := Model{store: st, rng: rng, dark: true, now: time.Now}
	m.settings, _ = st.LoadSettings() // a broken settings file just means defaults
	m.t = textsFor(m.settings.Language)
	m.toMenu()
	if !m.settings.TutorialDone {
		m.startTutorial()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, tick())
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
	case tickMsg:
		if m.mode == modePlaying && !m.showHelp && !m.confirmQuit && !m.tooSmall() {
			m.game.Tick(time.Second)
		}
		return m, tick()
	case tea.KeyPressMsg:
		return m.handleKey(msg.String())
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			return m.handleClick(m.hit(msg))
		}
	}
	return m, nil
}

func (m Model) handleKey(key string) (Model, tea.Cmd) {
	m.status = ""
	if key == "ctrl+c" {
		return m, m.quit()
	}
	if m.tooSmall() {
		if key == "q" {
			return m, m.quit() // the board is hidden, so quit without asking
		}
		return m, nil
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}
	if m.confirmQuit {
		return m.quitKey(key)
	}
	switch m.mode {
	case modeMenu:
		return m.menuKey(key)
	case modeDifficulty:
		return m.difficultyKey(key)
	case modePlaying, modeTutorial:
		return m.boardKey(key)
	case modeWon:
		return m.wonKey(key)
	}
	return m, nil
}

func (m Model) menuKey(key string) (Model, tea.Cmd) {
	items := m.menuItems()
	switch key {
	case "up", "k":
		m.menuIndex = max(m.menuIndex-1, 0)
	case "down", "j":
		m.menuIndex = min(m.menuIndex+1, len(items)-1)
	case "enter":
		return m.chooseMenu(items[m.menuIndex])
	case "q", "esc":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) menuItems() []menuItem {
	if m.game != nil {
		return []menuItem{itemContinue, itemNewGame, itemTutorial, itemLanguage, itemQuit}
	}
	return []menuItem{itemNewGame, itemTutorial, itemLanguage, itemQuit}
}

func (m Model) chooseMenu(item menuItem) (Model, tea.Cmd) {
	switch item {
	case itemContinue:
		m.mode, m.notesMode = modePlaying, false
	case itemNewGame:
		m.mode, m.menuIndex = modeDifficulty, 0
	case itemTutorial:
		m.startTutorial()
	case itemLanguage:
		m.toggleLanguage()
	case itemQuit:
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) difficultyKey(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.menuIndex = max(m.menuIndex-1, 0)
	case "down", "j":
		m.menuIndex = min(m.menuIndex+1, len(sudoku.Difficulties)-1)
	case "enter":
		m.newGame(sudoku.Difficulties[m.menuIndex])
	case "esc":
		m.toMenu()
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) boardKey(key string) (Model, tea.Cmd) {
	switch key {
	case "up", "k":
		m.move(-1, 0)
	case "down", "j":
		m.move(1, 0)
	case "left", "h":
		m.move(0, -1)
	case "right", "l":
		m.move(0, 1)
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		m.input(key[0] - '0')
	case "0", "x", "backspace", "delete":
		m.erase()
	case "n":
		m.notesMode = !m.notesMode
	case "u", "ctrl+z":
		m.undo()
	case "ctrl+r":
		m.redo()
	case "?":
		m.showHelp = true
	case "s":
		m.saveNow()
	case "enter":
		if m.mode == modeTutorial && tutorialSteps[m.tutorialStep].done == nil {
			m.tutorialStep++
		}
	case "esc":
		if m.mode == modeTutorial {
			m.finishTutorial("")
			return m, nil
		}
		m.save()
		m.toMenu()
	case "q":
		if m.mode == modeTutorial {
			return m, m.quit()
		}
		m.confirmQuit, m.quitIndex = true, quitSave
	}
	m.advanceTutorial()
	return m, nil
}

func (m Model) quitKey(key string) (Model, tea.Cmd) {
	switch key {
	case "s":
		return m.chooseQuit(quitSave)
	case "d":
		return m.chooseQuit(quitDiscard)
	case "esc", "c":
		return m.chooseQuit(quitCancel)
	case "up", "k", "left", "h":
		m.quitIndex = max(m.quitIndex-1, quitSave)
	case "down", "j", "right", "l":
		m.quitIndex = min(m.quitIndex+1, quitCancel)
	case "enter":
		return m.chooseQuit(m.quitIndex)
	}
	return m, nil
}

// chooseQuit closes the quit dialog: save and quit, discard the game and
// quit, or keep playing.
func (m Model) chooseQuit(choice int) (Model, tea.Cmd) {
	m.confirmQuit = false
	switch choice {
	case quitSave:
		m.save()
		return m, tea.Quit
	case quitDiscard:
		_ = m.store.DeleteGame() // quitting anyway; a leftover save only offers Continue
		return m, tea.Quit
	}
	return m, nil
}

// wonKey handles the win screen: menuIndex 0 is New game, 1 is Menu.
func (m Model) wonKey(key string) (Model, tea.Cmd) {
	switch key {
	case "left", "up", "h", "k":
		m.menuIndex = 0
	case "right", "down", "l", "j":
		m.menuIndex = 1
	case "enter":
		if m.menuIndex == 1 {
			m.toMenu()
		} else {
			m.mode, m.menuIndex = modeDifficulty, 0
		}
	case "n":
		m.mode, m.menuIndex = modeDifficulty, 0
	case "esc", "m":
		m.toMenu()
	case "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) move(dRow, dCol int) {
	row := min(max(sudoku.Row(m.cursor)+dRow, 0), 8)
	col := min(max(sudoku.Col(m.cursor)+dCol, 0), 8)
	m.cursor = row*9 + col
}

func (m *Model) input(v uint8) {
	if m.notesMode {
		m.game.ToggleNote(m.cursor, v)
	} else {
		m.game.Set(m.cursor, v)
	}
	m.afterMove()
}

func (m *Model) erase() { m.game.Clear(m.cursor); m.afterMove() }
func (m *Model) undo()  { m.game.Undo(); m.afterMove() }
func (m *Model) redo()  { m.game.Redo(); m.afterMove() }

// afterMove saves the game, or ends it when the board is solved.
func (m *Model) afterMove() {
	if m.mode != modePlaying {
		return
	}
	if m.game.Won() {
		m.mode, m.menuIndex = modeWon, 0
		result := store.Result{Difficulty: m.game.Difficulty, Elapsed: m.game.Elapsed, FinishedAt: m.now()}
		if err := m.store.AddResult(result); err != nil {
			m.status = m.t.SaveFailed
		}
		if err := m.store.DeleteGame(); err != nil {
			m.status = m.t.DeleteFailed
		}
		return
	}
	m.save()
}

func (m *Model) newGame(d sudoku.Difficulty) {
	m.game = game.New(d, m.rng)
	m.mode, m.cursor, m.notesMode = modePlaying, 40, false
	m.save()
}

func (m *Model) save() {
	if err := m.store.SaveGame(m.game); err != nil {
		m.status = m.t.SaveFailed
	}
}

// saveNow saves on request and says so; play also saves after every move.
func (m *Model) saveNow() {
	if m.mode != modePlaying {
		return
	}
	m.status = m.t.Saved
	m.save() // replaces the status on failure
}

func (m *Model) quit() tea.Cmd {
	if m.mode == modePlaying {
		m.save()
	}
	return tea.Quit
}

// toggleLanguage switches between English and Portuguese and saves the choice.
func (m *Model) toggleLanguage() {
	m.settings.Language = langPortuguese
	if m.t == &portuguese {
		m.settings.Language = langEnglish
	}
	m.t = textsFor(m.settings.Language)
	if err := m.store.SaveSettings(m.settings); err != nil {
		m.status = m.t.SaveFailed
	}
}

// toMenu opens the main menu, loading the saved game so it can be continued
// and the finished games to list.
func (m *Model) toMenu() {
	m.mode, m.menuIndex, m.notesMode, m.confirmQuit = modeMenu, 0, false, false
	m.history, _ = m.store.LoadHistory() // a broken history just shows nothing
	g, err := m.store.LoadGame()
	if err != nil {
		m.status = m.t.LoadFailed
	}
	if g != nil && g.Won() {
		g = nil // a solved save is left over from a crash; nothing to continue
	}
	m.game = g
}
