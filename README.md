# sudo-ku

Sudoku in your terminal. Play with the mouse or the keyboard, take notes,
undo anything, and pick up where you left off.

![sudo-ku demo](demo.gif)

## Install

With Homebrew:

```sh
brew install schmidt-gabriel/tap/sudo-ku
```

With Go 1.27 or newer:

```sh
go install github.com/schmidt-gabriel/sudo-ku/cmd/sudo-ku@latest
```

Or download a binary from the [releases page](https://github.com/schmidt-gabriel/sudo-ku/releases).

## Play

```sh
sudo-ku
```

The first run opens a short interactive tutorial. You can replay it from the menu.
The game starts in English; pick **Language** in the menu to switch to
Brazilian Portuguese (and back). The choice is remembered.
The terminal needs to be at least 73 columns by 42 rows.

| Action | Keyboard | Mouse |
|---|---|---|
| Select a cell | arrows or `hjkl` | click the cell |
| Place a digit | `1`-`9` | click `[1]`-`[9]` |
| Erase | `0`, `x`, Backspace | click `[Erase]` |
| Notes mode | `n` | click `[Notes]` |
| Undo / redo | `u` / `ctrl+r` | click `[Undo]` |
| Help | `?` | |
| Menu / quit | `esc` / `q` | |

Placing a digit removes it from the notes in the same row, column and box.
Repeated digits turn red. Your game is saved after every move in
`~/.config/sudo-ku` (or `$XDG_CONFIG_HOME/sudo-ku`).

## Develop

```sh
go test ./...
go run ./cmd/sudo-ku
```

The code is split so the game logic never depends on the terminal:

- `internal/sudoku`: rules, solver and puzzle generator
- `internal/game`: one game's moves, notes, undo and timer
- `internal/store`: save files
- `internal/ui`: the Bubble Tea interface

Releases are built by GoReleaser when a `v*` tag is pushed.

## License

[MIT](LICENSE)
