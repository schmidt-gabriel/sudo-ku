// Command sudo-ku is a sudoku game for the terminal.
package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/schmidt-gabriel/sudo-ku/internal/store"
	"github.com/schmidt-gabriel/sudo-ku/internal/ui"
)

// version is set at build time by GoReleaser.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("sudo-ku", version)
		return
	}

	st, err := store.Default()
	if err != nil {
		fail(err)
	}
	rng := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	if _, err := tea.NewProgram(ui.New(st, rng)).Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "sudo-ku:", err)
	os.Exit(1)
}
