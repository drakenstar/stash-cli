package app

import (
	"context"
	"path"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type scenesState struct {
	*App

	*paginator[stash.Scene]

	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter
}

func (s *scenesState) Init(ctx context.Context) error {
	s.paginator = NewPaginator[stash.Scene](40)

	s.query = ""
	s.sort = stash.SortDate
	s.sortDirection = stash.SortDirectionDesc

	s.sceneFilter = stash.SceneFilter{}
	s.Reset()
	return s.update(ctx)
}

func (s *scenesState) Update(ctx context.Context, in Input) error {
	switch in.Command() {
	case "":
		if s.Next() {
			if err := s.update(ctx); err != nil {
				return err
			}
		}
		if err := s.Opener(s.Current()); err != nil {
			return err
		}

	case "open", "o":
		if err := s.Opener(s.Current()); err != nil {
			return err
		}

	case "filter", "f":
		s.query = in.ArgString()
		s.Reset()
		if err := s.update(ctx); err != nil {
			return err
		}

	case "random", "r":
		s.sort = stash.RandomSort()
		s.Reset()
		if err := s.update(ctx); err != nil {
			return err
		}

	case "reset":
		if err := s.Init(ctx); err != nil {
			return err
		}

	case "refresh":
		if err := s.update(ctx); err != nil {
			return err
		}

	case "organised", "organized":
		organised := true
		s.sceneFilter.Organized = &organised
		if err := s.update(ctx); err != nil {
			return err
		}

	case "more":
		var ids []string
		for _, p := range s.Current().Performers {
			ids = append(ids, p.ID)
		}
		s.sceneFilter.Performers = &stash.MultiCriterion{
			Value:    ids,
			Modifier: stash.CriterionModifierIncludes,
		}
		if err := s.update(ctx); err != nil {
			return err
		}

	case "stash":
		s.Opener(path.Join("scenes", s.Current().ID))
	}
	return nil
}

func (s *scenesState) View() string {
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
	screenWidth := s.Out.ScreenWidth()
	return lipgloss.JoinVertical(0,
		sceneTable.Render(screenWidth, rows),
		s.renderStatusBar(screenWidth),
	)
}

func (s *scenesState) update(ctx context.Context) (err error) {
	f := stash.FindFilter{
		Query:     s.query,
		Page:      s.page,
		PerPage:   s.perPage,
		Sort:      s.sort,
		Direction: s.sortDirection,
	}
	s.items, s.total, err = s.Scenes(ctx, f, s.sceneFilter)
	return err
}

func (s *scenesState) renderStatusBar(width int) string {
	return ui.StatusRow(width, []string{
		"scenes",
	})
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
