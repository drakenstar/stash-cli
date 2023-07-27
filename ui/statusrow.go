package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func StatusRow(width int, cells []string) string {
	cellStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#483D8B")).
		Padding(0, 1)
	style := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("#2B2A60")).
		Width(width)

	return style.Render(cellStyle.Render(strings.Join(cells, " - ")))
}
