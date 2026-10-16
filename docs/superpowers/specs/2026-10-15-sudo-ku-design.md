# sudo-ku: v1 design

Date: 2026-10-15
Repo: https://github.com/schmidt-gabriel/sudo-ku

## Goal

A terminal sudoku (TUI) that is pleasant to actually play and polished enough
to publish as an open source project.

Success criteria:

- Install with one command (`brew` or `go install`), open it and play without reading docs.
- Fully playable with the keyboard (arrow keys). The mouse selects cells and drives the menus.
- Interactive tutorial on first run.
- Simple code: few layers, small single-purpose files, no speculative abstractions.
- The engine can be reused by another UI later (desktop GUI, web via WASM) without changes.

Out of scope for v1: hints, checking against the solution, statistics,
seeds/daily puzzle, themes, difficulty rated by solving techniques.

## Stack

- Go 1.27
- Bubble Tea v2 (`charm.land/bubbletea/v2`), Lip Gloss v2 (`charm.land/lipgloss/v2`), bubblezone v2 (`github.com/lrstanley/bubblezone/v2`) for mouse hit-testing
- GoReleaser v2 + GitHub Actions for releases, vhs for the README GIF

## Language

Code, README, docs, commits and PR descriptions are in **English**.

The UI is available in English and Brazilian Portuguese. All UI strings live
in one file (`internal/ui/text.go`) as one `texts` value per language. The game
always starts in English (no locale detection); a "Language" menu item toggles
between the two and the choice is saved in `settings.json` as
`"language": "en" | "pt-BR"`. An unknown value falls back to English.
Difficulty names are translated by the UI; `sudoku.Difficulty.String()` stays
English.

## Layout

```
cmd/sudo-ku/main.go      flags (--version), starts the program
internal/sudoku/         Grid, rules, solver, generator. No UI dependencies.
internal/game/           one game's state. No UI or disk dependencies.
internal/store/          paths, save.json, settings.json, atomic writes
internal/ui/             one tea.Model with modes; rendering, input, mouse, tutorial
```

Dependency rule: `ui -> game -> sudoku` and `ui -> store -> game`.
`sudoku` and `game` never import Bubble Tea, Lip Gloss, file I/O or `store`.
A new UI in the same module reuses `sudoku` and `game` and picks its own
persistence. If a UI in another repo needs the engine, `internal/sudoku` and
`internal/game` move to the module root.

## `internal/sudoku`: engine

- `type Grid [81]uint8`, 0 = empty. Cell `i = row*9 + col`.
- `Row(i)`, `Col(i)`, `Box(i)`.
- `Peers(i int) []int`: the 20 cells sharing a row, column or box (precomputed).
- `Conflicts(g Grid) []int`: filled cells whose value repeats in a peer, ascending.
- `IsComplete(g Grid) bool`: no empty cells and no conflicts.
- `Parse(s string) (Grid, error)`, `MustParse`, `Grid.String()`: 81 characters, `0` or `.` for empty.
- `Solve(g Grid) (Grid, bool)` and `CountSolutions(g Grid, limit int) int`:
  backtracking with per row/column/box candidate bitmasks, always branching on
  the cell with the fewest candidates. `CountSolutions` stops at `limit`.
- `type Difficulty int`: `Easy`, `Medium`, `Hard`, `Expert`, with `String()` and `Valid()`.
- `Generate(d Difficulty, rng *rand.Rand) (puzzle, solution Grid)`:
  1. Fill a random complete grid (solver with shuffled digit order).
  2. Empty cells in random order; if `CountSolutions(g, 2) != 1`, put the cell back.
  3. Stop at the clue target or when no cell can be removed.
  4. If a pass ends above the target, retry (up to 10 passes) and keep the best.

Clue targets: Easy 40, Medium 32, Hard 27. Expert has no target: it removes
as many clues as possible and retries until it has fewer than 27. Difficulty is
approximated by clue count; `Generate`'s signature allows switching to
technique-based rating later without changing callers.

## `internal/game`: game state

```go
type Game struct {
    Givens     sudoku.Grid   // read-only clues
    Values     sudoku.Grid   // clues + player moves
    Notes      [81]uint16    // bit v set = digit v pencilled in
    Solution   sudoku.Grid
    Difficulty sudoku.Difficulty
    Elapsed    time.Duration
    // unexported undo/redo history, not serialized
}
```

