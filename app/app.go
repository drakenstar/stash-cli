package app

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/ui"
)

// AppModel is a subview of the App application, and operates similarly to a tea.Model.
type AppModel interface {
	// Provided the size of the current area the AppModel is to render within. Should return any initialisation
	// commands to execute.
	Init(Size) tea.Cmd
	// Handle any tea.Msg that the app.Model passes on.  Should by a immutable operation and return a new AppModel
	// with the new state after handling the message.
	Update(tea.Msg) (AppModel, tea.Cmd)
	// Normal tea.Model:View method, should render the current state of the view as a string.
	View() string
	// Returns a string name to be used for the tab title
	TabTitle() string
}

// AppModelMapping maps a given AppModel to the commands that to map to it's activation.
type AppModelMapping struct {
	NewFunc  func() AppModel
	Commands []string
}

type Model struct {
	tabs            []AppModel
	commandMappings map[string]AppModelMapping
	active          int

	screen Size

	text         textinput.Model
	confirmation *ui.Confirmation
	err          error
}

// New returns a new Model with the AppModels. The first AppModel in the slice will be the active one.  A panic will
// occur if no AppModels are given.  Any duplicate command mappings will just get overwritten with last winning.
func New(models []AppModelMapping) *Model {
	if len(models) == 0 {
		panic("must provide at least a single AppModel to run App")
	}

	a := new(Model)

	a.commandMappings = make(map[string]AppModelMapping)
	for _, m := range models {
		if len(m.Commands) == 0 {
			panic("must provide at least one switch command per model")
		}
		for _, cmd := range m.Commands {
			a.commandMappings[cmd] = m
		}
	}

	a.tabs = append(a.tabs, models[0].NewFunc())

	a.text = textinput.New()
	a.text.Focus()

	return a
}

func (a Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		a.tabs[a.active].Init(a.screen),
	)
}

func (a Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return a, tea.Quit

		case tea.KeyEsc:
			if a.confirmation == nil {
				return a, tea.Quit
			}
			a.confirmation = nil

		case tea.KeyEnter:
			if a.confirmation == nil {
				cmd := NewInputCmd(a.text.Value())
				a.text.SetValue("")
				return a, cmd
			}

		case tea.KeyF1:
			a.TabSet(0)
			return a, nil
		case tea.KeyF2:
			a.TabSet(1)
			return a, nil
		case tea.KeyF3:
			a.TabSet(2)
			return a, nil
		case tea.KeyF4:
			a.TabSet(3)
			return a, nil
		case tea.KeyF5:
			a.TabSet(4)
			return a, nil
		// TODO more tab bindings?

		case tea.KeyCtrlW:
			a.TabClose(a.active)
			return a, nil
		}

	case Input:
		cmd := msg.Command()
		if _, ok := a.commandMappings[cmd]; ok {
			a.confirmation = nil
			a.tabs = append(a.tabs, a.commandMappings[cmd].NewFunc())
			a.active = len(a.tabs) - 1
			return a, a.tabs[a.active].Init(a.screen)
		}

		if cmd == "exit" {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

	case ConfirmationMsg:
		a.confirmation = &ui.Confirmation{
			Message: msg.Message,
			Options: []ui.ConfirmationOption{
				{Text: msg.CancelOption, Cmd: ConfirmationCancelCmd},
				{Text: msg.ConfirmOption, Cmd: tea.Batch(ConfirmationCancelCmd, msg.Cmd)},
			},
		}

	case ConfirmationCancelMsg:
		a.confirmation = nil

	case ErrorMsg:
		a.err = msg.error
		if a.err != nil {
			return a, func() tea.Msg {
				time.Sleep(5 * time.Second)
				return ErrorMsg{}
			}
		}
	}

	if a.confirmation != nil {
		a.confirmation, cmd = a.confirmation.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		a.text, cmd = a.text.Update(msg)
		cmds = append(cmds, cmd)
	}

	next, cmd := a.tabs[a.active].Update(msg)
	a.tabs[a.active] = next
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a Model) View() string {
	viewportStyle := lipgloss.NewStyle().
		Width(a.screen.Width).
		Height(a.screen.Height - 3)
	var bottom string
	if a.confirmation != nil {
		bottom = a.confirmation.View()
	} else {
		if a.err != nil {
			bottom += lipgloss.NewStyle().Foreground(ColorSalmon).Render(a.err.Error())
		}
		bottom += "\n" + a.text.View()
	}

	titles := make([]string, len(a.tabs))
	for i, tab := range a.tabs {
		titles[i] = tab.TabTitle()
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

// Input represents a complete line of input from the user
type Input string

func NewInputCmd(s string) tea.Cmd {
	return func() tea.Msg {
		return Input(strings.TrimSpace(s))
	}
}

// Command returns all characters up to the first encountered space in an input string.  This is to be interpretted
// as the command for the rest of the input.
func (i Input) Command() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return string(i)
	}
	return string(i[:idx])
}

// ArgString returns all text after the initial command.  This may be interpretted in any way an action deems appropriate.
func (i Input) ArgString() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return ""
	}
	return string(i[idx+1:])
}

// ArgInt attempts to parse any value given after the command as an integer.
func (i Input) ArgInt() (int, error) {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return 0, fmt.Errorf("no argument given")
	}
	return strconv.Atoi(string(i[idx+1:]))
}

// Args returns a tokenised set of arguments that come after the initial command, not including the command itself.
// Tokens are split on space, with multiple spaces being ignored.
func (i Input) Args() []string {
	return strings.Fields(i.ArgString())
}

// ConfirmationMessage is a message that will prompt the user to confirm a command before confirming it. If the user
// selects the ConfirmOption text, the command is dispatched.  Otherwise nothing occurs.
type ConfirmationMsg struct {
	Cmd           tea.Cmd
	Message       string
	ConfirmOption string
	CancelOption  string
}

// ConfirmationCancelMsg cancels any existing confirmation modal. This is intended as the return message for when
// cancel is selected from a confirmation.
type ConfirmationCancelMsg struct{}

func ConfirmationCancelCmd() tea.Msg {
	return ConfirmationCancelMsg{}
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
