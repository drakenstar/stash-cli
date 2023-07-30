package app

import (
	"context"
	"path"

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

	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter

	screen Size
}

func (s *ScenesModel) Init(size Size) tea.Cmd {
	s.paginator = NewPaginator[stash.Scene](40)

	s.query = ""
	s.sort = stash.SortDate
	s.sortDirection = stash.SortDirectionDesc

	s.sceneFilter = stash.SceneFilter{}
	s.Reset()
	return s.update
}

func (s ScenesModel) Update(msg tea.Msg) (AppModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
	case Input:
		switch msg.Command() {
		case "":
			if s.Next() {
				return &s, s.update
			}
			s.Opener(s.Current())

		case "open", "o":
			s.Opener(s.Current())

		case "filter", "f":
			s.query = msg.ArgString()
			s.Reset()
			return &s, s.update

		case "random", "r":
			s.sort = stash.RandomSort()
			s.Reset()
			return &s, s.update

		case "reset":
			return &s, s.Init(s.screen)

		case "refresh":
			return &s, s.update

		case "organised", "organized":
			organised := true
			s.sceneFilter.Organized = &organised
			return &s, s.update

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
			return &s, s.update

		case "stash":
			s.Opener(path.Join("scenes", s.Current().ID))
		}

	case scenesMessage:
		if msg.err != nil {
			// TODO handle update error somehow
		}
		s.items, s.total = msg.scenes, msg.total
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

	leftStatus := []string{
		"🎬",
		s.paginator.String(),
		sort(s.sort, s.sortDirection),
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
		sceneTable.Render(s.screen.Width, rows),
		statusBar.Render(s.screen.Width, leftStatus, rightStatus),
	)
}

type scenesMessage struct {
	scenes []stash.Scene
	total  int
	err    error
}

func (s *ScenesModel) update() tea.Msg {
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
