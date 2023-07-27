package ui

import (
	"bytes"

	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	Background     lipgloss.Color
	CellBackground lipgloss.Color
}

func (s StatusBar) Render(width int, leftCells []string, rightCells []string) string {
	cellStyle := lipgloss.NewStyle().
		Background(s.CellBackground).
		Padding(0, 1)

	style := lipgloss.NewStyle().
		Background(s.Background)

	var lc bytes.Buffer
	for i, c := range leftCells {
		lc.WriteString(cellStyle.Render(c))
		if i < len(leftCells)-1 {
			lc.WriteString(style.Render(" "))
		}
	}

	var rc bytes.Buffer
	for i, c := range rightCells {
		rc.WriteString(cellStyle.Render(c))
		if i < len(rightCells)-1 {
			rc.WriteString(style.Render(" "))
		}
	}

	midWidth := width - lipgloss.Width(lc.String()) - lipgloss.Width(rc.String()) - 2
	var mb bytes.Buffer
	for i := 0; i < midWidth; i++ {
		mb.WriteRune(' ')
	}
	lc.WriteString(style.Render(mb.String()))
	lc.WriteString(rc.String())

	return style.
		Padding(0, 1).
		Width(width).
		Render(lc.String())
}
