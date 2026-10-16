// Package store saves the game in progress and the user's settings as JSON
// files in $XDG_CONFIG_HOME/sudo-ku, or ~/.config/sudo-ku when it is unset.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/schmidt-gabriel/sudo-ku/internal/game"
)

const (
	saveFile     = "save.json"
	settingsFile = "settings.json"
)

// Settings are the user's preferences.
type Settings struct {
	TutorialDone bool   `json:"tutorialDone"`
	Language     string `json:"language,omitempty"` // "en" or "pt-BR"; empty means English
}

// Store reads and writes files in one directory.
type Store struct {
	dir string
}

// New returns a Store that keeps its files in dir.
func New(dir string) *Store { return &Store{dir: dir} }

// Default returns the Store in $XDG_CONFIG_HOME/sudo-ku or ~/.config/sudo-ku.
func Default() (*Store, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		base = filepath.Join(home, ".config")
	}
	return New(filepath.Join(base, "sudo-ku")), nil
}

// Dir returns the directory the Store writes to.
func (s *Store) Dir() string { return s.dir }

// LoadGame returns the saved game, or nil if there is none.
func (s *Store) LoadGame() (*game.Game, error) {
	b, err := s.read(saveFile)
	if b == nil || err != nil {
		return nil, err
	}
	var g game.Game
	if err := json.Unmarshal(b, &g); err != nil {
		return nil, fmt.Errorf("store: %s: %w", saveFile, err)
	}
	return &g, nil
}

// SaveGame replaces the saved game.
func (s *Store) SaveGame(g *game.Game) error {
	b, err := json.Marshal(g)
	if err != nil {
		return err
	}
	return s.write(saveFile, b)
}

// DeleteGame removes the saved game. A missing save is not an error.
func (s *Store) DeleteGame() error {
	err := os.Remove(filepath.Join(s.dir, saveFile))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

// LoadSettings returns the saved settings, or defaults if there are none.
// A corrupt file returns defaults together with the error.
func (s *Store) LoadSettings() (Settings, error) {
	var st Settings
	b, err := s.read(settingsFile)
	if b == nil || err != nil {
		return st, err
	}
	if err := json.Unmarshal(b, &st); err != nil {
		return Settings{}, fmt.Errorf("store: %s: %w", settingsFile, err)
	}
	return st, nil
}

// SaveSettings replaces the saved settings.
func (s *Store) SaveSettings(st Settings) error {
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return s.write(settingsFile, b)
}

// read returns nil and no error when the file does not exist.
func (s *Store) read(name string) ([]byte, error) {
	b, err := os.ReadFile(filepath.Join(s.dir, name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	return b, err
}

// write replaces a file atomically: the data goes to a temporary file in the
// same directory, is synced to disk, and only then renamed over the old file,
// so a crash or power loss keeps either the old or the new contents. The
// temporary file is removed if any step fails.
func (s *Store) write(name string, b []byte) (err error) {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(s.dir, name+".*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.Remove(f.Name())
		}
	}()
	if _, err := f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Chmod(f.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(f.Name(), filepath.Join(s.dir, name))
}
