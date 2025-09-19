package app

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/command"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

// TabModel is a subview of the App application, and operates similarly to a tea.Model.
type TabModel interface {
	tea.Model

	// Title should return a string name to be used for the tab title.
	Title() string
	// CommandConfig returns a Config for all commands the TabModel supports.
	CommandConfig() command.Config
	// Search should return a message to support searching.
	Search(string) tea.Msg
	// SetSize allows the app to inform a tab of the space is has for layout.  The tab can optionally return a tea.Cmd
	// to execute in response to this.
	// TODO probably this can just be handled by a message rather than an imperative call.
	SetSize(Size) tea.Cmd
}

// Command represets a command message that was input into the application in a specific mode.
type Command struct {
	Mode  Mode
	Input string
}

type tab struct {
	id    tabID
	model TabModel
}

type tabID = uint64

type Mode int

const (
	ModeNormal Mode = iota
	ModeCommand
	ModeFind
)

// TabModelConfig defines a "type" of tab that can be opened within the Model.
type TabModelConfig struct {
	NewFunc TabNewFunc
	Name    string
}

type TabNewFunc func(tabID) TabModel

var ModelDefaultKeymap = map[string]string{
	"s":      "tab new scenes",
	"g":      "tab new galleries",
	"1":      "tab switch 1",
	"2":      "tab switch 2",
	"3":      "tab switch 3",
	"4":      "tab switch 4",
	"5":      "tab switch 5",
	"6":      "tab switch 6",
	"7":      "tab switch 7",
	"8":      "tab switch 8",
	"9":      "tab switch 9",
	"ctrl+w": "tab close",
	"ctrl+c": "exit",
}

type Model struct {
	tabs     []tab
	tabsByID map[tabID]tab
	active   int
	tabID    tabID

	tabFuncs map[string](TabNewFunc)

	screen Size

	mode         Mode
	commandInput ui.CommandInput
	err          error

	footer ui.Footer

	cmdService *cmdService
	opener     config.Opener

	command command.Config
}

func New(stash stash.Stash, opener config.Opener) *Model {
	lookup := newCacheLookup()
	s := &cmdService{Stash: &cachingStash{stash, lookup}}

	models := []TabModelConfig{
		{
			NewFunc: func(id tabID) TabModel {
				s := &cmdServiceWithID{s, id}
				return NewScenesModel(s, lookup)
			},
			Name: "scenes",
		},
		{
			NewFunc: func(id tabID) TabModel {
				s := &cmdServiceWithID{s, id}
				return NewGalleriesModel(s, lookup)
			},
			Name: "galleries",
		},
	}

	m := &Model{
		cmdService: s,
		opener:     opener,
	}

	m.tabFuncs = make(map[string]TabNewFunc)
	for _, mdl := range models {
		m.tabFuncs[mdl.Name] = mdl.NewFunc
	}

	m.commandInput = ui.NewCommandInput()

	m.footer = ui.NewFooter()
	m.footer.Background = ColorBlack

	m.command = command.Config{
		"exit": static(tea.QuitMsg{}),
		"tab": {
			SubCommands: command.Config{
				"close": binder[ModelTabCloseMsg](),
				"new": {
					Resolve: func(i command.Iterator) (any, error) {
						next, err := i.Next()
						if err != nil {
							return nil, err
						}
						tabConfig, ok := m.tabFuncs[next.Value]
						if !ok {
							return ErrorMsg{fmt.Errorf("unknown tab type '%s'", next.Value)}, nil
						}
						return ModelTabNewMsg{tabConfig}, nil
					},
				},
				"switch": binder[ModelTabSwitchMsg](),
			},
		},
	}

	m.tabsByID = make(map[tabID]tab)

	id := m.nextTabID()
	m.tabs = []tab{{id, models[0].NewFunc(id)}}
	m.tabsByID[id] = m.tabs[0]

	return m
}

func (m *Model) nextTabID() tabID {
	return tabID(atomic.AddUint64(&m.tabID, 1))
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.commandInput.Init(),
		m.footer.Init(),
		m.tabs[m.active].model.Init(),
	)
}

type ModelTabNewMsg struct {
	NewFunc TabNewFunc
}

type ModelTabSwitchMsg struct {
	Index int `command:",positional"`
}

