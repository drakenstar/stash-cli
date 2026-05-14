package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
	"github.com/stretchr/testify/require"
)

func TestCtrlCQuitsInNormalMode(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok)
}

func TestCtrlCQuitsInCommandMode(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.mode = ModeCommand

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok)
}

func TestCtrlCQuitsWithConfirmationOpen(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.confirmation = &ui.Confirmation{}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok)
}

func TestCtrlCQuitsDuringPendingDelete(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.pendingDelete = &pendingDeleteState{
		tabID: 1,
		request: deleteRequestMsg{
			Entity: "scene",
			Title:  "Example",
			Path:   "/tmp/example.mp4",
		},
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok)
}
