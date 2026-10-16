package ui

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"slices"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
	"github.com/schmidt-gabriel/sudo-ku/internal/store"
	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

var (
	testPuzzle   = sudoku.MustParse("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	testSolution = sudoku.MustParse("534678912672195348198342567859761423426853791713924856961537284287419635345286179")
)

// fakeStore keeps everything in memory.
type fakeStore struct {
	saved    *game.Game
	settings store.Settings
	loadErr  error
	saveErr  error
	saves    int
}

func (f *fakeStore) LoadGame() (*game.Game, error) { return f.saved, f.loadErr }
func (f *fakeStore) SaveGame(g *game.Game) error {
	f.saves++
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = g
	return nil
}
func (f *fakeStore) DeleteGame() error                     { f.saved = nil; return nil }
func (f *fakeStore) LoadSettings() (store.Settings, error) { return f.settings, nil }
func (f *fakeStore) SaveSettings(s store.Settings) error   { f.settings = s; return nil }

// tutorialSeen returns a store for a player who already finished the tutorial.
func tutorialSeen() *fakeStore {
	return &fakeStore{settings: store.Settings{TutorialDone: true}}
}

func newModel(fs *fakeStore) Model { return New(fs, rand.New(rand.NewPCG(1, 2))) }

// playing returns a model in the middle of the test puzzle, on empty cell 2.
func playing(fs *fakeStore) Model {
	m := newModel(fs)
	m.game = game.FromPuzzle(testPuzzle, testSolution, sudoku.Easy)
	m.mode, m.cursor = modePlaying, 2
	return m
}

func key(k string) tea.KeyPressMsg {
	switch k {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "backspace":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "ctrl+r":
		return tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	}
	return tea.KeyPressMsg{Code: rune(k[0]), Text: k}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		next, _ := m.Update(key(k))
		m = next.(Model)
	}
	return m
}

func TestKeyHelperMatchesBubbleTea(t *testing.T) {
	for _, k := range []string{"up", "down", "left", "right", "enter", "esc", "backspace", "ctrl+r", "5", "n", "?"} {
		if got := key(k).String(); got != k {
			t.Errorf("key(%q).String() = %q", k, got)
		}
	}
}

func TestMenuWithoutSave(t *testing.T) {
	m := newModel(tutorialSeen())
	if m.mode != modeMenu {
		t.Fatalf("mode = %v, want menu", m.mode)
	}
	if got := m.menuItems(); !slices.Equal(got, []menuItem{itemNewGame, itemTutorial, itemQuit}) {
		t.Errorf("menu = %v, want New game, Tutorial, Quit", got)
	}
}

func TestMenuContinuesSavedGame(t *testing.T) {
	fs := tutorialSeen()
	fs.saved = game.FromPuzzle(testPuzzle, testSolution, sudoku.Hard)
	m := newModel(fs)
	if m.menuItems()[0] != itemContinue {
		t.Fatalf("first item = %v, want Continue", m.menuItems()[0])
	}
	m = press(m, "enter")
	if m.mode != modePlaying || m.game.Difficulty != sudoku.Hard {
		t.Errorf("mode %v difficulty %v, want playing the saved Hard game", m.mode, m.game.Difficulty)
	}
}

func TestLoadErrorShowsStatus(t *testing.T) {
	fs := tutorialSeen()
	fs.loadErr = errors.New("corrupt")
	m := newModel(fs)
	if m.status != textLoadFailed {
		t.Errorf("status = %q, want %q", m.status, textLoadFailed)
	}
	if m.menuItems()[0] == itemContinue {
		t.Error("Continue offered without a loadable save")
	}
}

func TestNewGameFlow(t *testing.T) {
	fs := tutorialSeen()
	m := press(newModel(fs), "enter")
	if m.mode != modeDifficulty {
		t.Fatalf("mode = %v, want difficulty", m.mode)
	}
	m = press(m, "down", "enter")
	if m.mode != modePlaying || m.game.Difficulty != sudoku.Medium {
		t.Fatalf("mode %v difficulty %v, want playing Medium", m.mode, m.game.Difficulty)
	}
	if m.cursor != 40 || fs.saves != 1 {
		t.Errorf("cursor %d saves %d, want 40 and 1", m.cursor, fs.saves)
	}
	if m = press(m, "esc"); m.mode != modeMenu {
		t.Errorf("esc from playing: mode %v, want menu", m.mode)
	}
}

func TestDifficultyEscGoesBack(t *testing.T) {
	m := press(newModel(tutorialSeen()), "enter", "esc")
	if m.mode != modeMenu {
		t.Errorf("mode = %v, want menu", m.mode)
	}
}

func TestCursorMovesAndClamps(t *testing.T) {
	m := playing(tutorialSeen())
	m.cursor = 0
	steps := []struct {
		key  string
		want int
	}{
		{"up", 0}, {"left", 0}, {"right", 1}, {"down", 10},
		{"l", 11}, {"j", 20}, {"h", 19}, {"k", 10},
	}
	for _, s := range steps {
		if m = press(m, s.key); m.cursor != s.want {
			t.Fatalf("after %q cursor = %d, want %d", s.key, m.cursor, s.want)
		}
	}
	m.cursor = 80
	if m = press(m, "down", "right"); m.cursor != 80 {
		t.Errorf("cursor left the board: %d", m.cursor)
	}
}

func TestDigitPlacesValueAndSaves(t *testing.T) {
	fs := tutorialSeen()
	m := press(playing(fs), "4")
	if m.game.Values[2] != 4 {
		t.Fatalf("Values[2] = %d, want 4", m.game.Values[2])
	}
	if fs.saved != m.game {
		t.Error("game not saved after a move")
	}
}

