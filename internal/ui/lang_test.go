package ui

import (
	"reflect"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/sudoku"
)

func TestEveryLanguageHasEveryText(t *testing.T) {
	for code, tx := range languages {
		v := reflect.ValueOf(*tx)
		for i := range v.NumField() {
			f, name := v.Field(i), v.Type().Field(i).Name
			switch f.Kind() {
			case reflect.String:
				if f.String() == "" {
					t.Errorf("%s: %s is empty", code, name)
				}
			case reflect.Array, reflect.Slice:
				if f.Len() == 0 {
					t.Errorf("%s: %s is empty", code, name)
				}
				for j := range f.Len() {
					if f.Index(j).Kind() == reflect.String && f.Index(j).String() == "" && name != "Help" {
						t.Errorf("%s: %s[%d] is empty", code, name, j)
					}
				}
			}
		}
	}
}

func TestTextsFitTheMinimumWidth(t *testing.T) {
	st := newStyles(true)
	for code, tx := range languages {
		m := playing(tutorialSeen())
		m.t = tx
		check := func(what, s string) {
			t.Helper()
			for _, line := range strings.Split(s, "\n") {
				if w := lipgloss.Width(line); w > minWidth {
					t.Errorf("%s: %s is %d wide, want at most %d: %q", code, what, w, minWidth, line)
				}
			}
		}
		check("number pad", m.renderPad(st))
		check("shortcut line", m.footer(st))
		check("header", m.header(st))
		for _, line := range tx.Help {
			check("help", line)
		}
		m.mode = modeTutorial
		for n := range tutorialSteps {
			m.tutorialStep = n
			footer := m.footer(st)
			check("tutorial step", footer)
			if lines := strings.Count(footer, "\n") + 1; lines > 2 {
				t.Errorf("%s: tutorial step %d takes %d lines, want at most 2", code, n+1, lines)
			}
		}
	}
}

func TestStartsInEnglish(t *testing.T) {
	m := newModel(tutorialSeen())
	if m.t != &english {
		t.Fatal("a new player should start in English")
	}
	if !strings.Contains(m.View().Content, "New game") {
		t.Error("menu is not in English")
	}
}

func TestLanguageMenuTogglesAndSaves(t *testing.T) {
	fs := tutorialSeen()
	m := press(newModel(fs), "down", "down", "enter") // New game, Tutorial, Language
	if m.mode != modeMenu || m.t != &portuguese || fs.settings.Language != "pt-BR" {
		t.Fatalf("mode %v, saved language %q: want the menu in Portuguese", m.mode, fs.settings.Language)
	}
	if !strings.Contains(m.View().Content, "Novo jogo") {
		t.Error("menu was not redrawn in Portuguese")
	}
	m = press(m, "enter")
	if m.t != &english || fs.settings.Language != "en" {
		t.Errorf("second toggle: saved language %q, want English", fs.settings.Language)
	}
}

func TestSavedLanguageIsUsed(t *testing.T) {
	fs := tutorialSeen()
	fs.settings.Language = "pt-BR"
	if m := newModel(fs); m.t != &portuguese {
		t.Error("saved pt-BR setting ignored")
	}
	fs.settings.Language = "klingon"
	if m := newModel(fs); m.t != &english {
		t.Error("unknown language should fall back to English")
	}
}

func TestDifficultyNamesAreTranslated(t *testing.T) {
	m := playing(tutorialSeen())
	m.t = &portuguese
	m.game.Difficulty = sudoku.Hard
	if h := m.header(newStyles(true)); !strings.Contains(h, "Difícil") {
		t.Errorf("header %q does not name the difficulty in Portuguese", h)
	}
}
