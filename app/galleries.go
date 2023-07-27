package app

import (
	"context"
	"path"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type galleriesState struct {
	*App

	opened bool
	*paginator

	galleries []stash.Gallery

	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter
}

func (s *galleriesState) Init(ctx context.Context) error {
	s.paginator = &paginator{
		Index:   0,
		Total:   0,
		Page:    1,
		PerPage: 40,
	}

	s.query = ""
	s.sort = stash.SortPath
	s.sortDirection = stash.SortDirectionAsc

	s.galleryFilter = stash.GalleryFilter{}
	s.resetPagination()
	return s.update(ctx)
}

func (s *galleriesState) Update(ctx context.Context, in Input) error {
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
		if err := s.Opener(s.galleries[s.Index]); err != nil {
			return err
		}

	case "open", "o":
		if err := s.Opener(s.galleries[s.Index]); err != nil {
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
		s.galleryFilter.Organized = &organised
		if err := s.update(ctx); err != nil {
			return err
		}

	case "more":
		var ids []string
		for _, p := range s.galleries[s.Index].Performers {
			ids = append(ids, p.ID)
		}
		s.galleryFilter.Performers = &stash.MultiCriterion{
			Value:    ids,
			Modifier: stash.CriterionModifierIncludes,
		}
		if err := s.update(ctx); err != nil {
			return err
		}

	case "stash":
		s.Opener(path.Join("galleries", s.galleries[s.Index].ID))
	}
	return nil
}

func (s *galleriesState) View() string {
	var rows []ui.Row
	for _, s := range s.galleries {
		gallery := galleryPresenter{s}
		rows = append(rows, []string{
			gallery.organised(),
			gallery.Date,
			gallery.title(),
			gallery.size(),
			gallery.Studio.Name,
			gallery.performerList(),
			gallery.tagList(),
			gallery.details(),
		})
	}
	return galleryTable.Render(s.Out.ScreenWidth(), rows)
}

func (s *galleriesState) resetPagination() {
	s.Index = 0
	s.Page = 1
	s.opened = false
}

func (s *galleriesState) update(ctx context.Context) (err error) {
	f := stash.FindFilter{
		Query:     s.query,
		Page:      s.Page,
		PerPage:   s.PerPage,
		Sort:      s.sort,
		Direction: s.sortDirection,
	}
	s.galleries, s.Total, err = s.Galleries(ctx, f, s.galleryFilter)
	return err
}

var (
	galleryTable = &ui.Table{
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
