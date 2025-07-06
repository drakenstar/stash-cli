package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandExecuteMsg is sent when the user has input a command and wants to execute it.
type CommandExecuteMsg struct {
	Command string
}

// Name returns all characters up to the first encountered space in an input string.  This is to be interpretted
// as the command for the rest of the input.
func (i CommandExecuteMsg) Name() string {
	idx := strings.Index(i.Command, " ")
	if idx == -1 {
		return i.Command
	}
	return string(i.Command[:idx])
}

// ArgString returns all text after the command name.  This may be interpretted in any way an action deems appropriate.
func (i CommandExecuteMsg) ArgString() string {
	idx := strings.Index(i.Command, " ")
	if idx == -1 {
		return ""
	}
	return i.Command[idx+1:]
}

// ArgInt attempts to parse any value given after the command as an integer.
func (i CommandExecuteMsg) ArgInt() (int, error) {
	idx := strings.Index(i.Command, " ")
	if idx == -1 {
		return 0, fmt.Errorf("no argument given")
	}
	return strconv.Atoi(i.Command[idx+1:])
}

// Args returns a tokenised set of arguments that come after the initial command, not including the command itself.
// Tokens are split on space, with multiple spaces being ignored.
func (i CommandExecuteMsg) Args() []string {
	return strings.Fields(i.ArgString())
}

// CommandExitMsg is emitted when the user opts to exit the command without executing.
type CommandExitMsg struct{}

// CommandInput is a single-line text input that allows a user to enter a command.  Upon pressing return, the command
// is returned as a tea.Cmd from Update.
type CommandInput struct {
	text   textinput.Model
	prefix string
}

// NewCommandInput returns a newly initialised CommandInput model.
func NewCommandInput() CommandInput {
	m := CommandInput{
		text: textinput.New(),
	}
	m.text.Prompt = ":"
	return m
}

func (m *CommandInput) Focus(prompt, prefix string) tea.Cmd {
	m.text.Prompt = prompt
	m.prefix = prefix
	return m.text.Focus()
}

func (m *CommandInput) Blur() {
	m.text.Blur()
}

func (m CommandInput) Update(msg tea.Msg) (CommandInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			m.text.SetValue("")
			m.prefix = ""
			return m, func() tea.Msg { return CommandExitMsg{} }

		case tea.KeyEnter:
			command := m.prefix + m.text.Value()
			m.text.SetValue("")
			m.prefix = ""
			return m, func() tea.Msg {
				return CommandExecuteMsg{Command: command}
			}
		}
	}

	var cmd tea.Cmd
	m.text, cmd = m.text.Update(msg)
	return m, cmd
}

func (m CommandInput) View() string {
	return m.text.View()
}
