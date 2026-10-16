package store

import (
	"errors"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "sudo-ku")) // directory does not exist yet
}

func TestDefaultUsesXDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	s, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(dir, "sudo-ku"); s.Dir() != want {
		t.Errorf("Dir() = %s, want %s", s.Dir(), want)
	}
}

func TestDefaultFallsBackToHomeConfig(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("HOME is not used for the home directory on Windows")
	}
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)
	s, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".config", "sudo-ku"); s.Dir() != want {
		t.Errorf("Dir() = %s, want %s", s.Dir(), want)
	}
}

func TestGameRoundTrip(t *testing.T) {
	s := newStore(t)
	g := game.New(sudoku.Easy, rand.New(rand.NewPCG(1, 1)))
	for i := range 81 {
		if !g.IsGiven(i) {
			g.Set(i, g.Solution[i])
			break
		}
	}
	if err := s.SaveGame(g); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadGame()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Values != g.Values || got.Givens != g.Givens {
		t.Fatal("loaded game differs from the saved one")
	}
	if _, err := os.Stat(filepath.Join(s.Dir(), saveFile+".tmp")); !errors.Is(err, fs.ErrNotExist) {
		t.Error("temporary file left behind")
	}
}

func TestLoadGameMissing(t *testing.T) {
	g, err := newStore(t).LoadGame()
	if g != nil || err != nil {
		t.Errorf("LoadGame() = %v, %v; want nil, nil", g, err)
	}
}

func TestLoadGameCorrupt(t *testing.T) {
	s := newStore(t)
	if err := s.write(saveFile, []byte("garbage")); err != nil {
		t.Fatal(err)
	}
	g, err := s.LoadGame()
	if g != nil || err == nil {
		t.Errorf("LoadGame() = %v, %v; want nil and an error", g, err)
	}
}

func TestDeleteGame(t *testing.T) {
	s := newStore(t)
	if err := s.SaveGame(game.New(sudoku.Easy, rand.New(rand.NewPCG(1, 1)))); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteGame(); err != nil {
		t.Fatal(err)
	}
	if g, _ := s.LoadGame(); g != nil {
		t.Error("game still loads after DeleteGame")
	}
	if err := s.DeleteGame(); err != nil {
		t.Errorf("deleting a missing save: %v", err)
	}
}

func TestSettingsKeepLanguage(t *testing.T) {
	s := newStore(t)
	if err := s.SaveSettings(Settings{Language: "pt-BR"}); err != nil {
		t.Fatal(err)
	}
	if st, _ := s.LoadSettings(); st.Language != "pt-BR" {
		t.Errorf("Language = %q, want pt-BR", st.Language)
	}
}

func TestSettings(t *testing.T) {
	s := newStore(t)
	st, err := s.LoadSettings()
	if err != nil || st.TutorialDone {
		t.Fatalf("missing settings: %+v, %v; want defaults", st, err)
	}
	if err := s.SaveSettings(Settings{TutorialDone: true}); err != nil {
		t.Fatal(err)
	}
	if st, _ := s.LoadSettings(); !st.TutorialDone {
		t.Error("TutorialDone not persisted")
	}
	if err := s.write(settingsFile, []byte("{oops")); err != nil {
		t.Fatal(err)
	}
	st, err = s.LoadSettings()
	if err == nil || st.TutorialDone {
		t.Errorf("corrupt settings: %+v, %v; want defaults and an error", st, err)
	}
}
