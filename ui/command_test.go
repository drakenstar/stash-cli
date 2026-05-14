package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestCommandInputHistoryRecordsCommandModeOnly(t *testing.T) {
	m := NewCommandInput()

	m.Focus(":")
	m.text.SetValue("filter updated=>2024-01-01")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	m.Focus("/")
	m.text.SetValue("search text")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	require.Equal(t, []string{"filter updated=>2024-01-01"}, m.history)
}

func TestCommandInputHistoryNavigate(t *testing.T) {
	m := NewCommandInput()
	m.history = []string{"first", "second", "third"}

	m.Focus(":")
	m.text.SetValue("draft")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "third", m.text.Value())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "second", m.text.Value())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "third", m.text.Value())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "draft", m.text.Value())
}

func TestCommandInputHistorySkipsEmptyCommands(t *testing.T) {
	m := NewCommandInput()

	m.Focus(":")
	m.text.SetValue("   ")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	require.Empty(t, m.history)
}
