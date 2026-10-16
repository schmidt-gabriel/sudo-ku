package ui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
	"github.com/schmidt-gabriel/sudo-ku/internal/store"
	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

var (
	finished = time.Date(2026, 10, 16, 14, 32, 0, 0, time.Local)
	errSave  = errors.New("disk full")
)

func TestWinRecordsResult(t *testing.T) {
	fs := tutorialSeen()
	m := playing(fs)
	m.now = func() time.Time { return finished }
	almost := testSolution
	almost[2] = 0
	m.game = game.FromPuzzle(almost, testSolution, sudoku.Medium)
	m.game.Elapsed = 252 * time.Second

	m = press(m, "4")
	want := store.Result{Difficulty: sudoku.Medium, Elapsed: 252 * time.Second, FinishedAt: finished}
	if len(fs.history) != 1 || fs.history[0] != want {
		t.Fatalf("history = %+v, want [%+v]", fs.history, want)
	}
	m = press(m, "esc")
	if !strings.Contains(ansi.Strip(m.View().Content), "04:12") {
		t.Error("the menu does not show the game just finished")
	}
}

func TestTutorialWinIsNotRecorded(t *testing.T) {
	fs := &fakeStore{}
	m := newModel(fs)
	for _, i := range tutorialBlanks {
		m, _ = m.handleClick(fmt.Sprintf("cell-%d", i))
		m = press(m, fmt.Sprint(tutorialSolution[i]))
	}
	if len(fs.history) != 0 {
		t.Errorf("tutorial win recorded: %+v", fs.history)
	}
}

func TestMenuShowsFiveRecentGames(t *testing.T) {
	fs := tutorialSeen()
	for n := range 6 {
		fs.history = append(fs.history, store.Result{
			Difficulty: sudoku.Medium,
			Elapsed:    252 * time.Second,
			FinishedAt: finished.AddDate(0, 0, -n),
		})
	}
	m := newModel(fs)
	m.width, m.height = 100, 50
	got := ansi.Strip(m.View().Content)
	for _, want := range []string{"Recent games", "Oct 16 14:32", "Medium", "04:12", "Oct 12 14:32"} {
		if !strings.Contains(got, want) {
			t.Errorf("menu should contain %q:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Oct 11") {
		t.Errorf("menu shows more than five games:\n%s", got)
	}
	m.t = &portuguese
	got = ansi.Strip(m.View().Content)
	for _, want := range []string{"Jogos recentes", "16/10 14:32", "Médio"} {
		if !strings.Contains(got, want) {
			t.Errorf("Portuguese menu should contain %q:\n%s", want, got)
		}
	}
}

func TestMenuWithoutHistory(t *testing.T) {
	m := newModel(tutorialSeen())
	if strings.Contains(m.View().Content, "Recent games") {
		t.Error("an empty history should not show a section")
	}
}

func quitCmd(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("no command, want quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatal("command is not quit")
	}
}

func TestQuitAsksToSaveOrDiscard(t *testing.T) {
	fs := tutorialSeen()
	m := press(playing(fs), "4")
	next, cmd := m.Update(key("q"))
	m = next.(Model)
	if !m.confirmQuit || cmd != nil {
		t.Fatal("q should open the quit dialog, not quit")
	}
	if m = press(m, "5"); m.game.Values[2] != 4 || !m.confirmQuit {
		t.Fatal("keys in the dialog must not reach the board")
	}
	_, cmd = m.Update(key("s"))
	quitCmd(t, cmd)
	if fs.saved != m.game {
		t.Error("save and quit did not save")
	}

	fs = tutorialSeen()
	m = press(playing(fs), "4", "q")
	_, cmd = m.Update(key("d"))
	quitCmd(t, cmd)
	if fs.saved != nil {
		t.Error("discard and quit kept the save")
	}

	m = press(playing(tutorialSeen()), "q", "esc")
	if m.confirmQuit || m.mode != modePlaying {
		t.Error("esc should cancel the dialog and keep playing")
	}
}

func TestQuitDialogByMouseAndArrows(t *testing.T) {
	fs := tutorialSeen()
	m := press(playing(fs), "4")
	m, _ = m.handleClick("btn-quit")
	if !m.confirmQuit {
		t.Fatal("the Quit button did not open the dialog")
	}
	m, _ = m.handleClick("quit-2")
	if m.confirmQuit {
		t.Fatal("Cancel did not close the dialog")
	}
	m, _ = m.handleClick("btn-quit")
	_, cmd := m.handleClick("quit-1")
	quitCmd(t, cmd)
	if fs.saved != nil {
		t.Error("clicking Discard kept the save")
	}

	m = press(playing(tutorialSeen()), "q", "down", "down")
	if m.quitIndex != 2 {
		t.Fatalf("arrows: option %d, want 2", m.quitIndex)
	}
	if m = press(m, "enter"); m.confirmQuit {
		t.Error("enter on Cancel did not close the dialog")
	}
}

func TestTimerPausedWhileConfirmingQuit(t *testing.T) {
	m := press(playing(tutorialSeen()), "q")
	next, _ := m.Update(tickMsg{})
	if next.(Model).game.Elapsed != 0 {
		t.Error("timer ran while the quit dialog was open")
	}
}

func TestQuitDialogAndButtonAreShown(t *testing.T) {
	m := playing(tutorialSeen())
	m.width, m.height = 100, 50
	if got := ansi.Strip(m.View().Content); !strings.Contains(got, "[Quit]") {
		t.Errorf("game screen has no Quit button:\n%s", got)
	}
	m = press(m, "q")
	got := ansi.Strip(m.View().Content)
	for _, want := range []string{"Quit the game?", "Save and quit", "Discard and quit", "Cancel"} {
		if !strings.Contains(got, want) {
			t.Errorf("quit dialog should contain %q:\n%s", want, got)
		}
	}
}

func TestSaveButtonAndKey(t *testing.T) {
	fs := tutorialSeen()
	m := playing(fs)
	m.width, m.height = 100, 50
	if got := ansi.Strip(m.View().Content); !strings.Contains(got, "[Save]") {
		t.Errorf("game screen has no Save button:\n%s", got)
	}
	m, _ = m.handleClick("btn-save")
	if fs.saves != 1 || fs.saved != m.game || m.status != english.Saved {
		t.Fatalf("Save button: saves %d, status %q", fs.saves, m.status)
	}
	m = press(m, "s")
	if fs.saves != 2 || m.status != english.Saved {
		t.Errorf("s key: saves %d, status %q", fs.saves, m.status)
	}
	fs.saveErr = errSave
	if m = press(m, "s"); m.status != english.SaveFailed {
		t.Errorf("failed save: status %q, want %q", m.status, english.SaveFailed)
	}
}
