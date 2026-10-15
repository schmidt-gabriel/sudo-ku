# sudo-ku: design da v1

Data: 2026-10-15
Repo: https://github.com/schmidt-gabriel/sudo-ku

## Objetivo

Um sudoku de terminal (TUI) para jogar de verdade e publicar como projeto open
source bem acabado.

Critérios de sucesso:

- Instalar com um comando (`brew` ou `go install`), abrir e jogar sem ler documentação.
- Jogável inteiramente com mouse **ou** inteiramente com teclado (setas).
- Tutorial interativo no primeiro uso.
- Código simples: poucas camadas, arquivos pequenos de propósito único, sem abstrações especulativas.
- Motor reutilizável por outra UI no futuro (GUI, web via WASM) sem alterações.

Fora do escopo da v1: dica, verificação contra a solução, estatísticas,
seed/puzzle diário, temas, dificuldade por técnicas de resolução.

## Stack

- Go (1.27)
- Bubble Tea (loop da TUI), Lip Gloss (estilo), bubblezone (hit-testing do mouse)
- GoReleaser + GitHub Actions (release), vhs (GIF do README)

## Idioma

Código, README e textos da interface em **inglês** (público open source).
Todos os textos da UI ficam em um único arquivo (`internal/ui/text.go`), o que
deixa uma tradução futura trivial. Este documento é em português.

## Estrutura

```
cmd/sudo-ku/main.go      flags (--version), inicia o programa
internal/sudoku/         Grid, regras, solver, gerador. Sem dependência de UI.
internal/game/           estado da partida. Sem dependência de UI nem de disco.
internal/store/          caminhos, save.json, settings.json, gravação atômica
internal/ui/             tea.Model único com modos; view, teclas, mouse, tutorial
```

Regra de dependência: `ui -> game -> sudoku` e `ui -> store -> game`.
`sudoku` e `game` nunca importam Bubble Tea, Lip Gloss, `os` para arquivos ou o
pacote `store`. Uma UI nova (no mesmo módulo) reutiliza `sudoku` e `game` e
escolhe sua própria persistência. Se uma UI em outro repo precisar do motor,
`internal/sudoku` e `internal/game` são movidos para a raiz do módulo.

## `internal/sudoku`: motor

- `type Grid [81]uint8`, 0 = vazio. Índice `i = linha*9 + coluna`.
- `Peers(i int) []int`: as 20 células da mesma linha, coluna e caixa (tabela pré-calculada).
- `Conflicts(g Grid) []int`: índices de células preenchidas com valor repetido em algum peer.
- `IsComplete(g Grid) bool`: sem zeros e sem conflitos.
- `Solve(g Grid) (Grid, bool)` e `CountSolutions(g Grid, limit int) int`:
  backtracking com bitmasks de candidatos por linha/coluna/caixa, escolhendo a
  célula com menos candidatos. `CountSolutions` para ao atingir `limit`.
- `type Difficulty int` com `Easy`, `Medium`, `Hard`, `Expert`.
- `Generate(d Difficulty, rng *rand.Rand) (puzzle, solution Grid)`:
  1. Grade completa aleatória (solver com ordem de candidatos embaralhada).
  2. Remove células em ordem aleatória; se `CountSolutions(g, 2) != 1`, devolve a célula.
  3. Para ao atingir a meta de pistas ou quando nenhuma remoção é possível.

Metas de pistas: Easy 40, Medium 32, Hard 27, Expert = o mínimo que o passo 2
alcançar (tipicamente 22 a 24). A dificuldade é aproximada por número de pistas;
a assinatura de `Generate` permite trocar por classificação por técnicas sem
mudar quem a chama.

## `internal/game`: estado da partida

```go
type Game struct {
    Givens     sudoku.Grid   // pistas, imutáveis
    Values     sudoku.Grid   // pistas + jogadas
    Notes      [81]uint16    // bits 1..9 = candidatos anotados
    Solution   sudoku.Grid
    Difficulty sudoku.Difficulty
    Elapsed    time.Duration
    // histórico de undo/redo (não exportado, não serializado)
}
```

Operações (todas são no-op se a célula for pista):

- `Set(i, v)`: define o valor, zera as notas da célula e remove o bit `v` das notas dos 20 peers.
- `Clear(i)`: apaga o valor.
- `ToggleNote(i, v)`: só atua em célula vazia.
- `Undo()` / `Redo()`: snapshots de `{Values, Notes}` antes de cada ação que muda estado; limite de 200. Uma nova ação limpa a pilha de redo.
- `Conflicts() []int`, `Won() bool` (= `sudoku.IsComplete(Values)`).
- `Tick(d time.Duration)`: soma em `Elapsed`. Quem chama decide quando o tempo corre.
- `New(d, rng)`, `MarshalJSON` / `UnmarshalJSON`. O JSON não inclui o histórico. `UnmarshalJSON` valida (tamanhos, valores 0..9, pistas presentes em `Values`) e retorna erro se inválido.

## `internal/store`: persistência

- Diretório: `$XDG_CONFIG_HOME/sudo-ku/`, ou `~/.config/sudo-ku/` se a variável não estiver definida. O mesmo em macOS, Linux e Windows.
- `save.json`: partida em andamento. Gravado após cada jogada e ao sair; apagado na vitória.
- `settings.json`: `{"tutorialDone": bool}`.
- Gravação atômica: escreve `<arquivo>.tmp` e faz `os.Rename`.
- API: `LoadGame() (*game.Game, error)`, `SaveGame(*game.Game) error`, `DeleteGame() error`, `LoadSettings() (Settings, error)`, `SaveSettings(Settings) error`. Arquivo ausente não é erro (retorna `nil` / valores zero).

