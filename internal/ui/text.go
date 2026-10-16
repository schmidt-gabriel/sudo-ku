package ui

// texts holds every string the player sees, in one language.
type texts struct {
	Title, Continue, NewGame, Tutorial, Quit, Menu string
	// Language is the menu item that switches to the other language.
	Language         string
	ChooseDifficulty string
	Difficulties     [4]string // indexed by sudoku.Difficulty

	MenuHelp, GameHelp, WonHelp string
	ModeLabel, ModeValue        string
	ModeNotes                   string
	Erase, Notes, Undo          string

	Won        string
	WonDetail  string // difficulty, time
	LoadFailed string
	SaveFailed string
	TooSmall   string // min width, min height, width, height

	TutorialSkip  string
	TutorialDone  string
	TutorialSteps [6]string // one per tutorialSteps entry
	Help          []string  // the first line is the title
}

const (
	langEnglish    = "en"
	langPortuguese = "pt-BR"
)

// languages maps a settings code to its texts.
var languages = map[string]*texts{
	langEnglish:    &english,
	langPortuguese: &portuguese,
}

// textsFor returns the texts for a settings code, English if unknown.
func textsFor(code string) *texts {
	if t, ok := languages[code]; ok {
		return t
	}
	return &english
}

var english = texts{
	Title:            "sudo-ku",
	Continue:         "Continue",
	NewGame:          "New game",
	Tutorial:         "Tutorial",
	Quit:             "Quit",
	Menu:             "Menu",
	Language:         "Language: English",
	ChooseDifficulty: "Choose a difficulty",
	Difficulties:     [4]string{"Easy", "Medium", "Hard", "Expert"},

	MenuHelp:  "↑↓ move · enter select · esc back · q quit",
	GameHelp:  "←↑↓→ move · 1-9 place · n notes · u undo · ? help · esc menu",
	WonHelp:   "enter new game · esc menu · q quit",
	ModeLabel: " · Mode: ",
	ModeValue: "Value",
	ModeNotes: "Notes",
	Erase:     "Erase",
	Notes:     "Notes",
	Undo:      "Undo",

	Won:        "Solved!",
	WonDetail:  "%s puzzle in %s",
	LoadFailed: "Couldn't load saved game",
	SaveFailed: "Couldn't save game",
	TooSmall:   "Make the terminal at least %d×%d (it is %d×%d)",

	TutorialSkip: " · esc skips",
	TutorialDone: "Tutorial complete. Pick New game to play!",
	TutorialSteps: [6]string{
		"Fill every row, column and 3×3 box with 1-9, each digit once. Press enter or click to start.",
		"Move to the yellow cell with the arrow keys (or hjkl), or just click it.",
		"Its row already has 1-4 and 6-9, so only 5 fits. Type 5 or click [5] below.",
		"Not sure about a cell? Press n (or click [Notes]) and add two notes to another empty cell.",
		"Press n to leave notes mode, type a digit already in that row to see it turn red, then press u.",
		"Now fill in the remaining cells. Press ? anytime for help.",
	},
	Help: []string{
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
		"Press any key or click to close",
	},
}

var portuguese = texts{
	Title:            "sudo-ku",
	Continue:         "Continuar",
	NewGame:          "Novo jogo",
	Tutorial:         "Tutorial",
	Quit:             "Sair",
	Menu:             "Menu",
	Language:         "Idioma: Português",
	ChooseDifficulty: "Escolha a dificuldade",
	Difficulties:     [4]string{"Fácil", "Médio", "Difícil", "Expert"},

	MenuHelp:  "↑↓ mover · enter escolher · esc voltar · q sair",
	GameHelp:  "←↑↓→ mover · 1-9 colocar · n notas · u desfazer · ? ajuda",
	WonHelp:   "enter novo jogo · esc menu · q sair",
	ModeLabel: " · Modo: ",
	ModeValue: "Valor",
	ModeNotes: "Notas",
	Erase:     "Apagar",
	Notes:     "Notas",
	Undo:      "Desf.",

	Won:        "Resolvido!",
	WonDetail:  "Nível %s resolvido em %s",
	LoadFailed: "Não foi possível carregar o jogo salvo",
	SaveFailed: "Não foi possível salvar o jogo",
	TooSmall:   "Aumente o terminal para pelo menos %d×%d (está %d×%d)",

	TutorialSkip: " · esc pula",
	TutorialDone: "Tutorial concluído. Escolha Novo jogo para jogar!",
	TutorialSteps: [6]string{
		"Preencha cada linha, coluna e caixa 3×3 com 1 a 9, sem repetir. Aperte enter ou clique para começar.",
		"Vá até a casa amarela com as setas (ou hjkl), ou clique nela.",
		"A linha já tem 1-4 e 6-9, então só cabe o 5. Digite 5 ou clique em [5] abaixo.",
		"Na dúvida? Aperte n (ou clique em [Notas]) e anote dois números em outra casa vazia.",
		"Aperte n para sair das notas, digite um número que já está na linha para vê-lo em vermelho e aperte u.",
		"Agora preencha as casas que faltam. Aperte ? para ver a ajuda a qualquer hora.",
	},
	Help: []string{
		"Como jogar",
		"",
		"Preencha cada linha, coluna e caixa 3×3 com os dígitos de 1",
		"a 9, sem repetir nenhum dígito na linha, coluna ou caixa.",
		"",
		"←↑↓→ hjkl ou clique   selecionar uma casa",
		"1-9 ou clique [n]     colocar um dígito (nota no modo notas)",
		"0 x backspace         apagar",
		"n                     ligar ou desligar o modo notas",
		"u ctrl+z / ctrl+r     desfazer / refazer",
		"esc                   voltar ao menu (o jogo fica salvo)",
		"q                     sair (o jogo fica salvo)",
		"",
		"Aperte qualquer tecla ou clique para fechar",
	},
}
