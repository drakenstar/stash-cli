package ui

import (
	"bytes"

	"github.com/charmbracelet/lipgloss"
)

type Tabs struct {
	Background       lipgloss.Color
	CellBackground   lipgloss.Color
	ActiveBackground lipgloss.Color
}

func (s Tabs) Render(width int, tabs []string, active int) string {
	cellStyle := lipgloss.NewStyle().
		Background(s.CellBackground).
		Padding(0, 1)

	activeStyle := lipgloss.NewStyle().
		Background(s.ActiveBackground).
		Padding(0, 1)

	style := lipgloss.NewStyle().
		Background(s.Background)

	var lc bytes.Buffer
	for i, c := range tabs {
		if i == active {
			lc.WriteString(activeStyle.Render(c))
		} else {
			lc.WriteString(cellStyle.Render(c))
		}

		if i < len(tabs)-1 {
			lc.WriteString(style.Render(" "))
		}
	}

	return style.
		Padding(0, 1).
		Width(width).
		Render(lc.String())
}