Operations (no-ops on givens):

- `Set(i, v)`: places v, clears the cell's notes and removes v from the notes of its 20 peers.
- `Clear(i)`: empties the cell (value and notes).
- `ToggleNote(i, v)`: only on empty cells.
- `Undo()` / `Redo()`: snapshots of `{Values, Notes}` taken before every change; at most 200. A new change clears redo.
- `Conflicts()`, `Won()` (= `sudoku.IsComplete(Values)`), `IsGiven(i)`, `HasNote(i, v)`.
- `Tick(d)`: adds to `Elapsed`. The caller decides when time runs.
- `New(d, rng)`, `FromPuzzle(puzzle, solution, d)`.
- `MarshalJSON` / `UnmarshalJSON`. JSON excludes history. `UnmarshalJSON` rejects data that could not come from a real game (bad grids, givens overwritten, unsolved solution, invalid notes, unknown difficulty, negative time).

## `internal/store`: persistence

- Directory: `$XDG_CONFIG_HOME/sudo-ku/`, or `~/.config/sudo-ku/` when unset. Same on macOS, Linux and Windows.
- `save.json`: game in progress. Written after every move, when leaving to the menu and on quit; deleted on a win.
- `settings.json`: `{"tutorialDone": bool, "language": "en" | "pt-BR"}`.
- `history.json`: finished games, newest first, at most 100: `[{"difficulty", "elapsed", "finishedAt"}]`. Written on every win outside the tutorial; a corrupt file reads as empty and is replaced on the next win.
- Atomic writes: write `<file>.tmp`, then `os.Rename`.
- API on `*Store`: `LoadGame() (*game.Game, error)`, `SaveGame`, `DeleteGame`, `LoadSettings() (Settings, error)`, `SaveSettings`. A missing file is not an error (nil game / zero settings).

## `internal/ui`: interface

One `tea.Model` with a `mode`: `menu`, `difficulty`, `playing`, `tutorial`,
`won`. Help (`?`) is a flag that replaces the screen until any key or click.

Files: `model.go` (state, `Update`, key handling), `view.go` (screens),
`board.go` (board drawing and click-to-cell math), `mouse.go` (zones and
clicks), `styles.go`, `text.go`, `tutorial.go`.

### Screens

