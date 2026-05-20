package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestConfirmationZXNavigationAndSpaceConfirm(t *testing.T) {
	confirmed := false
	c := Confirmation{Options: []ConfirmationOption{
		{Text: "Cancel", Cmd: func() tea.Msg { return "cancel" }},
		{Text: "Delete", Cmd: func() tea.Msg {
			confirmed = true
			return "delete"
		}},
	}}

	updated, cmd := c.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	require.Nil(t, cmd)
	require.Equal(t, 1, updated.selected)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	require.Nil(t, cmd)
	require.Equal(t, 0, updated.selected)

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	require.Nil(t, cmd)
	require.Equal(t, 1, updated.selected)

	_, cmd = updated.Update(tea.KeyMsg{Type: tea.KeySpace})
	require.NotNil(t, cmd)
	require.Equal(t, "delete", cmd())
	require.True(t, confirmed)
}
