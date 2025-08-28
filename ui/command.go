package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandInputMsg is emitted when an input is confirmed.
type CommandExecMsg struct {
	Command string
}

// CommandExitMsg is emitted when the user opts to exit the command without executing.
type CommandExitMsg struct{}

// CommandInput is a single-line text input that allows a user to enter a command.  Upon pressing return, the command
// is returned as a tea.Cmd from Update.
type CommandInput struct {
	text textinput.Model
}

// NewCommandInput returns a newly initialised CommandInput model.  Takes a function to run when a command is entered
// and a prompt value to have at the start of the input.
func NewCommandInput(prompt string) CommandInput {
	m := CommandInput{
		text: textinput.New(),
	}
	m.text.Prompt = prompt
	return m
}

func (m *CommandInput) Init() tea.Cmd {
	return textinput.Blink
}

func (m *CommandInput) Focus() tea.Cmd {
	return m.text.Focus()
}

func (m CommandInput) Update(msg tea.Msg) (CommandInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.text.SetValue("")
			m.text.Blur()
			return m, func() tea.Msg { return CommandExitMsg{} }

		case tea.KeyBackspace:
			if m.text.Value() == "" {
				m.text.Blur()
				return m, func() tea.Msg { return CommandExitMsg{} }
			}

		case tea.KeyEnter:
			command := m.text.Value()
			m.text.SetValue("")
			m.text.Blur()
			return m, func() tea.Msg { return CommandExecMsg{command} }
		}
	}

	var cmd tea.Cmd
	m.text, cmd = m.text.Update(msg)
	return m, cmd
}

func (m CommandInput) View() string {
	return m.text.View()
}
