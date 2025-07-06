package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/ui"
	"github.com/stretchr/testify/require"
)

type mockTabModel struct {
	name string
}

func (m mockTabModel) Init(Size) tea.Cmd {
	return func() tea.Msg {
		return m.name + "Init"
	}
}
func (m mockTabModel) Update(tea.Msg) (TabModel, tea.Cmd) {
	return m, nil
}
func (m mockTabModel) View() string {
	return m.name + "View"
}

func (m mockTabModel) Title() string {
	return "Title"
}

func TestApp(t *testing.T) {
	t.Run("new panics without TabModels or commands", func(t *testing.T) {
		require.Panics(t, func() {
			New([]TabModelMapping{})
		})
		require.Panics(t, func() {
			New([]TabModelMapping{{NewFunc: func() TabModel { return mockTabModel{} }}})
		})
	})

	t.Run("new initalises state", func(t *testing.T) {
		a := New([]TabModelMapping{
			{
				NewFunc:  func() TabModel { return mockTabModel{"mock1"} },
				Commands: []string{"mock1", "m1"},
			},
			{
				NewFunc:  func() TabModel { return mockTabModel{"mock2"} },
				Commands: []string{"mock2", "m2"},
			},
		})
		require.Equal(t, map[string]int{"mock1": 0, "m1": 0, "mock2": 1, "m2": 1}, a.commandMappings)

		t.Run("init initialises active TabModel", func(t *testing.T) {
			cmds := a.Init()
			assertCmdsReturnMsg(t, cmds, "mock1Init")
		})

		t.Run("view renders active TabModel", func(t *testing.T) {
			v := a.View()
			require.Contains(t, v, "mock1View")
			require.Contains(t, v, ">")
		})

		t.Run("switch active state", func(t *testing.T) {
			_, cmd := a.Update(ui.CommandExecuteMsg{Command: "mock2"})
			assertCmdsReturnMsg(t, cmd, "mock2Init")
		})

		t.Run("exit functions", func(t *testing.T) {
			_, cmd := a.Update(ui.CommandExecuteMsg{Command: "exit"})
			assertCmdsReturnMsg(t, cmd, tea.QuitMsg{})

			_, cmd = a.Update(tea.KeyMsg{Type: tea.KeyEsc})
			assertCmdsReturnMsg(t, cmd, tea.QuitMsg{})

			_, cmd = a.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			assertCmdsReturnMsg(t, cmd, tea.QuitMsg{})
		})
	})
}

func assertCmdsReturnMsg(t *testing.T, cmd tea.Cmd, msg tea.Msg) {
	t.Helper()
	require.True(t, cmdReturnsMsg(cmd, msg))
}

func assertNotCmdsReturnMsg(t *testing.T, cmd tea.Cmd, msg tea.Msg) {
	t.Helper()
	require.False(t, cmdReturnsMsg(cmd, msg))
}

func cmdReturnsMsg(cmd tea.Cmd, msg tea.Msg) bool {
	if cmd == nil {
		return msg == nil
	}
	cmdMsg := cmd()
	if msg == cmdMsg {
		return true
	}
	if batch, ok := cmdMsg.(tea.BatchMsg); ok {
		for _, cmd := range batch {
			if cmdReturnsMsg(cmd, msg) {
				return true
			}
		}
	}
	return false
}
