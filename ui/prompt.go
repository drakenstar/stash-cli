package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var promptStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#7D56F4"))

func Prompt() string {
	return promptStyle.Render(fmt.Sprintf(">> "))
}
