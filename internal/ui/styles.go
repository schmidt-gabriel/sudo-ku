package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// styles holds every color and style, picked for a light or dark terminal.
type styles struct {
	text, muted, accent, conflict, onColor color.Color
	peerBg, selectedBg, targetBg           color.Color
	playerBg, playerSelBg                  color.Color
	conflictBg, conflictSelBg              color.Color

	// frame draws the box borders; border the thin lines between cells.
	frame, border, title, help, status, item, itemOn lipgloss.Style
}

func newStyles(dark bool) styles {
	ld := lipgloss.LightDark(dark)
	c := lipgloss.Color
	s := styles{
		// All board colors are xterm-256 palette entries, so they look the
		// same in 256-color and true-color terminals. A cell's background
		// tells what it is: blue for the player's digits, red for conflicts,
		// and a lighter shade of the same color when it is selected.
		text:          ld(c("235"), c("255")),
		muted:         ld(c("245"), c("244")),
		accent:        ld(c("25"), c("81")),
		conflict:      ld(c("160"), c("203")),
		onColor:       ld(c("16"), c("231")),
		peerBg:        ld(c("254"), c("237")),
		selectedBg:    ld(c("250"), c("240")),
		targetBg:      ld(c("229"), c("58")),
		playerBg:      ld(c("153"), c("24")),
		playerSelBg:   ld(c("117"), c("31")),
		conflictBg:    ld(c("217"), c("88")),
		conflictSelBg: ld(c("210"), c("124")),
	}
	s.frame = lipgloss.NewStyle().Bold(true).Foreground(ld(c("238"), c("250")))
	s.border = lipgloss.NewStyle().Foreground(ld(c("250"), c("238")))
	s.title = lipgloss.NewStyle().Bold(true).Foreground(s.accent)
	s.help = lipgloss.NewStyle().Foreground(s.muted)
	s.status = lipgloss.NewStyle().Bold(true).Foreground(s.accent)
	s.item = lipgloss.NewStyle().Foreground(s.text)
	s.itemOn = lipgloss.NewStyle().Bold(true).Foreground(s.accent)
	return s
}
