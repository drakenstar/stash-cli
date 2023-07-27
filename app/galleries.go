package app

import (
	"context"
	"path"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type GalleriesState struct {
	Stash  stash.Stash
	Opener Opener

	*paginator[stash.Gallery]

	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter
}

func (s *GalleriesState) Init(ctx context.Context) error {
	s.paginator = NewPaginator[stash.Gallery](40)

	s.query = ""
	s.sort = stash.SortPath
	s.sortDirection = stash.SortDirectionAsc

	s.galleryFilter = stash.GalleryFilter{}
	s.Reset()
	return s.update(ctx)
}

func (s *GalleriesState) Update(ctx context.Context, in Input) error {
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
		s.galleryFilter.Organized = &organised
		if err := s.update(ctx); err != nil {
			return err
		}

	case "more":
		var ids []string
		for _, p := range s.Current().Performers {
			ids = append(ids, p.ID)
		}
		s.galleryFilter.Performers = &stash.MultiCriterion{
			Value:    ids,
			Modifier: stash.CriterionModifierIncludes,
		}
		s.Reset()
		if err := s.update(ctx); err != nil {
			return err
		}

	case "stash":
		s.Opener(path.Join("galleries", s.Current().ID))
	}
	return nil
}

func (s *GalleriesState) View(width int) string {
	var rows []ui.Row
	for i, gallery := range s.items {
		rows = append(rows, ui.Row{
			Values: []string{
				organised(gallery.Organized),
				gallery.Date,
				galleryTitle(gallery),
				gallerySize(gallery),
				gallery.Studio.Name,
				performerList(gallery.Performers),
				tagList(gallery.Tags),
				details(gallery.Details),
			}})
		if s.index == i {
			rows[i].Background = &ColorRowSelected
		}
	}

	leftStatus := []string{
		"📷",
		s.paginator.String(),
		sort(s.sort, s.sortDirection),
	}
	rightStatus := []string{}
	if s.query != "" {
		rightStatus = append(rightStatus, "\""+s.query+"\"")
	}
	if s.galleryFilter.Organized != nil {
		if *s.galleryFilter.Organized {
			rightStatus = append(rightStatus, "organized")
		} else {
			rightStatus = append(rightStatus, "not organized")
		}
	}

	return lipgloss.JoinVertical(0,
		galleriesTable.Render(width, rows),
		statusBar.Render(width, leftStatus, rightStatus),
	)
}

func (s *GalleriesState) update(ctx context.Context) (err error) {
	f := stash.FindFilter{
		Query:     s.query,
		Page:      s.page,
		PerPage:   s.perPage,
		Sort:      s.sort,
		Direction: s.sortDirection,
	}
	s.items, s.total, err = s.Stash.Galleries(ctx, f, s.galleryFilter)
	return err
}

var (
	galleriesTable = &ui.Table{
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
