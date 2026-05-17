package app

import (
	"fmt"
	"math"
	"path"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/command"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type sceneFilterState struct {
	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter

	pageState pageState
}

type SceneService interface {
	Scenes(stash.FindFilter, stash.SceneFilter) tea.Cmd
	DeleteScene(string) tea.Cmd
	TagScene(stash.Scene, []string) tea.Cmd
	ResolveTags([]string) tea.Cmd
	ResolveStudios([]string) tea.Cmd
	ResolvePerformers([]string) tea.Cmd
}

type ScenesModel struct {
	SceneService
	StashLookup

	pageState pageState
	scenes    []stash.Scene

	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter

	history []sceneFilterState

	screen Size

	pendingFilterRequestID uint64
	pendingFilter          *pendingSceneFilter
}

type pendingSceneFilter struct {
	requestID       uint64
	msg             ScenesModelFilterMsg
	tagIDs          []string
	studioIDs       []string
	performerIDs    []string
	performerTagIDs []string
	waitingOn       int
}

func NewScenesModel(sceneService SceneService, lookup StashLookup) *ScenesModel {
	m := &ScenesModel{
		SceneService: sceneService,
		StashLookup:  lookup,
	}
	m.pageState.PerPage = 5
	m.reset()
	return m
}

func (m *ScenesModel) reset() tea.Cmd {
	m.query = ""
	m.sort = stash.SortDate
	m.sortDirection = stash.SortDirectionDesc
	m.sceneFilter = stash.SceneFilter{}
	m.pageState.Reset()

	return m.updateCmd()
}

// SetSize takes a size input that indicates the size of the tab being rendered.
func (m *ScenesModel) SetSize(s Size) tea.Cmd {
	m.screen = s
	m.pageState.PerPage = (s.Height - 1) // account for status line
	return m.updateCmd()
}

func (m *ScenesModel) Init() tea.Cmd {
	return m.updateCmd()
}

func (m *ScenesModel) Title() string {
	t := "Scenes"
	if m.query != "" {
		t = fmt.Sprintf("\"%s\"", m.query)
	} else if m.sceneFilter.Performers != nil {
		var performers []string
		for _, p := range m.sceneFilter.Performers.Value {
			perf, _ := m.StashLookup.GetPerformer(p)
			performers = append(performers, perf.Name)
		}
		t = strings.Join(performers, ", ")
	}

	return fmt.Sprintf("%c %s (%s)", '\U000f0fce', t, humanNumber(m.pageState.total))
}

func (m *ScenesModel) Current() stash.Scene {
	return m.scenes[m.pageState.index]
}

func (m *ScenesModel) PushState(mutate func(*ScenesModel)) (*ScenesModel, tea.Cmd) {
	m.history = append(m.history, sceneFilterState{
		query:         m.query,
		sort:          m.sort,
		sortDirection: m.sortDirection,
		sceneFilter:   m.sceneFilter,
		pageState:     m.pageState,
	})
	mutate(m)
	m.pageState.Reset()
	return m, m.updateCmd()
}

// Pop sets the current state to the previous state from the history stack.  If the history stack is empty this is a
// noop.
func (m *ScenesModel) Pop() (*ScenesModel, tea.Cmd) {
	if len(m.history) == 0 {
		return m, nil
	}

	state := m.history[len(m.history)-1]
	m.history = m.history[0 : len(m.history)-1]

	// Restore previous state, including pagination
	m.pageState = state.pageState
	m.query = state.query
	m.sort = state.sort
	m.sortDirection = state.sortDirection
	m.sceneFilter = state.sceneFilter
	m.scenes = []stash.Scene{}

	return m, m.updateCmd()
}

