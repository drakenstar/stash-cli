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

	footer ui.Footer

	cmdService *cmdService
}

func New(stash stash.Stash, opener config.Opener) *Model {
	lookup := newCacheLookup()
	s := &cmdService{Stash: &cachingStash{stash, lookup}}

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

	m := new(Model)
	m.cmdService = s

	m.tabFuncs = make(map[string](func() TabModel))
	m.keyBinds = make(map[string]tea.Cmd)
	for _, mdl := range models {
		m.tabFuncs[mdl.Name] = mdl.NewFunc
		m.keyBinds[mdl.KeyBind] = func() tea.Msg { return TabOpenMsg{mdl.NewFunc} }
	}

	m.tabs = append(m.tabs, models[0].NewFunc())

	m.commandInput = ui.NewCommandInput()

	m.footer = ui.NewFooter()
	m.footer.Background = ColorBlack

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.commandInput.Init(),
		m.footer.Init(),
		m.tabs[m.active].Init(m.screen),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.commandInput, cmd = m.commandInput.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	m.footer, cmd = m.footer.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlW:
			m.TabClose(m.active)
			return m, nil
		}

		if m.mode == ModeCommand {
			newInput, cmd := m.commandInput.Update(msg)
			m.commandInput = newInput
			return m, cmd
		} else {
			switch msg.String() {
			case "1", "2", "3", "4", "5":
				i, _ := strconv.Atoi(msg.String())
				m.TabSet(i - 1)
				return m, nil

			case ":":
				return m, NewModeCommandCmd(":", "")
			}

			if bind, ok := m.keyBinds[msg.String()]; ok {
				return m, bind
			}
		}

	case ui.CommandExecuteMsg:
		m.mode = ModeNormal

		cmd := msg.Name()
		if cmd == "tab" {
			args := msg.Args()
			if len(args) == 2 && args[0] == "new" {
				tabFunc, ok := m.tabFuncs[args[1]]
				if !ok {
					return m, NewErrorCmd(fmt.Errorf("invalid tab name '%s'", args[1]))
				}
				return m, func() tea.Msg { return TabOpenMsg{tabFunc} }
			}
		}

		if cmd == "exit" {
			return m, tea.Quit
		}

	case ui.CommandExitMsg:
		m.mode = ModeNormal
		m.commandInput.Blur()
		return m, nil

	case tea.WindowSizeMsg:
		m.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

	case ErrorMsg:
		m.err = msg.error
		if m.err != nil {
			return m, func() tea.Msg {
				time.Sleep(5 * time.Second)
				return ErrorMsg{}
			}
		}

	case ModeCommandMsg:
		m.mode = ModeCommand
		return m, m.commandInput.Focus(msg.prompt, msg.prefix)

	case TabOpenMsg:
		return m.TabOpen(msg.tabFunc())
	}

	// If the message was not handled somewhere above, then it may be a message for the current TabView to handle.
	m.tabs[m.active], cmd = m.tabs[m.active].Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	viewportStyle := lipgloss.NewStyle().
		Width(m.screen.Width).
		Height(m.screen.Height - 3)
	var bottom string

	if m.err != nil {
		bottom += lipgloss.NewStyle().Foreground(ColorSalmon).Render(m.err.Error())
	}

	if m.mode == ModeCommand {
		bottom += "\n" + m.commandInput.View()
	} else {
		bottom += "\n" + m.footer.Render(m.screen.Width, m.cmdService.AnyLoading())
	}

	titles := make([]string, len(m.tabs))
	for i, tab := range m.tabs {
		titles[i] = tab.Title()
	}

	return lipgloss.JoinVertical(0,
		tabBar.Render(m.screen.Width, titles, m.active),
		viewportStyle.Render(m.tabs[m.active].View()),
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
func (m *Model) TabSet(i int) {
	if len(m.tabs) > i {
		m.active = i
	}
}

// TabClose closes a tab at the specified index.  Is a noop if tab does not exist.  The final tab cannot be closed.
// If the current tab is active, then the previous tab will be set as active.
func (m *Model) TabClose(i int) {
	if len(m.tabs) > i && len(m.tabs) > 1 {
		if i >= m.active {
			m.active = max(m.active-1, 0)
		}
		m.tabs = append(m.tabs[:i], m.tabs[i+1:]...)
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
