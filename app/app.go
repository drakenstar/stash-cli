package app

import (
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
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

// TabModelConfig maps a given TabModel to the commands that to map to it's activation.
type TabModelConfig struct {
	NewFunc func() TabModel
	Name    string
	KeyBind string
}

type Model struct {
	tabs   []TabModel
	active int

	tabFuncs map[string](func() TabModel)
	keyBinds map[string]tea.Cmd

	screen Size

	mode         int
	commandInput ui.CommandInput
	err          error
}

func New(stash stash.Stash, opener config.Opener) *Model {
	lookup := newCacheLookup()
	s := &cachingStash{stash, lookup}

	models := []TabModelConfig{
		{
			NewFunc: func() TabModel { return NewScenesModel(s, lookup, opener) },
			Name:    "scenes",
			KeyBind: "s",
		},
		{
			NewFunc: func() TabModel { return NewGalleriesModel(s, lookup, opener) },
			Name:    "galleries",
			KeyBind: "g",
		},
	}

	a := new(Model)

	a.tabFuncs = make(map[string](func() TabModel))
	a.keyBinds = make(map[string]tea.Cmd)
	for _, m := range models {
		a.tabFuncs[m.Name] = m.NewFunc
		a.keyBinds[m.KeyBind] = func() tea.Msg { return TabOpenMsg{m.NewFunc} }
	}

	a.tabs = append(a.tabs, models[0].NewFunc())

	a.commandInput = ui.NewCommandInput()

	return a
}

func (a Model) Init() tea.Cmd {
	return a.tabs[a.active].Init(a.screen)
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
				return a, NewModeCommandCmd(":", "")
			}

			if bind, ok := a.keyBinds[msg.String()]; ok {
				return a, bind
			}
		}

	case ui.CommandExecuteMsg:
		a.mode = ModeNormal

		cmd := msg.Name()
		if cmd == "tab" {
			args := msg.Args()
			if len(args) == 2 && args[0] == "new" {
				tabFunc, ok := a.tabFuncs[args[1]]
				if !ok {
					return a, NewErrorCmd(fmt.Errorf("invalid tab name '%s'", args[1]))
				}
				return a, func() tea.Msg { return TabOpenMsg{tabFunc} }
			}
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

	case ModeCommandMsg:
		a.mode = ModeCommand
		return a, a.commandInput.Focus(msg.prompt, msg.prefix)

	case TabOpenMsg:
		return a.TabOpen(msg.tabFunc())
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

// TabOpen creates a new tab with the given TabModel and sets it as active.
func (a *Model) TabOpen(m TabModel) (tea.Model, tea.Cmd) {
	a.tabs = append(a.tabs, m)
	a.active = len(a.tabs) - 1
	return a, a.tabs[a.active].Init(a.screen)
}

// TabSet navigates to a specific Tab.  This is a noop if the tab does not exist.
func (a *Model) TabSet(i int) {
	if len(a.tabs) > i {
		a.active = i
	}
}

// TabClose closes a tab at the specified index.  Is a noop if tab does not exist.  The final tab cannot be closed.
// If the current tab is active, then the previous tab will be set as active.
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

type ModeCommandMsg struct {
	prompt string
	prefix string
}

func NewModeCommandCmd(prompt, prefix string) tea.Cmd {
	return func() tea.Msg {
		return ModeCommandMsg{prompt: prompt, prefix: prefix}
	}
}

type TabOpenMsg struct {
	tabFunc (func() TabModel)
}