## `internal/ui`: interface

Um único `tea.Model` com campo `mode`: `menu`, `difficulty`, `playing`,
`tutorial`, `won`. A ajuda (`?`) é um overlay booleano sobre qualquer modo.

Arquivos: `model.go` (estado e `Update`), `view.go` (render), `keys.go`
(keymap, usado também pela barra de ajuda), `mouse.go` (IDs de zona e cliques),
`tutorial.go`, `styles.go`, `text.go`.

### Telas

- **Menu:** Continue (só se houver save), New game, Tutorial, Quit.
- **Difficulty:** Easy, Medium, Hard, Expert.
- **Playing:** tabuleiro, painel lateral (dificuldade, timer, modo valor/notas), teclado numérico clicável, barra de atalhos no rodapé.
- **Won:** tempo e dificuldade; New game, Menu.

Menus navegáveis com setas + Enter e com clique.

### Tabuleiro

- Célula de 5x3 caracteres; a mini-grade 3x3 de notas cabe nela. Bordas grossas entre caixas 3x3.
- Tamanho mínimo do terminal: 60x36. Abaixo disso, mostra apenas uma mensagem pedindo para aumentar a janela.
- Teclado numérico: `[1]`..`[9]`, `[Erase]`, `[Notes]`, `[Undo]`.

### Controles

| Ação | Teclado | Mouse |
|---|---|---|
| Mover / selecionar | setas, `hjkl` | clicar na célula |
| Número (ou nota, no modo notas) | `1`-`9` | botão `[n]` |
| Apagar | `0`, Backspace, `x` | `[Erase]` |
| Alternar modo notas | `n` | `[Notes]` |
| Desfazer / refazer | `u` / `ctrl+r` | `[Undo]` |
| Ajuda / menu / sair | `?` / `esc` / `q` | (só teclado) |

`q` e `esc` salvam a partida antes de sair ou voltar ao menu.

### Visual

Cores `lipgloss.AdaptiveColor` (terminal claro e escuro):

- Pistas: negrito neutro. Valores do jogador: azul. Conflitos: vermelho.
- Linha, coluna e caixa da seleção: fundo suave. Célula selecionada: fundo forte.
- Células com o mesmo valor da seleção: destacadas.
- Notas: cinza.

### Timer

`tea.Tick` de 1s chama `game.Tick` apenas no modo `playing` sem overlay de ajuda.

### Mouse

`tea.WithMouseCellMotion()` + bubblezone. Cada célula tem zona `cell-<i>`, cada
botão `btn-<nome>`, cada item de menu `menu-<n>`. O `Update` consulta qual zona
contém o clique.

## Tutorial

- Inicia automaticamente quando `settings.tutorialDone` é false. `esc` pula a qualquer momento. Ao concluir ou pular, `tutorialDone = true`. Disponível no menu.
- Usa um puzzle fixo quase resolvido (cerca de 6 células vazias), embutido no código. Não grava `save.json`.
- Passos são dados: `[]step{text string; done func(*Model) bool}`. O balão mostra o texto do passo atual; o passo avança quando `done` retorna true.
- Passos:
  1. Regras do sudoku (Enter continua).
  2. Mover até a célula destacada (setas ou clique).
  3. Digitar o único número possível nela (o texto explica o porquê).
  4. Ativar notas com `n` e anotar dois candidatos em outra célula.
  5. Colocar um número repetido, ver o conflito em vermelho, desfazer com `u`.
  6. Completar o puzzle; `?` mostra a ajuda a qualquer hora.

## Tratamento de erros

- `save.json` ausente: menu sem "Continue".
- `save.json` corrompido ou inválido: ignorado; menu sem "Continue"; aviso no rodapé ("Couldn't load saved game"). Nunca causa panic.
- `settings.json` corrompido: tratado como valores zero (tutorial roda de novo).
- Falha ao gravar: aviso no rodapé; o jogo continua.
- Terminal pequeno: mensagem de tamanho mínimo, sem quebrar o layout.

## Testes

- `sudoku`: `Solve` em puzzles conhecidos (incluindo um de 17 pistas); `CountSolutions` com puzzle de múltiplas soluções; `Generate` com seed fixa produz solução única e respeita a meta de pistas por dificuldade; `Conflicts` e `Peers`.
- `game`: edição de pista é no-op; auto-remoção de notas nos peers; `ToggleNote` em célula preenchida é no-op; undo/redo, incluindo desfazer um `Set` que removeu notas; nova ação limpa redo; ida e volta JSON; JSON inválido retorna erro.
- `store`: ida e volta de save e settings usando `XDG_CONFIG_HOME` em diretório temporário; arquivo ausente; arquivo corrompido.
- `ui`: `Update` com mensagens de tecla e de mouse sintéticas, verificando seleção, modo notas, transições de modo e avanço do tutorial. Sem snapshot de `View()`.

## Distribuição

- **CI** (push e PR): `go vet ./...` e `go test -race ./...`.
- **Release** em tag `v*`: GoReleaser gera binários para darwin, linux e windows (amd64 e arm64), cria a GitHub Release e publica um cask em `schmidt-gabriel/homebrew-tap` (`Casks/sudo-ku.rb`). Requer o secret `HOMEBREW_TAP_TOKEN` (token com escrita no repo do tap).
- Instalação:
  - `brew install schmidt-gabriel/tap/sudo-ku`
  - `go install github.com/schmidt-gabriel/sudo-ku/cmd/sudo-ku@latest`
- `--version` imprime a versão injetada via ldflags.
- README: GIF de demonstração (script `demo.tape` do vhs versionado), instalação, controles, licença MIT.
