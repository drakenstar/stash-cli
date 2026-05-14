package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	suggestions  []string
	width        int
	suggestion   suggestionState
}

type suggestionState struct {
	start    int
	end      int
	draft    string
	items    []string
	selected int
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
	m.clearSuggestions()
	return m.text.Focus()
}

func (m *CommandInput) SetWidth(width int) {
	m.width = width
	m.text.Width = maxInt(width-len(m.text.Prompt), 0)
}

func (m CommandInput) Update(msg tea.Msg) (CommandInput, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if m.hasSuggestions() {
				m.suggestionUp()
				return m, nil
			}
			m.historyPrev()
			return m, nil

		case tea.KeyDown:
			if m.hasSuggestions() {
				m.suggestionDown()
				return m, nil
			}
			m.historyNext()
			return m, nil

		case tea.KeyTab:
			if m.hasSuggestions() {
				if m.suggestion.selected == -1 {
					m.suggestion.selected = len(m.suggestion.items) - 1
					m.applySuggestionPreview()
				}
				m.clearSuggestions()
				return m, nil
			}

		case tea.KeyEsc:
			m.text.SetValue("")
			m.text.Blur()
			m.resetHistoryCursor()
			m.clearSuggestions()
			return m, func() tea.Msg { return CommandExitMsg{} }

		case tea.KeyBackspace:
			if m.text.Value() == "" {
				m.text.Blur()
				m.resetHistoryCursor()
				m.clearSuggestions()
				return m, func() tea.Msg { return CommandExitMsg{} }
			}

		case tea.KeyEnter:
			command := m.text.Value()
			m.appendHistory(command)
			m.text.SetValue("")
			m.text.Blur()
			m.resetHistoryCursor()
			m.clearSuggestions()
			return m, func() tea.Msg { return CommandExecMsg{command} }
		}
	}

	var cmd tea.Cmd
	m.text, cmd = m.text.Update(msg)
	m.rebuildSuggestions()
	return m, cmd
}

func (m CommandInput) View() string {
	if !m.hasSuggestions() {
		return "\n" + m.text.View()
	}

	rows := make([]string, 0, len(m.suggestion.items)+1)
	indent := strings.Repeat(" ", len(m.text.Prompt)+(m.suggestion.start-1))
	rowWidth := maxInt(m.width-len(indent)+1, 0)
	for i, suggestion := range m.suggestion.items {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#D3D3D3")).Padding(0, 1).Width(rowWidth)
		if i == m.suggestion.selected {
			style = style.Background(lipgloss.Color("#483D8B")).Foreground(lipgloss.Color("#FFFFFF"))
		}
		rows = append(rows, indent+style.Render(suggestion))
	}
	rows = append(rows, m.text.View())
	return "\n" + strings.Join(rows, "\n")
}

func (m *CommandInput) SetSuggestions(suggestions []string) {
	m.suggestions = append([]string(nil), suggestions...)
	m.rebuildSuggestions()
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

func (m *CommandInput) rebuildSuggestions() {
	if m.text.Prompt != ":" {
		m.clearSuggestions()
		return
	}

	value := m.text.Value()
	if strings.ContainsRune(value, ' ') {
		m.clearSuggestions()
		return
	}

	prefix := strings.TrimSpace(value)
	if len(prefix) < 1 {
		m.clearSuggestions()
		return
	}

	var matches []string
	for _, suggestion := range m.suggestions {
		if prefix == "" || strings.HasPrefix(suggestion, prefix) {
			matches = append(matches, suggestion)
		}
		if len(matches) == 6 {
			break
		}
	}
	if len(matches) == 0 {
		m.clearSuggestions()
		return
	}

	m.suggestion = suggestionState{
		start:    0,
		end:      len(value),
		draft:    value,
		items:    matches,
		selected: -1,
	}
}

func (m *CommandInput) hasSuggestions() bool {
	return len(m.suggestion.items) > 0
}

func (m *CommandInput) clearSuggestions() {
	m.suggestion = suggestionState{selected: -1}
}

func (m *CommandInput) suggestionUp() {
	if !m.hasSuggestions() {
		return
	}

	if m.suggestion.selected == -1 {
		m.suggestion.selected = len(m.suggestion.items) - 1
	} else if m.suggestion.selected > 0 {
		m.suggestion.selected--
	}
	m.applySuggestionPreview()
}

func (m *CommandInput) suggestionDown() {
	if !m.hasSuggestions() {
		return
	}

	if m.suggestion.selected == -1 {
		return
	}
	if m.suggestion.selected < len(m.suggestion.items)-1 {
		m.suggestion.selected++
		m.applySuggestionPreview()
		return
	}

	m.suggestion.selected = -1
	m.replaceSuggestionText(m.suggestion.draft)
	m.text.CursorEnd()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m *CommandInput) applySuggestionPreview() {
	if !m.hasSuggestions() || m.suggestion.selected < 0 || m.suggestion.selected >= len(m.suggestion.items) {
		return
	}
	m.replaceSuggestionText(m.suggestion.items[m.suggestion.selected])
	m.text.CursorEnd()
}

func (m *CommandInput) replaceSuggestionText(replacement string) {
	value := m.text.Value()
	start := m.suggestion.start
	end := m.suggestion.end
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if start > len(value) {
		start = len(value)
	}
	if end > len(value) {
		end = len(value)
	}
	m.text.SetValue(value[:start] + replacement + value[end:])
	m.suggestion.end = start + len(replacement)
}
