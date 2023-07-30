package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
}

// AppModelMapping maps a given AppModel to the commands that to map to it's activation.
type AppModelMapping struct {
	Model    AppModel
	Commands []string
}

type Model struct {
	text            textinput.Model
	models          []AppModel
	commandMappings map[string]int
	active          int

	screen Size
	foo    AppModel
}

// New returns a new Model with the AppModels. The first AppModel in the slice will be the active one.  A panic will
// occur if no AppModels are given.  Any duplicate command mappings will just get overwritten with last winning.
func New(models []AppModelMapping) *Model {
	if len(models) == 0 {
		panic("must provide at least a single AppModel to run App")
	}

	a := new(Model)

	a.commandMappings = make(map[string]int)
	for i, m := range models {
		a.models = append(a.models, m.Model)
		for _, cmd := range m.Commands {
			a.commandMappings[cmd] = i
		}
	}

	a.text = textinput.New()
	a.text.Focus()

	return a
}

func (a Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		a.models[a.active].Init(a.screen),
	)
}

func (a Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return a, tea.Quit

		case tea.KeyEnter:
			cmd := NewInputCmd(a.text.Value())
			a.text.SetValue("")
			return a, cmd
		}

	case Input:
		cmd := msg.Command()
		if i, ok := a.commandMappings[cmd]; ok && i != a.active {
			a.active = i
			return a, a.models[a.active].Init(a.screen)
		}

		if cmd == "exit" {
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		a.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
	}

	a.text, cmd = a.text.Update(msg)
	cmds = append(cmds, cmd)

	next, cmd := a.models[a.active].Update(msg)
	a.models[a.active] = next
	cmds = append(cmds, cmd)

	return a, tea.Batch(cmds...)
}

func (a Model) View() string {
	return lipgloss.JoinVertical(0,
		a.models[a.active].View(),
		a.text.View(),
	)
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

// Returns all text after the initial command.  This may be interpretted in any way an action deems appropriate.
func (i Input) ArgString() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return ""
	}
	return string(i[idx+1:])
}
