package app

import (
	"fmt"
	"math"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/action"
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

// TODO probably it's the responsiblity of the parent to tell this model exactly how tall it is, so that it's not
// doing it's own math to solve this.
func (m *ScenesModel) SetHeight(height int) {
	m.pageState.PerPage = 0
	if height >= 5 {
		m.pageState.PerPage = height - 5
	}
}

func (m *ScenesModel) Init(size Size) tea.Cmd {
	m.screen = size
	m.SetHeight(size.Height)
	return m.updateCmd()
}

func (s *ScenesModel) Title() string {
	t := "Scenes"
	if s.query != "" {
		t = fmt.Sprintf("\"%s\"", s.query)
	} else if s.sceneFilter.Performers != nil {
		var performers []string
		for _, p := range s.sceneFilter.Performers.Value {
			perf, _ := s.StashLookup.GetPerformer(p)
			performers = append(performers, perf.Name)
		}
		t = strings.Join(performers, ", ")
	}

	return fmt.Sprintf("%c %s (%s)", '\U000f0fce', t, humanNumber(s.pageState.total))
}

func (s *ScenesModel) Current() stash.Scene {
	return s.scenes[s.pageState.index]
}

func (s *ScenesModel) PushState(mutate func(*ScenesModel)) (*ScenesModel, tea.Cmd) {
	s.history = append(s.history, sceneFilterState{
		query:         s.query,
		sort:          s.sort,
		sortDirection: s.sortDirection,
		sceneFilter:   s.sceneFilter,
		pageState:     s.pageState,
	})
	mutate(s)
	s.pageState.Reset()
	return s, s.updateCmd()
}

// Pop sets the current state to the previous state from the history stack.  If the history stack is empty this is a
// noop.
func (s *ScenesModel) Pop() (*ScenesModel, tea.Cmd) {
	if len(s.history) == 0 {
		return s, nil
	}

	state := s.history[len(s.history)-1]
	s.history = s.history[0 : len(s.history)-1]

	// Restore previous state, including pagination
	s.pageState = state.pageState
	s.query = state.query
	s.sort = state.sort
	s.sortDirection = state.sortDirection
	s.sceneFilter = state.sceneFilter
	s.scenes = []stash.Scene{}

	return s, s.updateCmd()
}

// Keymap allows mapping a keyboard shortcut to a command.  Commands are interpretted in command mode and do not take
// additional parameters.
var ScenesModelDefaultKeymap = map[string]string{
	"up":    "skip -1",
	"down":  "skip 1",
	"enter": "open skip=1",
	" ":     "open skip=1", // space
	"z":     "skip -1",
	"x":     "skip 1",
	"o":     "open",
	"r":     "sort random",
	"u":     "undo", // state pop?  Maybe some sort of generic state management command
	"f":     "filter favourite=1",
	"p":     "filter performer=current",
	"`":     "openurl stash",
}

// Command aliases can be used to alias useful commands.  This will act as a prefix for a command, meaning that
// additional inputs can be given after the alias.
var ScenesModelDefaultCommandAlias = map[string]string{
	"recent": "filter createdAt=>-24h",
	"year":   "filter date=>-1y",
}

func (m ScenesModel) Interpret(c Command) (tea.Msg, error) {
	switch c.Mode {
	case ModeFind:
		return ScenesModelFilterMsg{
			Query: &c.Input,
		}, nil

	default:
		a, err := action.Parse(c.Input)
		if err != nil {
			return nil, err
		}

		switch a.Name {
		case "filter":
			var msg ScenesModelFilterMsg
			err := a.Arguments.Bind(&msg)
			if err != nil {
				return nil, err
			}
			return msg, nil

		case "open":
			var msg ScenesModelOpenMsg
			err := a.Arguments.Bind(&msg)
			if err != nil {
				return nil, err
			}
			return msg, nil

		case "openurl":
			var msg ScenesModelOpenURLMsg
			err := a.Arguments.Bind(&msg)
			if err != nil {
				return nil, err
			}
			return msg, nil

		case "sort":
			var msg ScenesModelSortMsg
			err := a.Arguments.Bind(&msg)
			if err != nil {
				return nil, err
			}
			return msg, nil

		case "skip":
			var msg ScenesModelSkipMsg
			err := a.Arguments.Bind(&msg)
			if err != nil {
				return nil, err
			}
			return msg, nil

		case "undo":
			return ScenesModelUndoMsg{}, nil
		}
	}

	return nil, nil
}

// ScenesModelFilterMsg controls the filtering of various fields on the model. Currently this has a bit of a limitation
// in that although pointers can be used to determine if the user intended to set a field or not, there is no way
// currently to reset a filter.  I'm not sure the best method for this yet, possibly we need some sort of wrapper type
// that implements action.Binder
type ScenesModelFilterMsg struct {
	Query     *string
	Favourite *bool
	Performer *string
}

type ScenesModelOpenMsg struct {
	Skip bool
}

type ScenesModelOpenURLMsg struct {
	Source string
}

type ScenesModelSkipMsg struct {
	Count int
}

type ScenesModelSortMsg struct {
	Field     string
	Direction string
}

type ScenesModelUndoMsg struct{}

func (s ScenesModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ScenesModelFilterMsg:
		return s.PushState(func(sm *ScenesModel) {
			if msg.Query != nil {
				sm.query = *msg.Query
			}
			if msg.Favourite != nil {
				sm.sceneFilter.PerformerFavourite = msg.Favourite
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
						Value:    []string{*msg.Performer},
						Modifier: stash.CriterionModifierIncludes,
					}
				}
			}
		})

	case ScenesModelOpenMsg:
		if msg.Skip && s.pageState.Next() {
			return &s, s.updateCmd()
		}
		cur := s.Current()
		return &s, func() tea.Msg { return OpenMsg{cur} }

	case ScenesModelOpenURLMsg:
		cur := s.Current()
		var src string
		switch msg.Source {
		default:
			src = path.Join("scenes", cur.ID)
		}
		return &s, func() tea.Msg { return OpenMsg{src} }

	case ScenesModelSortMsg:
		switch msg.Field {
		case "random":
			return s.PushState(func(sm *ScenesModel) {
				sm.sort = stash.RandomSort()
			})
		}

	case ScenesModelSkipMsg:
		if s.pageState.Skip(msg.Count) {
			return &s, s.updateCmd()
		}

	case ScenesModelUndoMsg:
		return s.Pop()

	case tea.KeyMsg:
		// TODO this is probably not where this ends up, instead we probably have some additional part of the TabModel
		// interface that exposes keymaps (maybe).  I'll slot this in here now and it can return an execute command.
		if cmd, ok := ScenesModelDefaultKeymap[msg.String()]; ok {
			return &s, func() tea.Msg { return ui.CommandExecMsg{Command: cmd} }
		}

	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
		s.SetHeight(msg.Height)
		return &s, s.updateCmd()

	case action.Action:
		switch msg.Name {
		case "open":
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }

		case "filter":
			var dst struct {
				Query string
			}
			if err := msg.Arguments.Bind(&dst); err != nil {
				return &s, NewErrorCmd(err)
			}
			return s.PushState(func(sm *ScenesModel) {
				sm.query = dst.Query
			})

		case "sort":
			var dst struct {
				Field string
			}
			if err := msg.Arguments.Bind(&dst); err != nil {
				return &s, NewErrorCmd(err)
			}

			switch dst.Field {
			case "date":
				return s.PushState(func(sm *ScenesModel) {
					sm.sort = dst.Field
					sm.sortDirection = stash.SortDirectionAsc
				})

			case "-date":
				return s.PushState(func(sm *ScenesModel) {
					sm.sort = dst.Field[1:]
					sm.sortDirection = stash.SortDirectionDesc
				})
			}

		case "random":
			return s.PushState(func(sm *ScenesModel) {
				sm.sort = stash.RandomSort()
			})

		case "recent":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.CreatedAt = &stash.TimestampCriterion{
					Value:    time.Now().Add(-24 * time.Hour * 7),
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "year":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Date = &stash.DateCriterion{
					Value:    time.Now().Add(-24 * time.Hour * 365),
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "before":
			var dst struct {
				Date time.Time
			}
			if err := msg.Arguments.Bind(&dst); err != nil {
				return &s, NewErrorCmd(err)
			}

			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Date = &stash.DateCriterion{
					Value:    dst.Date,
					Modifier: stash.CriterionModifierLessThan,
				}
			})

		case "6mo":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.CreatedAt = &stash.TimestampCriterion{
					Value:    time.Now().Add(-24 * time.Hour * 182),
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "1mo":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.CreatedAt = &stash.TimestampCriterion{
					Value:    time.Now().Add(-24 * time.Hour * 30),
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "pop":
			return s.Pop()

		case "duration":
			var dst struct {
				Seconds int
			}
			if err := msg.Arguments.Bind(&dst); err != nil {
				return &s, NewErrorCmd(err)
			}
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Duration = &stash.IntCriterion{
					Value:    dst.Seconds,
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "reset":
			return &s, s.reset()

		case "refresh":
			return &s, s.updateCmd()

		case "organised", "organized":
			var dst struct {
				Organised *bool
			}
			if err := msg.Arguments.Bind(&dst); err != nil {
				return &s, NewErrorCmd(err)
			}
			return s.PushState(func(gm *ScenesModel) {
				if dst.Organised != nil {
					gm.sceneFilter.Organized = dst.Organised
				} else {
					organised := true
					gm.sceneFilter.Organized = &organised
				}
			})

		case "studio":
			// TODO check for unconsumed positional arguments
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Studios = &stash.HierarchicalMultiCriterion{
					Value:    msg.Arguments.Positional(),
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "tags":
			// TODO check for unconsumed positional arguments
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Tags = &stash.HierarchicalMultiCriterion{
					Value:    msg.Arguments.Positional(),
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "pt":
			// TODO check for unconsumed positional arguments
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.PerformerTags = &stash.HierarchicalMultiCriterion{
					Value:    msg.Arguments.Positional(),
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "favourite", "favorite":
			var dst struct {
				Favourite *bool
			}
			if err := msg.Arguments.Bind(&dst); err != nil {
				return &s, NewErrorCmd(err)
			}
			return s.PushState(func(gm *ScenesModel) {
				if dst.Favourite != nil {
					gm.sceneFilter.PerformerFavourite = dst.Favourite
				} else {
					favourite := true
					gm.sceneFilter.PerformerFavourite = &favourite
				}
			})

		case "rated":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Rating100 = &stash.IntCriterion{
					Modifier: stash.CriterionModifierNotNull,
				}
			})

		case "stash":
			msg := OpenMsg{path.Join("scenes", s.Current().ID)}
			return &s, func() tea.Msg { return msg }

		case "delete":
			return &s, s.doDeleteConfirmCmd()
		}

	case scenesMsg:
		s.scenes, s.pageState.total = msg.scenes, msg.total

	case DeleteMsg:
		return &s, s.deleteCmd(msg.Scene.ID)
	}

	return &s, nil
}

func (s ScenesModel) View() string {
	var rows []ui.Row
	for i, scene := range s.scenes {
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
		if s.pageState.index == i {
			rows[i].Background = &ColorRowSelected
		}
	}

	leftStatus := []string{
		s.pageState.String(),
		sort(s.sort, s.sortDirection),
	}

	rightStatus := sceneFilterStatus(s.sceneFilter, s.StashLookup)
	if s.query != "" {
		rightStatus = append(rightStatus, "\""+s.query+"\"")
	}
	if len(s.history) > 0 {
		rightStatus = append(rightStatus, fmt.Sprintf("[%d]", len(s.history)))
	}

	return lipgloss.JoinVertical(0,
		statusBar.Render(s.screen.Width, leftStatus, rightStatus),
		sceneTable.Render(s.screen.Width, rows),
	)
}

// updateCmd sets initial loading state then returns a tea.Cmd to execute loading of scenes.
func (m *ScenesModel) updateCmd() tea.Cmd {
	return m.SceneService.Scenes(stash.FindFilter{
		Query:     m.query,
		Page:      m.pageState.page + 1,
		PerPage:   m.pageState.PerPage,
		Sort:      m.sort,
		Direction: m.sortDirection,
	}, m.sceneFilter)
}

// doDeleteConfirmCmd returns a command to display a confirmation message about the current content.
func (s *ScenesModel) doDeleteConfirmCmd() tea.Cmd {
	return nil
	// TODO reimplement deletion once app modal confirmation is in a better state
	// return func() tea.Msg {
	// 	s := s.Current()
	// 	titleStyle := lipgloss.NewStyle().
	// 		Foreground(ColorOffWhite)
	// 	return ConfirmationMsg{
	// 		Message:       fmt.Sprintf("Are you sure you want to delete %s?", titleStyle.Render(sceneTitle(s))),
	// 		ConfirmOption: "Delete",
	// 		CancelOption:  "Cancel",
	// 		Cmd: func() tea.Msg {
	// 			return DeleteMsg{Scene: s}
	// 		},
	// 	}
	// }
}

type DeleteMsg struct {
	Scene stash.Scene
}

// doDeleteCmd takes a DeleteMessage and attempts to delete the provided scene.  After successful deletion the current
// scenes data is refreshed.
func (m *ScenesModel) deleteCmd(id string) tea.Cmd {
	return m.SceneService.DeleteScene(id)
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
