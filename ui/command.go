package ui

import (
	"strings"

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
	text         textinput.Model
	history      []string
	historyIndex int
	historyDraft string
}

// NewCommandInput returns a newly initialised CommandInput model.  Takes a function to run when a command is entered
// and a prompt value to have at the start of the input.
func NewCommandInput() CommandInput {
	m := CommandInput{
		text:         textinput.New(),
		historyIndex: -1,
	}
	return m
}

func (m *CommandInput) Init() tea.Cmd {
	return textinput.Blink
}

// Focus focusses the element, and set's it's prompt to the given input value.
func (m *CommandInput) Focus(prompt string) tea.Cmd {
	m.text.Prompt = prompt
	m.resetHistoryCursor()
	return m.text.Focus()
}

func (m CommandInput) Update(msg tea.Msg) (CommandInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.historyPrev()
			return m, nil

		case tea.KeyDown:
			m.historyNext()
			return m, nil

		case tea.KeyEsc:
			m.text.SetValue("")
			m.text.Blur()
			m.resetHistoryCursor()
			return m, func() tea.Msg { return CommandExitMsg{} }

		case tea.KeyBackspace:
			if m.text.Value() == "" {
				m.text.Blur()
				m.resetHistoryCursor()
				return m, func() tea.Msg { return CommandExitMsg{} }
			}

		case tea.KeyEnter:
			command := m.text.Value()
			m.appendHistory(command)
			m.text.SetValue("")
			m.text.Blur()
			m.resetHistoryCursor()
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

func (m *CommandInput) appendHistory(command string) {
	if m.text.Prompt != ":" {
		return
	}
	if strings.TrimSpace(command) == "" {
		return
	}
	m.history = append(m.history, command)
}

func (m *CommandInput) historyPrev() {
	if len(m.history) == 0 {
		return
	}
	if m.historyIndex == -1 {
		m.historyDraft = m.text.Value()
		m.historyIndex = len(m.history) - 1
	} else if m.historyIndex > 0 {
		m.historyIndex--
	}
	m.text.SetValue(m.history[m.historyIndex])
	m.text.CursorEnd()
}

func (m *CommandInput) historyNext() {
	if m.historyIndex == -1 {
		return
	}
	if m.historyIndex < len(m.history)-1 {
		m.historyIndex++
		m.text.SetValue(m.history[m.historyIndex])
	} else {
		m.historyIndex = -1
		m.text.SetValue(m.historyDraft)
	}
	m.text.CursorEnd()
}

func (m *CommandInput) resetHistoryCursor() {
	m.historyIndex = -1
	m.historyDraft = ""
}