// Keymap allows mapping a keyboard shortcut to a command.  Commands are interpretted in command mode and do not take
// additional parameters.
var ScenesModelDefaultKeymap = map[string]string{
	"up":    "skip -1",
	"down":  "skip 1",
	"D":     "delete",
	"enter": "open skip",
	" ":     "open skip", // space
	"z":     "skip -1",
	"x":     "skip 1",
	"o":     "open",
	"r":     "sort random",
	"u":     "undo", // state pop?  Maybe some sort of generic state management command
	"f":     "filter favourite=1",
	"p":     "filter performer=current",
	"`":     "open-url source=stash",
}

// Command aliases can be used to alias useful commands.  This will act as a prefix for a command, meaning that
// additional inputs can be given after the alias.
var ScenesModelDefaultCommandAlias = map[string]string{
	"recent": "filter created=>-24h",
	"year":   "filter date=>-1y",
}

var ScenesModelCommandConfig command.Config = command.Config{
	"delete":   binder[ScenesModelDeleteMsg](),
	"filter":   binder[ScenesModelFilterMsg](),
	"open":     binder[ScenesModelOpenMsg](),
	"open-url": binder[ScenesModelOpenURLMsg](),
	"refresh":  binder[ScenesModelRefresh](),
	"reset":    binder[ScenesModelResetMsg](),
	"sort":     binder[ScenesModelSortMsg](),
	"skip":     binder[ScenesModelSkipMsg](),
	"tag":      binder[ScenesModelTagMsg](),
	"undo":     binder[ScenesModelUndoMsg](),
}

func (m ScenesModel) CommandConfig() command.Config {
	return ScenesModelCommandConfig
}

func (m ScenesModel) Search(query string) tea.Msg {
	return ScenesModelFilterMsg{
		Query: &query,
	}
}

// ScenesModelFilterMsg controls the filtering of various fields on the model. Currently this has a bit of a limitation
// in that although pointers can be used to determine if the user intended to set a field or not, there is no way
// currently to reset a filter.
type ScenesModelFilterMsg struct {
	Query        *string
	Favourite    *bool
	Organised    *bool
	Rating       *int
	Date         *dateFilterValue
	Created      *dateFilterValue `command:"created"`
	Updated      *dateFilterValue `command:"updated"`
	Performer    *string
	Duration     *int
	PerformerTag *string
	Tag          []string
	Studio       []string
}

type sceneTagsResolvedMsg struct {
	requestID uint64
	ids       []string
}

type sceneStudiosResolvedMsg struct {
	requestID uint64
	ids       []string
}

type scenePerformersResolvedMsg struct {
	requestID uint64
	ids       []string
}

type scenePerformerTagsResolvedMsg struct {
	requestID uint64
	ids       []string
}

type resolvedSceneFilterIDs struct {
	tagIDs          []string
	studioIDs       []string
	performerIDs    []string
	performerTagIDs []string
}

type ScenesModelOpenMsg struct {
	Skip bool `command:",positional"`
}

type ScenesModelOpenURLMsg struct {
	Source string
}

type ScenesModelDeleteMsg struct {
	Confirm bool
}

type ScenesModelTagMsg struct {
	Tags []string `command:",positional"`
}

type ScenesModelRefresh struct{}

type ScenesModelResetMsg struct{}

type ScenesModelSkipMsg struct {
	Count int `command:",positional"`
}

type ScenesModelSortMsg struct {
	Field     string `command:",positional"`
	Direction string
}

type ScenesModelUndoMsg struct{}

