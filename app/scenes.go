package app

import (
	"context"
	"fmt"
	"path"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type ScenesModel struct {
	Stash  stash.Stash
	Opener config.Opener

	paginator[stash.Scene]
	loading bool
	spinner spinner.Model

	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter

	screen Size
}

func NewScenesModel(stash stash.Stash, opener config.Opener) *ScenesModel {
	s := &ScenesModel{
		Stash:  stash,
		Opener: opener,
	}

	s.spinner = spinner.New(spinner.WithSpinner(spinner.Globe))

	return s
}

func (s *ScenesModel) Init(size Size) tea.Cmd {
	s.paginator = NewPaginator[stash.Scene](40)

	s.query = ""
	s.sort = stash.SortDate
	s.sortDirection = stash.SortDirectionDesc

	s.sceneFilter = stash.SceneFilter{}
	s.Reset()
	return tea.Batch(
		s.doUpdateCmd(),
		spinner.Tick,
	)
}

func (s ScenesModel) Update(msg tea.Msg) (AppModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if !s.loading && s.Skip(-1) {
				return &s, s.doUpdateCmd()
			}
		case tea.KeyDown:
			if !s.loading && s.Skip(1) {
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
				return &s, s.doUpdateCmd()
			}
			s.Opener(s.Current())

		case "open", "o":
			s.Opener(s.Current())

		case "filter", "f":
			s.query = msg.ArgString()
			s.Reset()
			return &s, s.doUpdateCmd()

		case "random", "r":
			s.sort = stash.RandomSort()
			s.Reset()
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
			s.sceneFilter.Organized = &organised
			return &s, s.doUpdateCmd()

		case "more":
			var ids []string
			for _, p := range s.Current().Performers {
				ids = append(ids, p.ID)
			}
			s.sceneFilter.Performers = &stash.MultiCriterion{
				Value:    ids,
				Modifier: stash.CriterionModifierIncludes,
			}
			s.Reset()
			return &s, s.doUpdateCmd()

		case "stash":
			s.Opener(path.Join("scenes", s.Current().ID))

		case "delete":
			return &s, s.doDeleteConfirmCmd()
		}

	case scenesMessage:
		s.loading = false
		if msg.err != nil {
			// TODO handle update error somehow
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
	rightStatus := []string{}
	if s.query != "" {
		rightStatus = append(rightStatus, "\""+s.query+"\"")
	}
	if s.sceneFilter.Organized != nil {
		if *s.sceneFilter.Organized {
			rightStatus = append(rightStatus, "organized")
		} else {
			rightStatus = append(rightStatus, "not organized")
		}
	}
	if s.sceneFilter.Performers != nil {
		rightStatus = append(rightStatus, "performers") // TODO actually output performer names
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
			Query:     s.query,
			Page:      s.page,
			PerPage:   s.perPage,
			Sort:      s.sort,
			Direction: s.sortDirection,
		}
		var m scenesMessage
		m.scenes, m.total, m.err = s.Stash.Scenes(context.Background(), f, s.sceneFilter)
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
		_, err := s.Stash.DeleteScene(context.Background(), d.Scene.ID)
		if err != nil {
			panic(err)
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
