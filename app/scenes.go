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

	pageState paginator
}

type SceneService interface {
	Scenes(stash.FindFilter, stash.SceneFilter) tea.Cmd
	DeleteScene(string) tea.Cmd
}

type ScenesModel struct {
	SceneService
	StashLookup

	paginator
	scenes []stash.Scene

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
	m.reset()
	return m
}

func (s *ScenesModel) reset() tea.Cmd {
	s.paginator = NewPaginator(40)

	s.query = ""
	s.sort = stash.SortDate
	s.sortDirection = stash.SortDirectionDesc
	s.sceneFilter = stash.SceneFilter{}

	return s.updateCmd()
}

func (s *ScenesModel) Init(size Size) tea.Cmd {
	s.screen = size
	return s.updateCmd()
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

	return fmt.Sprintf("%c %s (%s)", '\U000f0fce', t, humanNumber(s.total))
}

func (s *ScenesModel) Current() stash.Scene {
	return s.scenes[s.index]
}

func (s *ScenesModel) PushState(mutate func(*ScenesModel)) (*ScenesModel, tea.Cmd) {
	s.history = append(s.history, sceneFilterState{
		query:         s.query,
		sort:          s.sort,
		sortDirection: s.sortDirection,
		sceneFilter:   s.sceneFilter,
		pageState:     s.paginator,
	})
	mutate(s)
	s.paginator.Reset()
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
	s.paginator = state.pageState
	s.query = state.query
	s.sort = state.sort
	s.sortDirection = state.sortDirection
	s.sceneFilter = state.sceneFilter
	s.scenes = []stash.Scene{}

	return s, s.updateCmd()
}

func (s ScenesModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if s.Skip(-1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case tea.KeyDown:
			if s.Skip(1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case tea.KeyEnter, tea.KeySpace:
			if s.Next() {
				s.Clear()
				return &s, s.updateCmd()
			}
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }
		}

		switch msg.String() {
		case "z":
			if s.Skip(-1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case "x":
			if s.Skip(1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case "/":
			return &s, NewModeCommandCmd("/", "filter ")
		case "o":
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }
		case "r":
			return s.PushState(func(sm *ScenesModel) {
				sm.sort = stash.RandomSort()
			})
		case "u": // "Undo"
			return s.Pop()
		case "f":
			return s.PushState(func(sm *ScenesModel) {
				if sm.sceneFilter.PerformerFavourite == nil {
					favourite := true
					sm.sceneFilter.PerformerFavourite = &favourite
				} else {
					sm.sceneFilter.PerformerFavourite = nil
				}
			})
		case "p":
			return s.newTabPerformerCmd()
		case "`":
			msg := OpenMsg{path.Join("scenes", s.Current().ID)}
			return &s, func() tea.Msg { return msg }
		}

	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

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

		case "more":
			return s.newTabPerformerCmd()

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
		s.scenes, s.total = msg.scenes, msg.total

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
		if s.index == i {
			rows[i].Background = &ColorRowSelected
		}
	}

	leftStatus := []string{
		s.paginator.String(),
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
		Page:      m.page,
		PerPage:   m.perPage,
		Sort:      m.sort,
		Direction: m.sortDirection,
	}, m.sceneFilter)
}

// newTabPerformerCmd returns a command that opens a new tab filtered to the current set of performers
func (m *ScenesModel) newTabPerformerCmd() (*ScenesModel, tea.Cmd) {
	if len(m.Current().Performers) == 0 {
		return m, nil
	}

	var ids []string
	for _, p := range m.Current().Performers {
		ids = append(ids, p.ID)
	}
	return m, func() tea.Msg {
		return TabOpenMsg{
			tabFunc: func() TabModel {
				t := NewScenesModel(m.SceneService, m.StashLookup)
				t.sceneFilter.Performers = &stash.MultiCriterion{
					Value:    ids,
					Modifier: stash.CriterionModifierIncludes,
				}
				return t
			},
		}
	}
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