func (m *ScenesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ScenesModelFilterMsg:
		if m.filterNeedsAsyncResolution(msg) {
			return m.beginPendingFilter(msg)
		}
		return m.applyFilter(msg, resolvedSceneFilterIDs{
			tagIDs:          maybeIDs(msg.Tag),
			studioIDs:       maybeIDs(msg.Studio),
			performerIDs:    maybeSingleID(msg.Performer),
			performerTagIDs: maybeSingleID(msg.PerformerTag),
		})

	case sceneTagsResolvedMsg:
		if m.pendingFilter == nil || m.pendingFilter.requestID != msg.requestID {
			return m, nil
		}
		m.pendingFilter.tagIDs = msg.ids
		m.pendingFilter.waitingOn--
		if m.pendingFilter.waitingOn > 0 {
			return m, nil
		}

		pending := m.pendingFilter
		m.pendingFilter = nil
		return m.applyFilter(pending.msg, resolvedSceneFilterIDs{
			tagIDs:          pending.tagIDs,
			studioIDs:       pending.studioIDs,
			performerIDs:    pending.performerIDs,
			performerTagIDs: pending.performerTagIDs,
		})

	case sceneStudiosResolvedMsg:
		if m.pendingFilter == nil || m.pendingFilter.requestID != msg.requestID {
			return m, nil
		}
		m.pendingFilter.studioIDs = msg.ids
		m.pendingFilter.waitingOn--
		if m.pendingFilter.waitingOn > 0 {
			return m, nil
		}
		pending := m.pendingFilter
		m.pendingFilter = nil
		return m.applyFilter(pending.msg, resolvedSceneFilterIDs{
			tagIDs:          pending.tagIDs,
			studioIDs:       pending.studioIDs,
			performerIDs:    pending.performerIDs,
			performerTagIDs: pending.performerTagIDs,
		})

	case scenePerformersResolvedMsg:
		if m.pendingFilter == nil || m.pendingFilter.requestID != msg.requestID {
			return m, nil
		}
		m.pendingFilter.performerIDs = msg.ids
		m.pendingFilter.waitingOn--
		if m.pendingFilter.waitingOn > 0 {
			return m, nil
		}
		pending := m.pendingFilter
		m.pendingFilter = nil
		return m.applyFilter(pending.msg, resolvedSceneFilterIDs{
			tagIDs:          pending.tagIDs,
			studioIDs:       pending.studioIDs,
			performerIDs:    pending.performerIDs,
			performerTagIDs: pending.performerTagIDs,
		})

	case scenePerformerTagsResolvedMsg:
		if m.pendingFilter == nil || m.pendingFilter.requestID != msg.requestID {
			return m, nil
		}
		m.pendingFilter.performerTagIDs = msg.ids
		m.pendingFilter.waitingOn--
		if m.pendingFilter.waitingOn > 0 {
			return m, nil
		}
		pending := m.pendingFilter
		m.pendingFilter = nil
		return m.applyFilter(pending.msg, resolvedSceneFilterIDs{
			tagIDs:          pending.tagIDs,
			studioIDs:       pending.studioIDs,
			performerIDs:    pending.performerIDs,
			performerTagIDs: pending.performerTagIDs,
		})

	case ScenesModelOpenMsg:
		if msg.Skip && m.pageState.Next() {
			return m, m.updateCmd()
		}
		cur := m.Current()
		return m, func() tea.Msg { return OpenMsg{cur} }

	case ScenesModelOpenURLMsg:
		cur := m.Current()
		var src string
		switch msg.Source {
		default:
			src = path.Join("scenes", cur.ID)
		}
		return m, func() tea.Msg { return OpenMsg{src} }

	case ScenesModelDeleteMsg:
		if len(m.scenes) == 0 {
			return m, NewErrorCmd(fmt.Errorf("no scene selected"))
		}
		scene := m.Current()
		return m, func() tea.Msg {
			return deleteRequestMsg{
				Entity:      "scene",
				Title:       sceneTitle(scene),
				Path:        scene.FilePath(),
				SkipConfirm: msg.Confirm,
				DeleteCmd:   m.SceneService.DeleteScene(scene.ID),
			}
		}

	case ScenesModelTagMsg:
		if len(m.scenes) == 0 {
			return m, NewErrorCmd(fmt.Errorf("no scene selected"))
		}
		if len(msg.Tags) == 0 {
			return m, NewErrorCmd(fmt.Errorf("no tags specified"))
		}
		return m, m.SceneService.TagScene(m.Current(), msg.Tags)

	case ScenesModelRefresh:
		return m, m.updateCmd()

	case ScenesModelResetMsg:
		return m, m.reset()

	case ScenesModelSortMsg:
		switch msg.Field {
		case "random":
			return m.PushState(func(sm *ScenesModel) {
				sm.sort = stash.RandomSort()
			})
		case "date":
			return m.PushState(func(sm *ScenesModel) {
				sm.sort = "date"
				sm.sortDirection = stash.SortDirectionAsc
			})
		// TODO it's probably the case that we want to parse this in the Interpret step rather than here.  We can just
		// enumerate fields we're interested in for the time being.
		case "-date":
			return m.PushState(func(sm *ScenesModel) {
				sm.sort = "date"
				sm.sortDirection = stash.SortDirectionDesc
			})
		}

	case ScenesModelSkipMsg:
		if m.pageState.Skip(msg.Count) {
			return m, m.updateCmd()
		}

	case ScenesModelUndoMsg:
		return m.Pop()

	case tea.KeyMsg:
		// TODO this is probably not where this ends up, instead we probably have some additional part of the TabModel
		// interface that exposes keymaps (maybe).  I'll slot this in here now and it can return an execute command.
		if cmd, ok := ScenesModelDefaultKeymap[msg.String()]; ok {
			return m, func() tea.Msg { return ui.CommandExecMsg{Command: cmd} }
		}

	case scenesMsg:
		m.scenes, m.pageState.total = msg.scenes, msg.total

	case sceneDeletedMsg:
		m.pageState.DeleteCurrent()
		return m, m.updateCmd()

	case sceneTaggedMsg:
		if len(m.scenes) > 0 {
			m.scenes[m.pageState.index] = msg.scene
		}
	}

	return m, nil
}

