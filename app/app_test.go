package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

type mockAppModel struct {
	name string
}

func (m mockAppModel) Init(Size) tea.Cmd {
	return func() tea.Msg {
		return m.name + "Init"
	}
}
func (m mockAppModel) Update(tea.Msg) (AppModel, tea.Cmd) {
	return m, nil
}
func (m mockAppModel) View() string {
	return m.name + "View"
}

func (m mockAppModel) TabTitle() string {
	return "Title"
}

func TestApp(t *testing.T) {
	t.Run("new panics without AppModels or commands", func(t *testing.T) {
		require.Panics(t, func() {
			New([]AppModelMapping{})
		})
		require.Panics(t, func() {
			New([]AppModelMapping{{NewFunc: func() AppModel { return mockAppModel{} }}})
		})
	})

	t.Run("new initalises state", func(t *testing.T) {
		a := New([]AppModelMapping{
			{
				NewFunc:  func() AppModel { return mockAppModel{"mock1"} },
				Commands: []string{"mock1", "m1"},
			},
			{
				NewFunc:  func() AppModel { return mockAppModel{"mock2"} },
				Commands: []string{"mock2", "m2"},
			},
		})
		require.Equal(t, map[string]int{"mock1": 0, "m1": 0, "mock2": 1, "m2": 1}, a.commandMappings)

		t.Run("init initialises active AppModel", func(t *testing.T) {
			cmds := a.Init()
			assertCmdsReturnMsg(t, cmds, "mock1Init")
		})

		t.Run("view renders active AppModel", func(t *testing.T) {
			v := a.View()
			require.Contains(t, v, "mock1View")
			require.Contains(t, v, ">")
		})

		t.Run("switch active state", func(t *testing.T) {
			_, cmd := a.Update(Command("mock2"))
			assertCmdsReturnMsg(t, cmd, "mock2Init")
		})

		t.Run("set confirmation dialog", func(t *testing.T) {
			type confirmMsg struct{}

			showConfirmation, _ := a.Update(ConfirmationMsg{
				Cmd:           func() tea.Msg { return confirmMsg{} },
				Message:       "confirmation",
				ConfirmOption: "confirm",
				CancelOption:  "cancel",
			})
			require.Contains(t, showConfirmation.View(), "confirmation")

			hideConfirmation, cmd := showConfirmation.Update(tea.KeyMsg{Type: tea.KeyEsc})
			require.NotContains(t, hideConfirmation.View(), "confirmation")
			assertNotCmdsReturnMsg(t, cmd, tea.QuitMsg{})

			cancel, cmd := showConfirmation.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assertCmdsReturnMsg(t, cmd, ConfirmationCancelMsg{})
			cancel, _ = cancel.Update(ConfirmationCancelMsg{})
			require.NotContains(t, cancel.View(), "confirmation")

			confirm, _ := showConfirmation.Update(tea.KeyMsg{Type: tea.KeyRight})
			confirm, cmd = confirm.Update(tea.KeyMsg{Type: tea.KeyEnter})
			assertCmdsReturnMsg(t, cmd, confirmMsg{})
		})

		t.Run("exit functions", func(t *testing.T) {
			_, cmd := a.Update(Command("exit"))
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
