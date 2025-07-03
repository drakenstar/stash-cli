package app

import (
	"context"
	"fmt"
	"math"
	"path"
	"time"

	"github.com/brunoga/deep"
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
}

type ScenesModel struct {
	stash  stashCache
	Opener config.Opener

	paginator[stash.Scene]
	loading bool
	spinner spinner.Model

	current *filterState
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
	s.paginator = NewPaginator[stash.Scene](40)

	s.current = nil
	s.history = []filterState{}
	s.Push(&filterState{
		query: "",
		sort: stash.SortDate,
		sortDirection: stash.SortDirectionDesc,
		sceneFilter: stash.SceneFilter{},
	})
	
	return tea.Batch(
		s.doUpdateCmd(),
		s.spinner.Tick,
	)
}

func (s *ScenesModel) Clone() *filterState {
	f := deep.MustCopy(s.current)
	return f
}

// Push sets the new current state and pushes the existing state onto the history stack.  Pagination is reset.
func (s *ScenesModel) Push(f *filterState) {
	if (s.current != nil) {
		s.history = append(s.history, *s.current)
	}
	s.current = f
	s.Reset()
}

// Pop sets the current state to the previous state from the history stack.  If the history stack is empty this is a
// noop.
func (s *ScenesModel) Pop() {
	if (len(s.history) == 0) {
		return
	}
	s.current = &s.history[len(s.history) - 1]
	s.history = s.history[0:len(s.history) - 1]
	s.Reset()
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

	case Input:
		switch msg.Command() {
		case "":
			if s.Next() {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
			s.Opener(s.Current())

		case "open", "o":
			s.Opener(s.Current())

		case "filter", "f":
			f := s.Clone()
			f.query = msg.ArgString()
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "sort":
			sort := msg.ArgString()
			switch sort {
			case "date":
				f := s.Clone()
				f.sort = sort
				f.sortDirection = stash.SortDirectionAsc
				s.Push(f)
				return &s, s.doUpdateCmd()

			case "-date":
				f := s.Clone()
				f.sort = sort[1:]
				f.sortDirection = stash.SortDirectionDesc
				s.Push(f)
				return &s, s.doUpdateCmd()
			}
		
		case "studio":
			f := s.Clone()
			f.sceneFilter.Studios = &stash.HierarchicalMultiCriterion{
				Value:    msg.Args(),
				Modifier: stash.CriterionModifierIncludes,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "random", "r":
			f := s.Clone()
			f.sort = stash.RandomSort()
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "recent":
			f := s.Clone()
			f.sceneFilter.CreatedAt = &stash.TimestampCriterion{
				Value:    time.Now().Add(-24 * time.Hour * 7),
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "year":
			f := s.Clone()
			f.sceneFilter.Date = &stash.DateCriterion{
				Value:    time.Now().Add(-24 * time.Hour * 365),
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()
		
		case "before":
			val, err := msg.ArgInt()
			if err != nil {
				return &s, NewErrorCmd(err)
			}
			t := time.Date(val, time.January, 1, 0, 0, 0, 0, time.UTC)
			f := s.Clone()
			f.sceneFilter.Date = &stash.DateCriterion{
				Value:    t,
				Modifier: stash.CriterionModifierLessThan,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "6mo":
			f := s.Clone()
			f.sceneFilter.CreatedAt = &stash.TimestampCriterion{
				Value:    time.Now().Add(-24 * time.Hour * 182),
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "1mo":
			f := s.Clone()
			f.sceneFilter.CreatedAt = &stash.TimestampCriterion{
				Value:    time.Now().Add(-24 * time.Hour * 30),
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "pop", "p":
			s.Pop()
			return &s, s.doUpdateCmd()

		case "duration":
			val, err := msg.ArgInt()
			if err != nil {
				return &s, NewErrorCmd(err)
			}
			f := s.Clone()
			f.sceneFilter.Duration = &stash.IntCriterion{
				Value:    val,
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "reset":
			return &s, s.Init(s.screen)

		case "refresh":
			return &s, s.doUpdateCmd()

		case "organised", "organized":
			organised := true
			if msg.ArgString() == "false" {
				organised = false
			}
			f := s.Clone()
			f.sceneFilter.Organized = &organised
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "tags":
			f := s.Clone()
			f.sceneFilter.Tags = &stash.HierarchicalMultiCriterion{
				Value:    []string{msg.ArgString()},
				Modifier: stash.CriterionModifierIncludes,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "pt":
			f := s.Clone()
			f.sceneFilter.PerformerTags = &stash.HierarchicalMultiCriterion{
				Value:    []string{msg.ArgString()},
				Modifier: stash.CriterionModifierIncludes,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "favourite", "favorite":
			favourite := true
			if msg.ArgString() == "false" {
				favourite = false
			}
			f := s.Clone()
			f.sceneFilter.PerformerFavourite = &favourite
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "more":
			var ids []string
			for _, p := range s.Current().Performers {
				ids = append(ids, p.ID)
			}
			f := s.Clone()
			f.sceneFilter.Performers = &stash.MultiCriterion{
				Value:    ids,
				Modifier: stash.CriterionModifierIncludes,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

		case "rated":
			f := s.Clone()
			f.sceneFilter.Rating100 = &stash.IntCriterion{
				Modifier: stash.CriterionModifierNotNull,
			}
			s.Push(f)
			return &s, s.doUpdateCmd()

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
		s.items, s.total = msg.scenes, msg.total

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
	for i, scene := range s.items {
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
			sort(s.current.sort, s.current.sortDirection),
		}
	} else {
		leftStatus = []string{
			"🎬",
			s.paginator.String(),
			sort(s.current.sort, s.current.sortDirection),
		}
	}
	rightStatus := []string{fmt.Sprintf("[%d]", len(s.history))}
	rightStatus = append(rightStatus, sceneFilterStatus(s.current.sceneFilter, &s.stash)...)
	if s.current.query != "" {
		rightStatus = append(rightStatus, "\""+s.current.query+"\"")
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
	return func() tea.Msg {
		f := stash.FindFilter{
			Query:     s.current.query,
			Page:      s.page,
			PerPage:   s.perPage,
			Sort:      s.current.sort,
			Direction: s.current.sortDirection,
		}
		var m scenesMessage
		m.scenes, m.total, m.err = s.stash.Scenes(context.Background(), f, s.current.sceneFilter)
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
