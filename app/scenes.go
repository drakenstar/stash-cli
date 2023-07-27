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

	opened bool
	*paginator

	scenes []stash.Scene

	query         string
	sort          string
	sortDirection string
	sceneFilter   stash.SceneFilter
}

func (s *scenesState) Init(ctx context.Context) error {
	s.paginator = &paginator{
		Index:   0,
		Total:   0,
		Page:    1,
		PerPage: 40,
	}

	s.query = ""
	s.sort = stash.SortDate
	s.sortDirection = stash.SortDirectionDesc

	s.sceneFilter = stash.SceneFilter{}
	s.resetPagination()
	return s.update(ctx)
}

func (s *scenesState) Update(ctx context.Context, in Input) error {
	switch in.Command() {
	case "":
		if s.opened {
			if s.Skip(1) {
				if err := s.update(ctx); err != nil {
					return err
				}
			}
		} else {
			s.opened = true
		}
		if err := s.Opener(s.scenes[s.Index]); err != nil {
			return err
		}

	case "open", "o":
		if err := s.Opener(s.scenes[s.Index]); err != nil {
			return err
		}

	case "filter", "f":
		s.query = in.ArgString()
		s.resetPagination()
		if err := s.update(ctx); err != nil {
			return err
		}

	case "random", "r":
		s.sort = stash.RandomSort()
		s.resetPagination()
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
		for _, p := range s.scenes[s.Index].Performers {
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
		s.Opener(path.Join("scenes", s.scenes[s.Index].ID))
	}
	return nil
}

func (s *scenesState) View() string {
	var rows []ui.Row
	for i, scene := range s.scenes {
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
		if s.Index == i {
			rows[i].Background = &ColorRowSelected
		}
	}
	return sceneTable.Render(s.Out.ScreenWidth(), rows)
}

func (s *scenesState) resetPagination() {
	s.Index = 0
	s.Page = 1
	s.opened = false
}

func (s *scenesState) update(ctx context.Context) (err error) {
	f := stash.FindFilter{
		Query:     s.query,
		Page:      s.Page,
		PerPage:   s.PerPage,
		Sort:      s.sort,
		Direction: s.sortDirection,
	}
	s.scenes, s.Total, err = s.Scenes(ctx, f, s.sceneFilter)
	return err
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
