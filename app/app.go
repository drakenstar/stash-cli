package app

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/ui"
)

// TabModel is a subview of the App application, and operates similarly to a tea.Model.
type TabModel interface {
	// Provided the size of the current area the TabModel is to render within. Should return any initialisation
	// commands to execute.
	Init(Size) tea.Cmd
	// Handle any tea.Msg that the app.Model passes on.  Should by a immutable operation and return a new TabModel
	// with the new state after handling the message.
	Update(tea.Msg) (TabModel, tea.Cmd)
	// Normal tea.Model:View method, should render the current state of the view as a string.
	View() string
	// Returns a string name to be used for the tab title
	Title() string
}

const (
	ModeNormal = iota
	ModeCommand
)

// TabModelMapping maps a given TabModel to the commands that to map to it's activation.
type TabModelMapping struct {
	NewFunc  func() TabModel
	Commands []string
}

type Model struct {
	tabs            []TabModel
	commandMappings map[string]TabModelMapping
	active          int

	screen Size

	mode         int
	commandInput ui.CommandInput
	err          error
}

// New returns a new Model with the TabModels. The first TabModel in the slice will be the active one.  A panic will
// occur if no TabModels are given.  Any duplicate command mappings will just get overwritten with last winning.
func New(models []TabModelMapping) *Model {
	if len(models) == 0 {
		panic("must provide at least a single TabModel to run App")
	}

	a := new(Model)

	a.commandMappings = make(map[string]TabModelMapping)
	for _, m := range models {
		if len(m.Commands) == 0 {
			panic("must provide at least one switch command per model")
		}
		for _, cmd := range m.Commands {
			a.commandMappings[cmd] = m
		}
	}

	a.tabs = append(a.tabs, models[0].NewFunc())

	a.commandInput = ui.NewCommandInput()

	return a
}

func (a Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		a.tabs[a.active].Init(a.screen),
	)
}

func (a Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return a, tea.Quit

		case tea.KeyCtrlW:
			a.TabClose(a.active)
			return a, nil
		}

		if a.mode == ModeCommand {
			newInput, cmd := a.commandInput.Update(msg)
			a.commandInput = newInput
			return a, cmd
		} else {
			switch msg.String() {
			case "1", "2", "3", "4", "5":
				i, _ := strconv.Atoi(msg.String())
				a.TabSet(i - 1)
				return a, nil

			case ":":
				a.mode = ModeCommand
				a.commandInput.Focus()
				return a, nil
			}
		}

	case ui.CommandExecuteMsg:
		a.mode = ModeNormal

		cmd := msg.Name()
		if _, ok := a.commandMappings[cmd]; ok {
			a.tabs = append(a.tabs, a.commandMappings[cmd].NewFunc())
			a.active = len(a.tabs) - 1
			return a, a.tabs[a.active].Init(a.screen)
		}

		if cmd == "exit" {
			return a, tea.Quit
		}

	case ui.CommandExitMsg:
		a.mode = ModeNormal
		a.commandInput.Blur()
		return a, nil

	case tea.WindowSizeMsg:
		a.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

	case ErrorMsg:
		a.err = msg.error
		if a.err != nil {
			return a, func() tea.Msg {
				time.Sleep(5 * time.Second)
				return ErrorMsg{}
			}
		}
	}

	// If the message was not handled somewhere above, then it may be a message for the current TabView to handle.
	next, cmd := a.tabs[a.active].Update(msg)
	a.tabs[a.active] = next
	return a, cmd
}

func (a Model) View() string {
	viewportStyle := lipgloss.NewStyle().
		Width(a.screen.Width).
		Height(a.screen.Height - 3)
	var bottom string

	if a.err != nil {
		bottom += lipgloss.NewStyle().Foreground(ColorSalmon).Render(a.err.Error())
	}

	if a.mode == ModeCommand {
		bottom += "\n" + a.commandInput.View()
	} else {
		bottom += "\n"
	}

	titles := make([]string, len(a.tabs))
	for i, tab := range a.tabs {
		titles[i] = tab.Title()
	}

	return lipgloss.JoinVertical(0,
		tabBar.Render(a.screen.Width, titles, a.active),
		viewportStyle.Render(a.tabs[a.active].View()),
		bottom,
	)
}

// TabSet navigates to a specific Tab.  This is a noop if the tab does not exist.
func (a *Model) TabSet(i int) {
	if len(a.tabs) > i {
		a.active = i
	}
}

// TabClose closes a tab at the specified index.  Is a noop if tab does not
// exist.  The final tab cannot be closed.  If the current tab is active, then
// the previous tab will be set as active.
func (a *Model) TabClose(i int) {
	if len(a.tabs) > i && len(a.tabs) > 1 {
		if i >= a.active {
			a.active = max(a.active-1, 0)
		}
		a.tabs = append(a.tabs[:i], a.tabs[i+1:]...)
	}
}

type Size struct {
	Width  int
	Height int
}

// ErrorMsg is a message used to display an error to the user that dismisses after a few seconds.
type ErrorMsg struct {
	error
}

// NewErrorCmd is a way to generate a tea.Cmd that returns an ErrorMsg.
func NewErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{err}
	}
}
