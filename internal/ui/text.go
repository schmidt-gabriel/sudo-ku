package ui

// Every string the player sees lives in this file.
const (
	textTitle            = "sudo-ku"
	textContinue         = "Continue"
	textNewGame          = "New game"
	textTutorial         = "Tutorial"
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

	textTutorialSkip     = " · esc skips"
	textTutorialDone     = "Tutorial complete. Pick New game to play!"
	textTutorialRules    = "Fill every row, column and 3×3 box with 1-9, each digit once. Press enter to start."
	textTutorialMove     = "Move to the yellow cell with the arrow keys (or hjkl), or just click it."
	textTutorialPlace    = "Its row already has 1-4 and 6-9, so only 5 fits. Type 5 or click [5] below."
	textTutorialNotes    = "Not sure about a cell? Press n (or click [Notes]) and add two notes to another empty cell."
	textTutorialConflict = "Press n to leave notes mode, type a digit already in that row to see it turn red, then press u."
	textTutorialFinish   = "Now fill in the remaining cells. Press ? anytime for help."
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
