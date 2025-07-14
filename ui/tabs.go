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

	var lc bytes.Buffer

	if active == 0 {
		lc.WriteString(separatorStyle.Render(SeparatorIntoActive))
	} else {
		lc.WriteString(separatorStyle.Render(SeparatorNormal))
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

		lc.WriteString(cellStyle.Render(lipgloss.JoinHorizontal(
			lipgloss.Top,
			numberStyle.Render(fmt.Sprintf("%d ", i+1)),
			titleStyle.Render(title),
		)))

		switch i {
		case active - 1:
			lc.WriteString(separatorStyle.Render(SeparatorIntoActive))
		case active:
			lc.WriteString(inverseSeparatorStyle.Render(SeparatorIntoActive))
		default:
			lc.WriteString(separatorStyle.Render(SeparatorNormal))
		}
	}

	return style.
		Padding(0, 1).
		Width(width).
		Render(lc.String())
}
