package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// styles holds every color and style, picked for a light or dark terminal.
type styles struct {
	text, muted, accent, conflict        color.Color
	peerBg, sameBg, selectedBg, targetBg color.Color

	border, title, help, status, item, itemOn lipgloss.Style
}

func newStyles(dark bool) styles {
	ld := lipgloss.LightDark(dark)
	c := lipgloss.Color
	s := styles{
		text:       ld(c("#1f2328"), c("#e6edf3")),
		muted:      ld(c("#8c959f"), c("#6e7681")),
		accent:     ld(c("#0969da"), c("#58a6ff")),
		conflict:   ld(c("#cf222e"), c("#ff7b72")),
		peerBg:     ld(c("#eef1f4"), c("#1c2128")),
		sameBg:     ld(c("#d8e6f8"), c("#1f3a5f")),
		selectedBg: ld(c("#b6d4fe"), c("#2f5a8f")),
		targetBg:   ld(c("#fff1b8"), c("#5a4a00")),
	}
	s.border = lipgloss.NewStyle().Foreground(s.muted)
	s.title = lipgloss.NewStyle().Bold(true).Foreground(s.accent)
	s.help = lipgloss.NewStyle().Foreground(s.muted)
	s.status = lipgloss.NewStyle().Bold(true).Foreground(s.accent)
	s.item = lipgloss.NewStyle().Foreground(s.text)
	s.itemOn = lipgloss.NewStyle().Bold(true).Foreground(s.accent)
	return s
}