func (m *ScenesModel) filterNeedsAsyncResolution(msg ScenesModelFilterMsg) bool {
	return needsEntityResolution(msg.Tag) ||
		needsEntityResolution(msg.Studio) ||
		needsSingleEntityResolution(msg.Performer) ||
		needsSingleEntityResolution(msg.PerformerTag)
}

func (m *ScenesModel) beginPendingFilter(msg ScenesModelFilterMsg) (*ScenesModel, tea.Cmd) {
	requestID := atomic.AddUint64(&m.pendingFilterRequestID, 1)
	pending := &pendingSceneFilter{
		requestID: requestID,
		msg:       msg,
	}

	var cmds []tea.Cmd
	if len(msg.Tag) > 0 {
		pending.waitingOn++
		cmds = append(cmds, m.resolveSceneTagsCmd(requestID, msg.Tag))
	}
	if needsEntityResolution(msg.Studio) {
		pending.waitingOn++
		cmds = append(cmds, m.resolveSceneStudiosCmd(requestID, msg.Studio))
	}
	if needsSingleEntityResolution(msg.Performer) {
		pending.waitingOn++
		cmds = append(cmds, m.resolveScenePerformersCmd(requestID, []string{*msg.Performer}))
	}
	if needsSingleEntityResolution(msg.PerformerTag) {
		pending.waitingOn++
		cmds = append(cmds, m.resolveScenePerformerTagsCmd(requestID, []string{*msg.PerformerTag}))
	}

	m.pendingFilter = pending
	return m, tea.Batch(cmds...)
}

func (m *ScenesModel) resolveSceneTagsCmd(requestID uint64, rawTags []string) tea.Cmd {
	tags := append([]string(nil), rawTags...)
	return func() tea.Msg {
		resolved := m.SceneService.ResolveTags(tags)()
		switch msg := resolved.(type) {
		case resolvedTagIDsMsg:
			return sceneTagsResolvedMsg{requestID: requestID, ids: msg.ids}
		case loadingMsg:
			if payload, ok := msg.payload.(resolvedTagIDsMsg); ok {
				msg.payload = sceneTagsResolvedMsg{requestID: requestID, ids: payload.ids}
			}
			return msg
		default:
			return resolved
		}
	}
}

