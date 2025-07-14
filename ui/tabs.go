package ui

import (
	"bytes"
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Tabs struct {
	NumberForeground lipgloss.Color
	TitleForeground  lipgloss.Color

	Background       lipgloss.Color
	ActiveBackground lipgloss.Color
}

const (
	SeparatorNormal     = "\ue0bb" // diagonal separator
	SeparatorIntoActive = "\ue0ba" // half-filled diagonal separator
)

func (s Tabs) Render(width int, tabs []string, active int) string {
	cellStyle := lipgloss.NewStyle().
		Background(s.Background)

	activeStyle := lipgloss.NewStyle().
		Background(s.ActiveBackground)

	style := lipgloss.NewStyle().
		Background(s.Background)

	separatorStyle := cellStyle.Foreground(s.ActiveBackground)
	inverseSeparatorStyle := activeStyle.Foreground(s.Background)

	var out bytes.Buffer

	if active == 0 {
		out.WriteString(separatorStyle.Render(SeparatorIntoActive))
	} else {
		out.WriteString(separatorStyle.Render(SeparatorNormal))
	}

	for i, title := range tabs {
		baseStyle := cellStyle
		if i == active {
			baseStyle = activeStyle
		}

		cellStyle := baseStyle.Padding(0, 1)

		numberStyle := baseStyle.
			Foreground(s.NumberForeground)

		titleStyle := baseStyle.
			Foreground(s.TitleForeground)

		out.WriteString(cellStyle.Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			numberStyle.Render(fmt.Sprintf("%d ", i+1)),
			titleStyle.Render(title),
		)))

		switch i {
		case active - 1:
			out.WriteString(separatorStyle.Render(SeparatorIntoActive))
		case active:
			out.WriteString(inverseSeparatorStyle.Render(SeparatorIntoActive))
		default:
			out.WriteString(separatorStyle.Render(SeparatorNormal))
		}
	}

	return style.
		Padding(0, 1).
		Width(width).
		Render(out.String())
}
