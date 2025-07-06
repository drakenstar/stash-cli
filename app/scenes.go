package app

import (
	"context"
	"fmt"
	"math"
	"path"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type filterState struct {
	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter

	pageState paginator
}

type ScenesModel struct {
	stash  stashCache
	Opener config.Opener

	paginator
	loading bool
	spinner spinner.Model
	scenes  []stash.Scene

	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter

	history []filterState

	screen Size
}

func NewScenesModel(stash stash.Stash, opener config.Opener) *ScenesModel {
	s := &ScenesModel{
		stash:  newStashCache(stash),
		Opener: opener,
	}

	s.spinner = spinner.New(spinner.WithSpinner(spinner.Globe))

	return s
}

func (s *ScenesModel) Init(size Size) tea.Cmd {
	s.paginator = NewPaginator(40)

	s.query = ""
	s.sort = stash.SortDate
	s.sortDirection = stash.SortDirectionDesc
	s.sceneFilter = stash.SceneFilter{}

	s.screen = size

	return tea.Batch(
		s.doUpdateCmd(),
		s.spinner.Tick,
	)
}

func (s *ScenesModel) TabTitle() string {
	return "Scenes"
}

func (s *ScenesModel) Current() stash.Scene {
	return s.scenes[s.index]
}

func (s *ScenesModel) PushState(mutate func(*ScenesModel)) (*ScenesModel, tea.Cmd) {
	s.history = append(s.history, filterState{
		query:         s.query,
		sort:          s.sort,
		sortDirection: s.sortDirection,
		sceneFilter:   s.sceneFilter,
		pageState:     s.paginator,
	})
	mutate(s)
	s.paginator.Reset()
	return s, s.doUpdateCmd()
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

	return s, s.doUpdateCmd()
}

func (s ScenesModel) Update(msg tea.Msg) (AppModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if !s.loading && s.Skip(-1) {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
		case tea.KeyDown:
			if !s.loading && s.Skip(1) {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
		}

	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

	case Command:
		switch msg.Name() {
		case "":
			if s.Next() {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
			s.Opener(s.Current())

		case "open", "o":
			s.Opener(s.Current())

		case "filter", "f":
			return s.PushState(func(sm *ScenesModel) {
				sm.query = msg.ArgString()
			})

		case "sort":
			sort := msg.ArgString()
			switch sort {
			case "date":
				return s.PushState(func(sm *ScenesModel) {
					sm.sort = sort
					sm.sortDirection = stash.SortDirectionAsc
				})

			case "-date":
				return s.PushState(func(sm *ScenesModel) {
					sm.sort = sort[1:]
					sm.sortDirection = stash.SortDirectionDesc
				})
			}

		case "studio":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Studios = &stash.HierarchicalMultiCriterion{
					Value:    msg.Args(),
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "random", "r":
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
			val, err := msg.ArgInt()
			if err != nil {
				return &s, NewErrorCmd(err)
			}
			t := time.Date(val, time.January, 1, 0, 0, 0, 0, time.UTC)
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Date = &stash.DateCriterion{
					Value:    t,
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

		case "pop", "p":
			return s.Pop()

		case "duration":
			val, err := msg.ArgInt()
			if err != nil {
				return &s, NewErrorCmd(err)
			}
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Duration = &stash.IntCriterion{
					Value:    val,
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "reset":
			return &s, s.Init(s.screen)

		case "refresh":
			return &s, s.doUpdateCmd()

		case "organised", "organized":
			organised := true
			if msg.ArgString() == "false" {
				organised = false
			}
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Organized = &organised
			})

		case "tags":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Tags = &stash.HierarchicalMultiCriterion{
					Value:    []string{msg.ArgString()},
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "pt":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.PerformerTags = &stash.HierarchicalMultiCriterion{
					Value:    []string{msg.ArgString()},
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "favourite", "favorite":
			favourite := true
			if msg.ArgString() == "false" {
				favourite = false
			}
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.PerformerFavourite = &favourite
			})

		case "more":
			var ids []string
			for _, p := range s.Current().Performers {
				ids = append(ids, p.ID)
			}
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Performers = &stash.MultiCriterion{
					Value:    ids,
					Modifier: stash.CriterionModifierIncludes,
				}
			})

		case "rated":
			return s.PushState(func(sm *ScenesModel) {
				sm.sceneFilter.Rating100 = &stash.IntCriterion{
					Modifier: stash.CriterionModifierNotNull,
				}
			})

		case "stash":
			s.Opener(path.Join("scenes", s.Current().ID))

		case "delete":
			return &s, s.doDeleteConfirmCmd()
		}

	case scenesMessage:
		s.loading = false
		if msg.err != nil {
			return &s, NewErrorCmd(msg.err)
		}
		s.scenes, s.total = msg.scenes, msg.total

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return &s, cmd

	case DeleteMsg:
		s.loading = true
		return &s, s.doDeleteCmd(msg)
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

	var leftStatus []string
	if s.loading {
		leftStatus = []string{
			s.spinner.View(),
			"loading",
			sort(s.sort, s.sortDirection),
		}
	} else {
		leftStatus = []string{
			"🎬",
			s.paginator.String(),
			sort(s.sort, s.sortDirection),
		}
	}

	rightStatus := sceneFilterStatus(s.sceneFilter, &s.stash)
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

type scenesMessage struct {
	scenes []stash.Scene
	total  int
	err    error
}

// doUpdateCmd sets initial loading state then returns a tea.Cmd to execute loading of scenes.
func (s *ScenesModel) doUpdateCmd() tea.Cmd {
	s.loading = true
	sf := s.sceneFilter
	f := stash.FindFilter{
		Query:     s.query,
		Page:      s.page,
		PerPage:   s.perPage,
		Sort:      s.sort,
		Direction: s.sortDirection,
	}
	return func() tea.Msg {
		var m scenesMessage
		m.scenes, m.total, m.err = s.stash.Scenes(context.Background(), f, sf)
		return m
	}
}

// doDeleteConfirmCmd returns a command to display a confirmation message about the current content.
func (s *ScenesModel) doDeleteConfirmCmd() tea.Cmd {
	return func() tea.Msg {
		s := s.Current()
		titleStyle := lipgloss.NewStyle().
			Foreground(ColorOffWhite)
		return ConfirmationMsg{
			Message:       fmt.Sprintf("Are you sure you want to delete %s?", titleStyle.Render(sceneTitle(s))),
			ConfirmOption: "Delete",
			CancelOption:  "Cancel",
			Cmd: func() tea.Msg {
				return DeleteMsg{Scene: s}
			},
		}
	}
}

type DeleteMsg struct {
	Scene stash.Scene
}

// doDeleteCmd takes a DeleteMessage and attempts to delete the provided scene.  After successful deletion the current
// scenes data is refreshed.
func (s *ScenesModel) doDeleteCmd(d DeleteMsg) tea.Cmd {
	s.loading = true
	return func() tea.Msg {
		_, err := s.stash.DeleteScene(context.Background(), d.Scene.ID)
		if err != nil {
			return ErrorMsg{err}
		}
		return s.doUpdateCmd()()
	}
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
	return fmt.Sprintf("%d⭐", stars)
}
