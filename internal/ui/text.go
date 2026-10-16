package ui

// Every string the player sees lives in this file.
const (
	textTitle            = "sudo-ku"
	textContinue         = "Continue"
	textNewGame          = "New game"
	textQuit             = "Quit"
	textMenu             = "Menu"
	textChooseDifficulty = "Choose a difficulty"
	textMenuHelp         = "↑↓ move · enter select · esc back · q quit"
	textGameHelp         = "←↑↓→ move · 1-9 place · n notes · u undo · ? help · esc menu"
	textWonHelp          = "enter new game · esc menu · q quit"
	textModeLabel        = " · Mode: "
	textModeValue        = "Value"
	textModeNotes        = "Notes"
	textErase            = "Erase"
	textNotes            = "Notes"
	textUndo             = "Undo"
	textWon              = "Solved!"
	textWonDetail        = "%s puzzle in %s"
	textLoadFailed       = "Couldn't load saved game"
	textSaveFailed       = "Couldn't save game"
	textTooSmall         = "Make the terminal at least %d×%d (it is %d×%d)"
)

var helpLines = []string{
	"How to play",
	"",
	"Fill every row, column and 3×3 box with the digits 1 to 9,",
	"using each digit once per row, column and box.",
	"",
	"←↑↓→ hjkl or click    select a cell",
	"1-9 or click [n]      place a digit (a note in notes mode)",
	"0 x backspace         erase",
	"n                     toggle notes mode",
	"u ctrl+z / ctrl+r     undo / redo",
	"esc                   back to the menu (the game is saved)",
	"q                     quit (the game is saved)",
	"",
	"Press any key to close",
}
