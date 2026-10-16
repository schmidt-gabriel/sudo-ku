package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// styles holds every color and style, picked for a light or dark terminal.
type styles struct {
	text, muted, accent, conflict, same, onSelected color.Color
	peerBg, selectedBg, targetBg                    color.Color

	// frame draws the box borders; border the thin lines between cells.
	frame, border, title, help, status, item, itemOn lipgloss.Style
}

func newStyles(dark bool) styles {
	ld := lipgloss.LightDark(dark)
	c := lipgloss.Color
	s := styles{
		text:     ld(c("#1f2328"), c("#e6edf3")),
		muted:    ld(c("#8c959f"), c("#6e7681")),
		accent:   ld(c("#0969da"), c("#58a6ff")),
		conflict: ld(c("#cf222e"), c("#ff7b72")),
		// Backgrounds and the same-digit color are xterm-256 palette entries,
		// so they look the same in 256-color and true-color terminals.
		same:       ld(c("136"), c("179")),
		onSelected: ld(c("16"), c("231")), // black or white on the selection
		peerBg:     ld(c("254"), c("236")),
		selectedBg: ld(c("153"), c("25")),
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