func (m *ScenesModel) resolveSceneStudiosCmd(requestID uint64, rawStudios []string) tea.Cmd {
	studios := append([]string(nil), rawStudios...)
	return func() tea.Msg {
		resolved := m.SceneService.ResolveStudios(studios)()
		switch msg := resolved.(type) {
		case resolvedStudioIDsMsg:
			return sceneStudiosResolvedMsg{requestID: requestID, ids: msg.ids}
		case loadingMsg:
			if payload, ok := msg.payload.(resolvedStudioIDsMsg); ok {
				msg.payload = sceneStudiosResolvedMsg{requestID: requestID, ids: payload.ids}
			}
			return msg
		default:
			return resolved
		}
	}
}

func (m *ScenesModel) resolveScenePerformersCmd(requestID uint64, rawPerformers []string) tea.Cmd {
	performers := append([]string(nil), rawPerformers...)
	return func() tea.Msg {
		resolved := m.SceneService.ResolvePerformers(performers)()
		switch msg := resolved.(type) {
		case resolvedPerformerIDsMsg:
			return scenePerformersResolvedMsg{requestID: requestID, ids: msg.ids}
		case loadingMsg:
			if payload, ok := msg.payload.(resolvedPerformerIDsMsg); ok {
				msg.payload = scenePerformersResolvedMsg{requestID: requestID, ids: payload.ids}
			}
			return msg
		default:
			return resolved
		}
	}
}

func (m *ScenesModel) resolveScenePerformerTagsCmd(requestID uint64, rawTags []string) tea.Cmd {
	tags := append([]string(nil), rawTags...)
	return func() tea.Msg {
		resolved := m.SceneService.ResolveTags(tags)()
		switch msg := resolved.(type) {
		case resolvedTagIDsMsg:
			return scenePerformerTagsResolvedMsg{requestID: requestID, ids: msg.ids}
		case loadingMsg:
			if payload, ok := msg.payload.(resolvedTagIDsMsg); ok {
				msg.payload = scenePerformerTagsResolvedMsg{requestID: requestID, ids: payload.ids}
			}
			return msg
		default:
			return resolved
		}
	}
}

func (m *ScenesModel) applyFilter(msg ScenesModelFilterMsg, resolved resolvedSceneFilterIDs) (*ScenesModel, tea.Cmd) {
	return m.PushState(func(sm *ScenesModel) {
		if msg.Query != nil {
			sm.query = *msg.Query
		}
		if msg.Favourite != nil {
			sm.sceneFilter.PerformerFavourite = msg.Favourite
		}
		if msg.Organised != nil {
			sm.sceneFilter.Organized = msg.Organised
		}
		if msg.Rating != nil {
			sm.sceneFilter.Rating100 = &stash.IntCriterion{
				Modifier: stash.CriterionModifierEquals,
				Value:    *msg.Rating,
			}
		}
		if msg.Date != nil {
			sm.sceneFilter.Date = msg.Date.DateCriterion()
		}
		if msg.Created != nil {
			sm.sceneFilter.CreatedAt = msg.Created.TimestampCriterion()
		}
		if msg.Updated != nil {
			sm.sceneFilter.UpdatedAt = msg.Updated.TimestampCriterion()
		}
		if msg.Performer != nil {
			if *msg.Performer == "current" {
				var ids []string
				for _, p := range sm.Current().Performers {
					ids = append(ids, p.ID)
				}
				sm.sceneFilter.Performers = &stash.MultiCriterion{
					Value:    ids,
					Modifier: stash.CriterionModifierIncludes,
				}
			} else {
				sm.sceneFilter.Performers = &stash.MultiCriterion{
					Value:    fallbackIDs(resolved.performerIDs, []string{*msg.Performer}),
					Modifier: stash.CriterionModifierIncludes,
				}
			}
		}
		if len(msg.Studio) > 0 || len(resolved.studioIDs) > 0 {
			sm.sceneFilter.Studios = &stash.HierarchicalMultiCriterion{
				Value:    fallbackIDs(resolved.studioIDs, msg.Studio),
				Modifier: stash.CriterionModifierIncludes,
			}
		}
		if len(msg.Tag) > 0 || len(resolved.tagIDs) > 0 {
			sm.sceneFilter.Tags = &stash.HierarchicalMultiCriterion{
				Value:    fallbackIDs(resolved.tagIDs, msg.Tag),
				Modifier: stash.CriterionModifierIncludes,
			}
		}
		if msg.PerformerTag != nil {
			sm.sceneFilter.PerformerTags = &stash.HierarchicalMultiCriterion{
				Value:    fallbackIDs(resolved.performerTagIDs, []string{*msg.PerformerTag}),
				Modifier: stash.CriterionModifierIncludes,
			}
		}
		if msg.Duration != nil {
			sm.sceneFilter.Duration = &stash.IntCriterion{
				Value:    *msg.Duration,
				Modifier: stash.CriterionModifierGreaterThan,
			}
		}
	})
}

