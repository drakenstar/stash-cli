package app

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode"

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

	mode              Mode
	commandInput      ui.CommandInput
	confirmation      *ui.Confirmation
	pendingDelete     *pendingDeleteState
	tagsLoading       bool
	studiosLoading    bool
	performersLoading bool
	err               error

	footer ui.Footer

	cmdService *cmdService
	opener     config.Opener

	command command.Config
}

func New(stash stash.Stash, opener config.Opener) *Model {
	lookup := newCacheLookup()
	s := &cmdService{
		Stash: &cachingStash{stash, lookup},
		cache: lookup,
	}

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
	m.commandInput.SetWidth(m.screen.Width)
	m.commandInput.SetSuggestionProvider(m.commandSuggestionProvider())

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
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.pendingDelete != nil {
			return m, nil
		}
		if m.confirmation != nil {
			m.confirmation, cmd = m.confirmation.Update(msg)
			return m, cmd
		}

		switch m.mode {
		case ModeCommand, ModeFind:
			m.commandInput, cmd = m.commandInput.Update(msg)
			var cmds []tea.Cmd
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			if m.mode == ModeCommand {
				if loadCmd := m.refreshCommandAutocomplete(); loadCmd != nil {
					cmds = append(cmds, loadCmd)
				}
			}
			return m, tea.Batch(cmds...)

		default:
			switch msg.String() {
			case ":":
				m.mode = ModeCommand
				cmds := []tea.Cmd{m.commandInput.Focus(":")}
				if loadCmd := m.refreshCommandAutocomplete(); loadCmd != nil {
					cmds = append(cmds, loadCmd)
				}
				return m, tea.Batch(cmds...)

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

	case deleteRequestMsg:
		if msg.SkipConfirm {
			return m.beginDelete(msg)
		}
		confirmation := ui.Confirmation{
			Message: m.deleteConfirmationMessage(msg),
			Options: []ui.ConfirmationOption{
				{
					Text: "Cancel",
					Cmd:  func() tea.Msg { return dismissModalMsg{} },
				},
				{
					Text: "Delete",
					Cmd: func() tea.Msg {
						return confirmDeleteMsg{Request: msg}
					},
				},
			},
		}
		m.confirmation = &confirmation
		return m, nil

	case confirmDeleteMsg:
		return m.beginDelete(msg.Request)

	case dismissModalMsg:
		m.confirmation = nil
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

	case tagsLoadedMsg:
		m.tagsLoading = false
		if m.mode == ModeCommand {
			m.commandInput.RefreshSuggestions()
		}
		return m, nil

	case studiosLoadedMsg:
		m.studiosLoading = false
		if m.mode == ModeCommand {
			m.commandInput.RefreshSuggestions()
		}
		return m, nil

	case performersLoadedMsg:
		m.performersLoading = false
		if m.mode == ModeCommand {
			m.commandInput.RefreshSuggestions()
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
		m.commandInput.SetWidth(msg.Width)
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
		if m.pendingDelete != nil {
			m.pendingDelete = nil
		}
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
		if errMsg, ok := msg.payload.(ErrorMsg); ok {
			if m.pendingDelete != nil && m.pendingDelete.tabID == msg.id {
				m.pendingDelete = nil
			}
			return m.Update(errMsg)
		}
		_, cmd := m.tabsByID[msg.id].model.Update(msg.payload)
		if m.pendingDelete != nil && m.pendingDelete.tabID == msg.id {
			switch msg.payload.(type) {
			case ErrorMsg, scenesMsg, galleriesMsg:
				m.pendingDelete = nil
			}
		}
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
	var bottom string
	switch m.mode {
	case ModeCommand, ModeFind:
		bottom = m.commandInput.View()
	default:
		bottom = m.footer.Render(m.screen.Width, m.cmdService.AnyLoading())
	}

	if m.err != nil {
		errLine := lipgloss.NewStyle().Foreground(ColorSalmon).Render(m.err.Error())
		bottom = lipgloss.JoinVertical(0, errLine, bottom)
	}

	viewportHeight := max(m.screen.Height-1-lipgloss.Height(bottom), 0)
	viewportStyle := lipgloss.NewStyle().
		Width(m.screen.Width).
		Height(viewportHeight)

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

	view := lipgloss.JoinVertical(0,
		tabBar.Render(m.screen.Width, titles, m.active),
		viewportStyle.Render(truncateLines(m.tabs[m.active].model.View(), viewportHeight)),
		bottom,
	)

	if m.confirmation != nil {
		return m.renderModal("Confirm Delete", m.confirmation.View())
	}
	if m.pendingDelete != nil {
		return m.renderModal("Deleting", m.deleteProgressMessage())
	}
	return view
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

func (m *Model) beginDelete(request deleteRequestMsg) (*Model, tea.Cmd) {
	m.mode = ModeNormal
	m.confirmation = nil
	m.pendingDelete = &pendingDeleteState{
		tabID:   m.tabs[m.active].id,
		request: request,
	}
	return m, request.DeleteCmd
}

func (m Model) deleteConfirmationMessage(request deleteRequestMsg) string {
	return fmt.Sprintf(
		"Delete %s?\n\n%s\n%s\n\nThis will remove it from Stash and delete associated files.",
		request.Entity,
		request.Title,
		request.Path,
	)
}

func (m Model) deleteProgressMessage() string {
	if m.pendingDelete == nil {
		return ""
	}
	return fmt.Sprintf(
		"%s Deleting %s...\n\n%s\n%s",
		m.footer.SpinnerView(),
		m.pendingDelete.request.Entity,
		m.pendingDelete.request.Title,
		m.pendingDelete.request.Path,
	)
}

func (m Model) renderModal(title, body string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorStatusCell).
		Padding(1, 2).
		MaxWidth(max(m.screen.Width-8, 20)).
		Align(lipgloss.Left).
		Render(lipgloss.JoinVertical(1,
			lipgloss.NewStyle().Bold(true).Foreground(ColorOffWhite).Render(title),
			body,
		))

	return lipgloss.JoinVertical(0,
		tabBar.Render(m.screen.Width, m.tabTitles(), m.active),
		lipgloss.Place(m.screen.Width, max(m.screen.Height-1, 1), lipgloss.Center, lipgloss.Center, box),
	)
}

func (m Model) tabTitles() []ui.Tab {
	titles := make([]ui.Tab, len(m.tabs))
	for i, tab := range m.tabs {
		titles[i] = ui.Tab{
			Label: tab.model.Title(),
		}
		if i < 9 {
			titles[i].Prefix = strconv.Itoa(i + 1)
		}
	}
	return titles
}

func (m Model) commandSuggestions() []string {
	seen := make(map[string]struct{})
	var suggestions []string

	for _, name := range m.command.Commands() {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		suggestions = append(suggestions, name)
	}
	for _, name := range m.tabs[m.active].model.CommandConfig().Commands() {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		suggestions = append(suggestions, name)
	}

	slices.Sort(suggestions)
	return suggestions
}

func (m *Model) commandSuggestionProvider() ui.SuggestionProvider {
	return func(prompt, input string, cursor int) ui.SuggestionSet {
		set, _ := m.commandSuggestionSet(prompt, input, cursor)
		return set
	}
}

func (m *Model) refreshCommandAutocomplete() tea.Cmd {
	m.commandInput.RefreshSuggestions()
	_, needs := m.commandSuggestionSet(":", m.commandInput.Value(), m.commandInput.CursorPosition())

	var cmds []tea.Cmd
	if needs.tags && !m.cmdService.cache.TagsLoaded() && !m.tagsLoading {
		m.tagsLoading = true
		cmds = append(cmds, m.cmdService.TagsAll())
	}
	if needs.studios && !m.cmdService.cache.StudiosLoaded() && !m.studiosLoading {
		m.studiosLoading = true
		cmds = append(cmds, m.cmdService.StudiosAll())
	}
	if needs.performers && !m.cmdService.cache.PerformersLoaded() && !m.performersLoading {
		m.performersLoading = true
		cmds = append(cmds, m.cmdService.PerformersAll())
	}
	return tea.Batch(cmds...)
}

type suggestionRequirements struct {
	tags       bool
	studios    bool
	performers bool
}

func (m Model) commandSuggestionSet(prompt, input string, cursor int) (ui.SuggestionSet, suggestionRequirements) {
	if prompt != ":" {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}

	token, ok := tokenAtCursor(input, cursor)
	if !ok {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}

	if token.index == 0 {
		prefix := input[token.start:cursor]
		if strings.TrimSpace(prefix) == "" {
			return ui.SuggestionSet{}, suggestionRequirements{}
		}

		var suggestions []ui.Suggestion
		for _, name := range m.commandSuggestions() {
			if strings.HasPrefix(name, prefix) {
				suggestions = append(suggestions, ui.Suggestion{Display: name, Value: name})
			}
			if len(suggestions) == 6 {
				break
			}
		}
		return ui.SuggestionSet{
			Start:       token.start,
			End:         token.end,
			Suggestions: suggestions,
		}, suggestionRequirements{}
	}

	if len(token.tokens) > 0 && token.tokens[0].raw == "tag" {
		return m.tagCommandSuggestionSet(token, input, cursor)
	}

	if len(token.tokens) == 0 || token.tokens[0].raw != "filter" {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}

	eq := strings.IndexByte(token.raw, '=')
	if eq < 0 || cursor <= token.start+eq {
		return m.filterArgumentSuggestionSet(token, input, cursor), suggestionRequirements{}
	}

	argName := token.raw[:eq]
	valueStart := token.start + eq + 1
	if cursor < valueStart {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}

	prefixRaw := input[valueStart:cursor]
	searchPrefix := strings.TrimPrefix(prefixRaw, "\"")
	if searchPrefix == "" {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}
	switch argName {
	case "tag", "performertag":
		if !m.cmdService.cache.TagsLoaded() {
			return ui.SuggestionSet{}, suggestionRequirements{tags: true}
		}
		tags := m.cmdService.cache.TagsByPrefix(searchPrefix, 6)
		return entitySuggestionSet(valueStart, token.end, tagSuggestions(tags)), suggestionRequirements{}

	case "studio":
		if !m.cmdService.cache.StudiosLoaded() {
			return ui.SuggestionSet{}, suggestionRequirements{studios: true}
		}
		studios := m.cmdService.cache.StudiosByPrefix(searchPrefix, 6)
		return entitySuggestionSet(valueStart, token.end, studioSuggestions(studios)), suggestionRequirements{}

	case "performer":
		suggestions := currentSuggestion(searchPrefix)
		needs := suggestionRequirements{}
		if !m.cmdService.cache.PerformersLoaded() {
			needs.performers = true
			return entitySuggestionSet(valueStart, token.end, suggestions), needs
		}
		performers := m.cmdService.cache.PerformersByPrefix(searchPrefix, 5)
		suggestions = append(suggestions, performerSuggestions(performers)...)
		return entitySuggestionSet(valueStart, token.end, suggestions), needs
	}

	return ui.SuggestionSet{}, suggestionRequirements{}
}

func (m Model) tagCommandSuggestionSet(token commandToken, input string, cursor int) (ui.SuggestionSet, suggestionRequirements) {
	if token.index == 0 {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}
	if cursor < token.start {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}
	searchPrefix := strings.TrimPrefix(input[token.start:cursor], "\"")
	if searchPrefix == "" {
		return ui.SuggestionSet{}, suggestionRequirements{}
	}
	if !m.cmdService.cache.TagsLoaded() {
		return ui.SuggestionSet{}, suggestionRequirements{tags: true}
	}
	tags := m.cmdService.cache.TagsByPrefix(searchPrefix, 6)
	return entitySuggestionSet(token.start, token.end, tagSuggestions(tags)), suggestionRequirements{}
}

func (m Model) filterArgumentSuggestionSet(token commandToken, input string, cursor int) ui.SuggestionSet {
	prefix := input[token.start:cursor]
	if strings.TrimSpace(prefix) == "" {
		return ui.SuggestionSet{}
	}

	suggestions := make([]ui.Suggestion, 0, 6)
	for _, name := range m.filterArgumentNames() {
		if strings.HasPrefix(name, prefix) {
			suggestions = append(suggestions, ui.Suggestion{
				Display: name,
				Value:   name + "=",
			})
		}
		if len(suggestions) == 6 {
			break
		}
	}
	return ui.SuggestionSet{
		Start:       token.start,
		End:         token.end,
		Suggestions: suggestions,
	}
}

func (m Model) filterArgumentNames() []string {
	switch m.tabs[m.active].model.(type) {
	case *ScenesModel:
		return filterArgumentNamesFor[ScenesModelFilterMsg]()
	case *GalleriesModel:
		return filterArgumentNamesFor[GalleriesModelFilterMsg]()
	default:
		return nil
	}
}

func filterArgumentNamesFor[T any]() []string {
	var msg T
	t := reflect.TypeOf(msg)
	names := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" || field.Anonymous {
			continue
		}
		name := strings.ToLower(field.Name)
		if tag := field.Tag.Get(command.TagKey); tag != "" {
			name = strings.Split(tag, ",")[0]
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func entitySuggestionSet(start, end int, suggestions []ui.Suggestion) ui.SuggestionSet {
	return ui.SuggestionSet{
		Start:       start,
		End:         end,
		Suggestions: suggestions,
	}
}

func tagSuggestions(tags []stash.Tag) []ui.Suggestion {
	suggestions := make([]ui.Suggestion, 0, len(tags))
	for _, tag := range tags {
		suggestions = append(suggestions, quotedSuggestion(tag.Name))
	}
	return suggestions
}

func studioSuggestions(studios []stash.Studio) []ui.Suggestion {
	suggestions := make([]ui.Suggestion, 0, len(studios))
	for _, studio := range studios {
		suggestions = append(suggestions, quotedSuggestion(studio.Name))
	}
	return suggestions
}

func performerSuggestions(performers []stash.Performer) []ui.Suggestion {
	suggestions := make([]ui.Suggestion, 0, len(performers))
	for _, performer := range performers {
		suggestions = append(suggestions, quotedSuggestion(performer.Name))
	}
	return suggestions
}

func currentSuggestion(prefix string) []ui.Suggestion {
	if strings.HasPrefix("current", prefix) {
		return []ui.Suggestion{{Display: "current", Value: "current"}}
	}
	return nil
}

func quotedSuggestion(value string) ui.Suggestion {
	suggestion := ui.Suggestion{
		Display: value,
		Value:   value,
	}
	if strings.ContainsRune(value, ' ') {
		suggestion.Value = `"` + value + `"`
	}
	return suggestion
}

type commandToken struct {
	index  int
	start  int
	end    int
	raw    string
	tokens []commandToken
}

func tokenAtCursor(input string, cursor int) (commandToken, bool) {
	tokens := tokenizeCommandInput(input)
	for i := range tokens {
		tokens[i].tokens = tokens
		tokens[i].index = i
		if cursor >= tokens[i].start && cursor <= tokens[i].end {
			return tokens[i], true
		}
	}
	return commandToken{}, false
}

func tokenizeCommandInput(input string) []commandToken {
	var tokens []commandToken
	inQuote := byte(0)
	start := -1
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if start == -1 {
			if unicode.IsSpace(rune(ch)) {
				continue
			}
			start = i
			if ch == '"' || ch == '\'' {
				inQuote = ch
			}
			continue
		}

		if inQuote != 0 {
			if ch == inQuote && input[i-1] != '\\' {
				inQuote = 0
			}
			continue
		}

		if ch == '"' || ch == '\'' {
			inQuote = ch
			continue
		}
		if unicode.IsSpace(rune(ch)) {
			tokens = append(tokens, commandToken{start: start, end: i, raw: input[start:i]})
			start = -1
		}
	}
	if start != -1 {
		tokens = append(tokens, commandToken{start: start, end: len(input), raw: input[start:]})
	}
	return tokens
}

func truncateLines(s string, height int) string {
	if height <= 0 {
		return ""
	}

	lines := strings.Split(s, "\n")
	if len(lines) <= height {
		return s
	}
	return strings.Join(lines[:height], "\n")
}
