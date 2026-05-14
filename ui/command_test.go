package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func typeText(t *testing.T, m CommandInput, text string) CommandInput {
	t.Helper()
	for _, r := range text {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

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
	m = typeText(t, m, "re")

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
	m = typeText(t, m, "filter ta")
	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteUpDownFallbackToHistory(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"filter"})
	m.history = []string{"sort date"}
	m.Focus(":")
	m = typeText(t, m, "x")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	require.Equal(t, "sort date", m.text.Value())
}

func TestCommandInputAutocompleteDoesNotReopenAfterHistoryNavigation(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"sort", "studio"})
	m.history = []string{"st"}
	m.Focus(":")
	m = typeText(t, m, "x")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "st", m.text.Value())
	require.False(t, m.hasSuggestions())

	m.RefreshSuggestions()
	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteTabDismissesWithoutChangingInput(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m = typeText(t, m, "re")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	require.Equal(t, "reset", m.text.Value())
	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteNeedsOneCharacter(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m.rebuildSuggestions()

	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteTabAcceptsNearestSuggestionWhenNoneSelected(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m = typeText(t, m, "re")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})

	require.Equal(t, "reset", m.text.Value())
	require.False(t, m.hasSuggestions())
}

func TestCommandInputAutocompleteKeepsDraftBasisWhilePreviewing(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"refresh", "reset"})
	m.Focus(":")
	m = typeText(t, m, "re")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "reset", m.text.Value())

	m.RefreshSuggestions()
	require.True(t, m.hasSuggestions())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.Equal(t, "refresh", m.text.Value())
}

func TestCommandInputAutocompleteReenablesAfterTypingPostHistory(t *testing.T) {
	m := NewCommandInput()
	m.SetSuggestions([]string{"sort", "studio"})
	m.history = []string{"st"}
	m.Focus(":")
	m = typeText(t, m, "x")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	require.False(t, m.hasSuggestions())

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	require.Equal(t, "stu", m.text.Value())
	require.True(t, m.hasSuggestions())
}
