# sudo-ku

Sudoku in your terminal. Play with the keyboard (click to pick a cell), take notes,
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
The terminal needs to be at least 91 columns by 32 rows.

| Action | Keys |
|---|---|
| Select a cell | arrows, `hjkl`, or click the cell |
| Place a digit | `1`-`9` |
| Erase | `0`, `x`, Backspace |
| Notes mode | `n` |
| Undo / redo | `u` / `ctrl+r` |
| Save now | `s`, or click `[Save]` |
| Help | `?` |
| Menu | `esc` (saves) |
| Quit | `q`, or click `[Quit]`: save and quit, or discard and quit |

Placing a digit removes it from the notes in the same row, column and box.
Blue digits are yours and can be changed; the others are clues. Repeated
digits turn red. Your game is saved after every move in `~/.config/sudo-ku`
(or `$XDG_CONFIG_HOME/sudo-ku`), and the menu lists your last finished
games with their difficulty and time.

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