func (m ScenesModel) View() string {
	var rows []ui.Row
	for i, scene := range m.scenes {
		rows = append(rows, ui.Row{
			Values: []string{
				organised(scene.Organized),
				scene.Date,
				rating(scene.Rating),
				sceneTitle(scene),
				sceneSize(scene),
				scene.Studio.Name,
				performerList(scene.Performers),
				tagList(scene.Tags),
				details(scene.Details),
			},
		})
		if m.pageState.index == i {
			rows[i].Background = &ColorRowSelected
		}
	}

	leftStatus := []string{
		m.pageState.String(),
		sort(m.sort, m.sortDirection),
	}

	rightStatus := sceneFilterStatus(m.sceneFilter, m.StashLookup)
	if m.query != "" {
		rightStatus = append(rightStatus, "\""+m.query+"\"")
	}
	if len(m.history) > 0 {
		rightStatus = append(rightStatus, fmt.Sprintf("[%d]", len(m.history)))
	}

	return lipgloss.JoinVertical(0,
		statusBar.Render(m.screen.Width, leftStatus, rightStatus),
		sceneTable.Render(m.screen.Width, rows),
	)
}

// updateCmd sets initial loading state then returns a tea.Cmd to execute loading of scenes.
func (m *ScenesModel) updateCmd() tea.Cmd {
	if m.pageState.PerPage == 0 {
		return nil
	}
	return m.SceneService.Scenes(stash.FindFilter{
		Query:     m.query,
		Page:      m.pageState.page + 1,
		PerPage:   m.pageState.PerPage,
		Sort:      m.sort,
		Direction: m.sortDirection,
	}, m.sceneFilter)
}

var (
	sceneTable = &ui.Table{
		AltBackground: ColorBlack,
		Cols: []ui.Column{
			{
				Name: "Organised",
			},
			{
				Name:       "Date",
				Foreground: &ColorGrey,
			},
			{
				Name: "Rating",
			},
			{
				Name:       "Title",
				Foreground: &ColorOffWhite,
				Bold:       true,
				Weight:     1,
			},
			{
				Name:       "Size",
				Foreground: &ColorBlue,
				Align:      lipgloss.Right,
			},
			{
				Name:       "Studio",
				Foreground: &ColorSalmon,
				Weight:     1,
			},
			{
				Name:       "Perfomers",
				Foreground: &ColorYellow,
				Weight:     1,
			},
			{
				Name:       "Tags",
				Foreground: &ColorPurple,
				Weight:     1,
			},
			{
				Name:       "Description",
				Foreground: &ColorGrey,
				Flex:       true,
			},
		},
	}
)

func rating(value int) string {
	if value <= 0 {
		return ""
	}
	if value > 100 {
		value = 100
	}
	stars := int(math.Ceil(float64(value) * 5.0 / 100.0))
	return fmt.Sprintf("%d\U000f04ce", stars)
}