type ModelTabCloseMsg struct{}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.footer, cmd = m.footer.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch m.mode {
		case ModeCommand, ModeFind:
			m.commandInput, cmd = m.commandInput.Update(msg)
			return m, cmd

		default:
			switch msg.String() {
			case ":":
				m.mode = ModeCommand
				return m, m.commandInput.Focus(":")

			case "/":
				m.mode = ModeFind
				return m, m.commandInput.Focus("/")
			}

			if cmd, ok := ModelDefaultKeymap[msg.String()]; ok {
				return m, func() tea.Msg { return ui.CommandExecMsg{Command: cmd} }
			}
		}

	case ModelTabCloseMsg:
		m.TabClose(m.active)
		return m, nil

	case ModelTabNewMsg:
		m.TabOpen(msg.NewFunc)
		return m, tea.Batch(
			m.tabs[m.active].model.Init(),
			m.tabs[m.active].model.SetSize(Size{Width: m.screen.Width, Height: m.screen.Height - 5}))

	case ModelTabSwitchMsg:
		m.TabSet(msg.Index - 1)
		return m, nil

	case ui.CommandExecMsg:
		if m.mode == ModeFind {
			m.mode = ModeNormal
			findMsg := m.tabs[m.active].model.Search(msg.Command)
			return m, func() tea.Msg { return findMsg }
		}

		m.mode = ModeNormal

		// First attempt to resolve to a Model command, since these take precedence.
		ret, err := m.command.Resolve(command.Parser(msg.Command))
		if err != nil {
			if !errors.As(err, &command.UnmatchedCommandError{}) {
				return m, NewErrorCmd(err)
			}
		} else {
			return m, func() tea.Msg { return ret }
		}

		// We're still here, means that the command wasn't matched.  So let's attempt to match a TabModel command.
		ret, err = m.tabs[m.active].model.CommandConfig().Resolve(command.Parser(msg.Command))
		if err != nil {
			return m, NewErrorCmd(err)
		}
		return m, func() tea.Msg { return ret }

	case ui.CommandExitMsg:
		m.mode = ModeNormal
		return m, nil

	case tea.WindowSizeMsg:
		m.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
		// All tabs should be notified about the change in window size, as it will cause them to refetch their results.
		tabSize := Size{
			Width:  msg.Width,
			Height: msg.Height - 5,
		}
		cmds := make([]tea.Cmd, len(m.tabs))
		for i := range m.tabs {
			cmds[i] = m.tabs[i].model.SetSize(tabSize)
		}
		return m, tea.Batch(cmds...)

	case ErrorMsg:
		m.err = msg.error
		if m.err != nil {
			return m, func() tea.Msg {
				time.Sleep(5 * time.Second)
				return ErrorMsg{}
			}
		}

	case OpenMsg:
		return m.openCmd(msg.target)

	// loadingMsg handles routing of a return loading message to the correct tab located by ID.
	case loadingMsg:
		_, cmd := m.tabsByID[msg.id].model.Update(msg.payload)
		return m, cmd
	}

	// If the message was not handled somewhere above, then it may be a message for the current TabView to handle.
	_, cmd = m.tabs[m.active].model.Update(msg)
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

	switch m.mode {
	case ModeCommand, ModeFind:
		bottom += "\n" + m.commandInput.View()
	default:
		bottom += "\n" + m.footer.Render(m.screen.Width, m.cmdService.AnyLoading())
	}

	// Build our tabs, we provide keyboard shortcuts for the numbers 1-9.
	// TODO This should probably be dynamic based on a key map.
	titles := make([]ui.Tab, len(m.tabs))
	for i, tab := range m.tabs {
		titles[i] = ui.Tab{
			Label: tab.model.Title(),
		}
		if i < 9 {
			titles[i].Prefix = strconv.Itoa(i + 1)
		}
	}

	return lipgloss.JoinVertical(0,
		tabBar.Render(m.screen.Width, titles, m.active),
		viewportStyle.Render(m.tabs[m.active].model.View()),
		bottom,
	)
}

// openCmd opens the given target in the system opener asynchronously.  Errors will be displayed.
func (m *Model) openCmd(target any) (*Model, tea.Cmd) {
	return m, func() tea.Msg {
		err := m.opener(target)
		if err != nil {
			return ErrorMsg{err}
		}
		return nil
	}
}

// TabOpen creates a new tab with the given TabModel and sets it as active.
func (m *Model) TabOpen(newFunc TabNewFunc) {
	id := m.nextTabID()
	m.tabs = append(m.tabs, tab{id, newFunc(id)})
	m.active = len(m.tabs) - 1
	m.tabsByID[id] = m.tabs[m.active]
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
		t := m.tabs[i]
		delete(m.tabsByID, t.id)
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
	// Ignore EOF inputs, likely coming from end of argument input.
	if err == io.EOF {
		return nil
	}
	return func() tea.Msg {
		return ErrorMsg{err}
	}
}

type OpenMsg struct {
	target any
}