- **Menu:** Continue (only with a save), New game, Tutorial, Language, Quit. Below it, "Recent games": the last 5 finished games (date and time in the language's format, difficulty, elapsed time).
- **Difficulty:** Easy, Medium, Hard, Expert.
- **Playing:** header line (difficulty, timer, value/notes mode, and clickable `[Save]` `[Menu]` on the right; `[Menu]` saves and returns to the menu like `esc`), board, shortcut line, status line.
- **Quit dialog** (`q` during a game): Save and quit (`s`), Discard and quit (`d`, deletes the save), Cancel (`esc`). Arrows + enter and clicks work too; the timer pauses while it is open. In the tutorial and the menus `q` quits directly; `ctrl+c` always saves and quits.
- **Won:** difficulty and time; New game, Menu.

Lists work with arrows + enter and with clicks.

### Board

- Cells are 9x2 characters. A digit sits centered on the first line; notes 1-5 fill the first line (`1 2 3 4 5`) and 6-9 the second, shifted half a step (` 6 7 8 9 `), each digit in a fixed spot.
- Thin `│` and `─` lines between all cells in a dim color; box borders `━┃` with matching junctions (`┯`, `┿`, `╂`, ...) in a bright bold color so boxes stand out. Lines are never painted, so every highlighted cell sits inside the same even margin.
- The board is 91x28. Minimum terminal: 91x32. Below that, only a message asking for a bigger window is shown.

### Controls

| Action | Keyboard | Mouse |
|---|---|---|
| Move / select | arrows, `hjkl` | click a cell |
| Digit (or note in notes mode) | `1`-`9` | |
| Erase | `0`, `x`, Backspace, Delete | |
| Toggle notes mode | `n` | |
| Undo / redo | `u`, `ctrl+z` / `ctrl+r` | |
| Save now (play also autosaves after every move) | `s` | `[Save]` |
| Help | `?` | |
| Back to the menu (saves) | `esc` | `[Menu]` |
| Quit (asks save or discard) | `q` | |

`esc` saves the game before returning to the menu.

### Look

Colors come from `lipgloss.LightDark`, driven by `tea.BackgroundColorMsg`
(dark is assumed until the terminal answers):

- A cell's background tells what it holds: blue for the player's digits, red for conflicts (white text on both), none for clues (bold text). The selected cell uses a lighter shade of its own color (light gray for clues and empty cells). Its row, column and box get a soft gray, except blue and red cells, which keep their color.
- Digits equal to the selected one: underlined and bold.
- Notes: grey (main text color on the selected cell).
- Backgrounds and amber are xterm-256 palette colors so 256-color terminals show them unchanged.

### Timer

A 1s `tea.Tick` calls `game.Tick` only in `playing` mode while help is closed.

### Mouse

`View().MouseMode = tea.MouseModeCellMotion` plus bubblezone. The whole board is
one zone; the click position inside it maps to a cell with plain arithmetic
(`cellAt`). List items and buttons are zones (`menu-<n>`, `diff-<n>`,
`won-<name>`). Only zones of the current mode are
checked, so stale zones from a previous screen never fire.

## Tutorial

- Starts automatically when `tutorialDone` is false. `esc` skips at any time. Finishing or skipping sets `tutorialDone = true`. Available from the menu. `q` quits without marking it done.
- Uses a fixed puzzle built into the code: a known solution with 6 blanks in 6 different rows, so each blank has exactly one possible digit. It never writes `save.json` and the timer does not run.
- Steps are data: `[]tutorialStep{text string; done func(*Model) bool}`. A nil `done` waits for enter. After every key or click the tutorial skips past steps whose `done` is true.
- The header shows "Tutorial · esc skips"; the step text replaces the shortcut line (at most 2 lines). The target cell is highlighted yellow during the first steps.
- Steps:
  1. Sudoku rules (enter continues).
  2. Move to the highlighted cell (arrows or click).
  3. Type the only digit that fits (the text explains why).
  4. Turn on notes with `n` and add two notes to another empty cell.
  5. Leave notes mode, place a repeated digit, see the red conflict, undo with `u`.
  6. Fill the rest; `?` shows help anytime.

## Error handling

- `save.json` missing: menu without Continue.
- `save.json` corrupt or invalid: ignored; menu without Continue; status line says "Couldn't load saved game". Never panics.
- `settings.json` corrupt: treated as defaults (tutorial runs again).
- Write failure: status line says "Couldn't save game"; play continues.
- Terminal too small: size message instead of a broken layout.

## Testing

- `sudoku`: `Solve` on known puzzles (including a 17-clue one); `CountSolutions` on an empty grid and dead ends; `Generate` with fixed seeds gives a unique solution and hits each clue target; `Conflicts`, `Peers`, `Parse`.
- `game`: givens read-only; peer note removal; `ToggleNote` on filled cells is a no-op; undo/redo including undoing a `Set` that removed notes; new change clears redo; history bound; JSON round trip; invalid JSON is rejected.
- `store`: save/settings round trips in a temp dir; `XDG_CONFIG_HOME` and `HOME` fallback; missing and corrupt files; no `.tmp` left behind.
- `ui`: `Update` with synthetic key messages and `handleClick` with zone IDs, checking selection, notes mode, mode transitions, saving, status messages and the full tutorial walkthrough; `cellAt` math; a render smoke test per screen. No `View()` snapshots.

## Distribution

- **CI** (push to main and PRs): `go vet ./...` and `go test -race ./...`.
- **Release** on `v*` tags: GoReleaser builds darwin, linux and windows (amd64 and arm64), creates the GitHub Release and publishes a cask to `schmidt-gabriel/homebrew-tap` (`Casks/sudo-ku.rb`). Needs the `HOMEBREW_TAP_TOKEN` secret (a token with write access to the tap repo).
- Install:
  - `brew install schmidt-gabriel/tap/sudo-ku`
  - `go install github.com/schmidt-gabriel/sudo-ku/cmd/sudo-ku@latest`
- `--version` prints the version injected with ldflags.
- README: demo GIF (from a versioned vhs `demo.tape`), install, controls, MIT license.
