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

func TestCommandInputAutocompleteAcceptsSuggestion(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"delete", "filter", "refresh", "reset", "sort"})
	m.Focus(":")
	m.text.SetValue("re")
	m.rebuildSuggestions()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "reset", m.text.Value())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "refresh", m.text.Value())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "reset", m.text.Value())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	require.Equal(t, "re", m.text.Value())
}

func TestCommandInputAutocompleteOnlyAppliesToFirstToken(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"filter", "sort"})
	m.Focus(":")
	m.text.SetValue("filter ta")

	m.rebuildSuggestions()
	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteUpDownFallbackToHistory(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"filter"})
	m.history = []string{"sort date"}
	m.Focus(":")
	m.text.SetValue("x")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	require.Equal(t, "sort date", m.text.Value())
}

func TestCommandInputAutocompleteTabDismissesWithoutChangingInput(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m.text.SetValue("re")
	m.rebuildSuggestions()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	require.Equal(t, "reset", m.text.Value())
	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteNeedsOneCharacter(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m.text.SetValue("")
	m.rebuildSuggestions()

	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteTabAcceptsNearestSuggestionWhenNoneSelected(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m.text.SetValue("re")
	m.rebuildSuggestions()

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	require.Equal(t, "reset", m.text.Value())
	require.False(t, m.hasSuggestions())
}
