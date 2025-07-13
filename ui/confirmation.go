package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmationOption struct {
	Cmd  tea.Cmd
	Text string
}

type Confirmation struct {
	Message string
	Options []ConfirmationOption

	selected int
}

var (
	ConfirmationOptionStyle = lipgloss.NewStyle().
				PaddingRight(1)

	ConfirmationSelectedStyle = ConfirmationOptionStyle.
					Bold(true).
					Foreground(lipgloss.Color("#FFFFFF"))
)

func (c Confirmation) Update(msg tea.Msg) (*Confirmation, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return &c, c.Options[c.selected].Cmd

		case tea.KeyLeft:
			c.selected = max(0, c.selected-1)

		case tea.KeyRight:
			c.selected = min(len(c.Options)-1, c.selected+1)
		}
	}
	return &c, nil
}

func (c Confirmation) View() string {
	var options []string
	for i, o := range c.Options {
		var option string
		if i == c.selected {
			option = ConfirmationSelectedStyle.Render("> " + o.Text)
		} else {
			option = ConfirmationOptionStyle.Render("  " + o.Text)
		}
		options = append(options, option)
	}
	return lipgloss.JoinVertical(0,
		c.Message,
		lipgloss.JoinHorizontal(0, options...),
	)
}