func TestGivenIsNotChanged(t *testing.T) {
	m := playing(tutorialSeen())
	m.cursor = 0
	if m = press(m, "9"); m.game.Values[0] != 5 {
		t.Errorf("given changed to %d", m.game.Values[0])
	}
}

func TestNotesMode(t *testing.T) {
	m := press(playing(tutorialSeen()), "n", "1", "3")
	if !m.game.HasNote(2, 1) || !m.game.HasNote(2, 3) || m.game.Values[2] != 0 {
		t.Fatal("notes mode should add notes, not values")
	}
	m = press(m, "n", "4")
	if m.game.Values[2] != 4 || m.game.Notes[2] != 0 {
		t.Error("value mode should place 4 and clear the notes")
	}
}

func TestEraseUndoRedo(t *testing.T) {
	m := press(playing(tutorialSeen()), "4", "backspace")
	if m.game.Values[2] != 0 {
		t.Fatal("backspace did not erase")
	}
	if m = press(m, "u"); m.game.Values[2] != 4 {
		t.Fatal("u did not undo the erase")
	}
	if m = press(m, "ctrl+r"); m.game.Values[2] != 0 {
		t.Fatal("ctrl+r did not redo the erase")
	}
	for _, k := range []string{"x", "0"} {
		if m = press(m, "4", k); m.game.Values[2] != 0 {
			t.Errorf("%q did not erase", k)
		}
	}
}

func TestWinning(t *testing.T) {
	fs := tutorialSeen()
	m := playing(fs)
	almost := testSolution
	almost[2] = 0
	m.game = game.FromPuzzle(almost, testSolution, sudoku.Easy)
	fs.saved = m.game

	m = press(m, "4")
	if m.mode != modeWon {
		t.Fatalf("mode = %v, want won", m.mode)
	}
	if fs.saved != nil {
		t.Error("save not deleted after winning")
	}
	if m = press(m, "enter"); m.mode != modeDifficulty {
		t.Errorf("enter on the win screen: mode %v, want difficulty", m.mode)
	}
}

func TestEscSavesAndContinues(t *testing.T) {
	m := press(playing(tutorialSeen()), "4", "esc")
	if m.mode != modeMenu || m.menuItems()[0] != itemContinue {
		t.Fatal("esc should return to a menu offering Continue")
	}
	if m = press(m, "enter"); m.mode != modePlaying || m.game.Values[2] != 4 {
		t.Error("Continue did not restore the game")
	}
}

func TestQuitSaves(t *testing.T) {
	fs := tutorialSeen()
	_, cmd := playing(fs).Update(key("q"))
	if cmd == nil {
		t.Fatal("q returned no command")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("q did not quit")
	}
	if fs.saves != 1 {
		t.Errorf("saves = %d, want 1", fs.saves)
	}
}

func TestHelpOverlay(t *testing.T) {
	m := press(playing(tutorialSeen()), "?")
	if !m.showHelp {
		t.Fatal("? did not open help")
	}
	m = press(m, "4")
	if m.showHelp || m.game.Values[2] != 0 {
		t.Error("the key that closes help must not reach the board")
	}
}

func TestSaveErrorShowsStatus(t *testing.T) {
	fs := tutorialSeen()
	fs.saveErr = errors.New("disk full")
	m := press(playing(fs), "4")
	if m.status != textSaveFailed {
		t.Fatalf("status = %q, want %q", m.status, textSaveFailed)
	}
	if m = press(m, "right"); m.status != "" {
		t.Error("status not cleared by the next key")
	}
}

func TestSolvedSaveIsNotOffered(t *testing.T) {
	fs := tutorialSeen()
	fs.saved = game.FromPuzzle(testSolution, testSolution, sudoku.Easy)
	if m := newModel(fs); m.menuItems()[0] == itemContinue {
		t.Error("Continue offered for an already solved game")
	}
}

func TestUndoAfterContinueIsHarmless(t *testing.T) {
	fs := tutorialSeen()
	fs.saved = game.FromPuzzle(testPuzzle, testSolution, sudoku.Easy)
	fs.saved.Set(2, 4)
	fs.saved, _ = roundTrip(fs.saved) // a loaded game has no undo history
	m := press(newModel(fs), "enter", "u", "ctrl+r")
	if m.mode != modePlaying || m.game.Values[2] != 4 {
		t.Errorf("undo with no history changed the game: mode %v cell %d", m.mode, m.game.Values[2])
	}
}

func roundTrip(g *game.Game) (*game.Game, error) {
	data, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	var out game.Game
	return &out, json.Unmarshal(data, &out)
}

func TestCtrlCQuitsFromAnyScreen(t *testing.T) {
	fs := tutorialSeen()
	_, cmd := newModel(fs).Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Error("ctrl+c in the menu did not quit")
	}
	if fs.saves != 0 {
		t.Error("ctrl+c in the menu wrote a save")
	}
}

func TestTimerRunsOnlyWhilePlaying(t *testing.T) {
	m := playing(tutorialSeen())
	tickOnce := func() {
		next, cmd := m.Update(tickMsg{})
		m = next.(Model)
		if cmd == nil {
			t.Fatal("tick did not schedule the next tick")
		}
	}
	tickOnce()
	if m.game.Elapsed != time.Second {
		t.Fatalf("Elapsed = %v, want 1s", m.game.Elapsed)
	}
	m.showHelp = true
	tickOnce()
	m.showHelp = false
	m.mode = modeMenu
	tickOnce()
	if m.game.Elapsed != time.Second {
		t.Errorf("timer ran outside play: %v", m.game.Elapsed)
	}
}
