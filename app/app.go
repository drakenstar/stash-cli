package app

import (
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/args"
	"github.com/drakenstar/stash-cli/bind"
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
	// Takes a command and interprets it, returning a tea.Msg that will be dispatched to the application.  The primary
	// use is for input commands to be convered into tea.Msg that can be executed in the TabModel (or other TabModel
	// instances).  This command must return a tea.Msg instead of a tea.Cmd, implying that it should not do I/O
	Interpret(Command) (tea.Msg, error)
}

// Command represets a command message that was input into the application in a specific mode.
type Command struct {
	Mode  Mode
	Input string
}

var tabID uint64

func nextTabID() uint {
	return uint(atomic.AddUint64(&tabID, 1))
}

type tab struct {
	id    uint
	model TabModel
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeCommand
	ModeFind
)

// TabModelConfig maps a given TabModel to the commands that to map to it's activation.
type TabModelConfig struct {
	NewFunc func() TabModel
	Name    string
	KeyBind string
}

type Model struct {
	tabs   []tab
	active int

	tabFuncs map[string](func() TabModel)
	keyBinds map[string]tea.Cmd

	screen Size

	mode         Mode
	commandInput ui.CommandInput
	err          error

	footer ui.Footer

	cmdService *cmdService
	opener     config.Opener
}

func New(stash stash.Stash, opener config.Opener) *Model {
	lookup := newCacheLookup()
	s := &cmdService{Stash: &cachingStash{stash, lookup}}

	models := []TabModelConfig{
		{
			NewFunc: func() TabModel { return NewScenesModel(s, lookup) },
			Name:    "scenes",
			KeyBind: "s",
		},
		{
			NewFunc: func() TabModel { return NewGalleriesModel(s, lookup) },
			Name:    "galleries",
			KeyBind: "g",
		},
	}

	m := &Model{
		cmdService: s,
		opener:     opener,
	}

	m.tabFuncs = make(map[string](func() TabModel))
	m.keyBinds = make(map[string]tea.Cmd)
	for _, mdl := range models {
		m.tabFuncs[mdl.Name] = mdl.NewFunc
		m.keyBinds[mdl.KeyBind] = func() tea.Msg { return TabOpenMsg{mdl.NewFunc} }
	}

	m.tabs = append(m.tabs, tab{uint(nextTabID()), models[0].NewFunc()})

	m.commandInput = ui.NewCommandInput()

	m.footer = ui.NewFooter()
	m.footer.Background = ColorBlack

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.commandInput.Init(),
		m.footer.Init(),
		m.tabs[m.active].model.Init(m.screen),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

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

		switch m.mode {
		case ModeCommand, ModeFind:
			m.commandInput, cmd = m.commandInput.Update(msg)
			return m, cmd

		default:
			switch msg.String() {
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				i, _ := strconv.Atoi(msg.String())
				m.TabSet(i - 1)
				return m, nil

			case ":":
				m.mode = ModeCommand
				return m, m.commandInput.Focus(":")

			case "/":
				m.mode = ModeFind
				return m, m.commandInput.Focus("/")
			}

			if bind, ok := m.keyBinds[msg.String()]; ok {
				return m, bind
			}
		}

	case ui.CommandExecMsg:
		m.mode = ModeNormal
		a := args.Parser(msg.Command)
		arg, err := a.Next()
		if err != nil {
			return m, NewErrorCmd(err)
		}

		switch arg.Value {
		case "tab":
			arg, err = a.Next()
			if err == io.EOF {
				return m, nil
			}
			if err != nil {
				return m, NewErrorCmd(err)
			}
			switch arg.Value {
			case "new":
				var dst struct {
					Name string `actions:",positional"`
				}
				if err := bind.Bind(a, &dst); err != nil {
					return m, NewErrorCmd(err)
				}

				tabFunc, ok := m.tabFuncs[dst.Name]
				if !ok {
					return m, NewErrorCmd(fmt.Errorf("invalid tab name '%s'", dst.Name))
				}
				return m, func() tea.Msg { return TabOpenMsg{tabFunc} }
			}

		case "exit":
			return m, tea.Quit
		}

		imsg, err := m.tabs[m.active].model.Interpret(Command{
			Mode:  m.mode,
			Input: msg.Command,
		})
		if err != nil && err != io.EOF {
			return m, NewErrorCmd(err)
		}
		return m, func() tea.Msg { return imsg }

	case ui.CommandExitMsg:
		m.mode = ModeNormal
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

	case TabOpenMsg:
		return m.TabOpen(msg.tabFunc())

	case OpenMsg:
		return m.openCmd(msg.target)
	}

	// If the message was not handled somewhere above, then it may be a message for the current TabView to handle.
	m.tabs[m.active].model, cmd = m.tabs[m.active].model.Update(msg)
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
func (a *Model) TabOpen(m TabModel) (tea.Model, tea.Cmd) {
	a.tabs = append(a.tabs, tab{nextTabID(), m})
	a.active = len(a.tabs) - 1
	return a, a.tabs[a.active].model.Init(a.screen)
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
	// Ignore EOF inputs, likely coming from end of argument input.
	if err == io.EOF {
		return nil
	}
	return func() tea.Msg {
		return ErrorMsg{err}
	}
}

type TabOpenMsg struct {
	tabFunc (func() TabModel)
}

type OpenMsg struct {
	target any
}
