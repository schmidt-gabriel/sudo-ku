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
}

type menuItem int

const (
	itemContinue menuItem = iota
	itemNewGame
	itemQuit
)

type tickMsg struct{}

// Model is the whole UI state.
type Model struct {
	store    Store
	rng      *rand.Rand
	settings store.Settings

	mode      mode
	game      *game.Game // the game being played, or the saved game in the menu
	cursor    int        // selected cell, 0-80
	notesMode bool
	showHelp  bool
	menuIndex int
	status    string // one-line message under the board or menu

	width, height int
	dark          bool
}

// New loads settings and any saved game, and opens the menu.
func New(st Store, rng *rand.Rand) Model {
	m := Model{store: st, rng: rng, dark: true}
	m.settings, _ = st.LoadSettings() // a broken settings file just means defaults
	m.toMenu()
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
		if m.mode == modePlaying && !m.showHelp {
			m.game.Tick(time.Second)
		}
		return m, tick()
	case tea.KeyPressMsg:
		return m.handleKey(msg.String())
	}
	return m, nil
}

func (m Model) handleKey(key string) (Model, tea.Cmd) {
	m.status = ""
	if key == "ctrl+c" {
		return m, m.quit()
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
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
		return []menuItem{itemContinue, itemNewGame, itemQuit}
	}
	return []menuItem{itemNewGame, itemQuit}
}

func (m Model) chooseMenu(item menuItem) (Model, tea.Cmd) {
	switch item {
	case itemContinue:
		m.mode, m.notesMode = modePlaying, false
	case itemNewGame:
		m.mode, m.menuIndex = modeDifficulty, 0
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
	case "esc":
		m.save()
		m.toMenu()
	case "q":
		return m, m.quit()
	}
	return m, nil
}

func (m Model) wonKey(key string) (Model, tea.Cmd) {
	switch key {
	case "enter", "n":
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
		m.mode = modeWon
		if err := m.store.DeleteGame(); err != nil {
			m.status = textSaveFailed
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
		m.status = textSaveFailed
	}
}

func (m *Model) quit() tea.Cmd {
	if m.mode == modePlaying {
		m.save()
	}
	return tea.Quit
}

// toMenu opens the main menu, loading the saved game so it can be continued.
func (m *Model) toMenu() {
	m.mode, m.menuIndex, m.notesMode = modeMenu, 0, false
	g, err := m.store.LoadGame()
	if err != nil {
		m.status = textLoadFailed
	}
	if g != nil && g.Won() {
		g = nil // a solved save is left over from a crash; nothing to continue
	}
	m.game = g
}
