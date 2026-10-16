package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// styles holds every color and style, picked for a light or dark terminal.
type styles struct {
	text, muted, accent, conflict color.Color
	peerBg, selectedBg, targetBg  color.Color

	// frame draws the box borders; border the thin lines between cells.
	frame, border, title, help, status, item, itemOn lipgloss.Style
}

func newStyles(dark bool) styles {
	ld := lipgloss.LightDark(dark)
	c := lipgloss.Color
	s := styles{
		// All board colors are xterm-256 palette entries, so they look the
		// same in 256-color and true-color terminals. Text colors keep their
		// meaning on every background: white clues, blue player digits.
		text:       ld(c("235"), c("255")),
		muted:      ld(c("245"), c("244")),
		accent:     ld(c("25"), c("81")),
		conflict:   ld(c("160"), c("203")),
		peerBg:     ld(c("254"), c("237")),
		selectedBg: ld(c("250"), c("240")),
		targetBg:   ld(c("229"), c("58")),
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
