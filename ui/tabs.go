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
	SeparatorLeftNormal  = "\ue0bb"
	SeparatorRightNormal = "\ue0b9"
	SeparatorLeftActive  = "\ue0ba"
	SeparatorRightActive = "\ue0be"
)

type Tab struct {
	Prefix string
	Label  string
}

func (s Tabs) Render(width int, tabs []Tab, active int) string {
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
		out.WriteString(separatorStyle.Render(SeparatorLeftActive))
	} else {
		out.WriteString(separatorStyle.Render(SeparatorLeftNormal))
	}

	for i, t := range tabs {
		baseStyle := cellStyle
		if i == active {
			baseStyle = activeStyle
		}

		cellStyle := baseStyle.Padding(0, 1)

		numberStyle := baseStyle.
			Foreground(s.NumberForeground)

		titleStyle := baseStyle.
			Foreground(s.TitleForeground)

		// If a Prefix value exists, then output it first with a different style.  The intention of this is to be used
		// as a keyboard shortcut hint.
		if t.Prefix != "" {
			out.WriteString(cellStyle.Render(lipgloss.JoinHorizontal(
				lipgloss.Top,
				numberStyle.Render(fmt.Sprintf("%d ", i+1)),
				titleStyle.Render(t.Label),
			)))
		} else {
			out.WriteString(cellStyle.Render(
				titleStyle.Render(t.Label),
			))
		}

		if i == active-1 {
			out.WriteString(separatorStyle.Render(SeparatorLeftActive))
		} else if i == active {
			out.WriteString(inverseSeparatorStyle.Render(SeparatorRightActive))
		} else if i < active {
			out.WriteString(separatorStyle.Render(SeparatorLeftNormal))
		} else {
			out.WriteString(separatorStyle.Render(SeparatorRightNormal))
		}
	}

	return style.
		Padding(0, 1).
		Width(width).
		Render(out.String())
}
