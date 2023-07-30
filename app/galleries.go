package app

import (
	"context"
	"path"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type GalleriesModel struct {
	Stash  stash.Stash
	Opener config.Opener

	paginator[stash.Gallery]
	loading bool
	spinner spinner.Model

	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter

	screen Size
}

func NewGalleriesModel(stash stash.Stash, opener config.Opener) *GalleriesModel {
	s := &GalleriesModel{
		Stash:  stash,
		Opener: opener,
	}

	s.spinner = spinner.New(spinner.WithSpinner(spinner.Globe))

	return s
}

func (s *GalleriesModel) Init(size Size) tea.Cmd {
	s.paginator = NewPaginator[stash.Gallery](40)

	s.query = ""
	s.sort = stash.SortPath
	s.sortDirection = stash.SortDirectionAsc

	s.galleryFilter = stash.GalleryFilter{}
	s.Reset()

	s.screen = size

	return tea.Batch(
		s.doUpdateCmd(),
		spinner.Tick,
	)
}

func (s GalleriesModel) Update(msg tea.Msg) (AppModel, tea.Cmd) {
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
			s.galleryFilter.Organized = &organised
			return &s, s.doUpdateCmd()

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
			return &s, s.doUpdateCmd()

		case "stash":
			s.Opener(path.Join("galleries", s.Current().ID))
		}

	case galleriesMessage:
		s.loading = false
		if msg.err != nil {
			// TODO handle update error somehow
		}
		s.items, s.total = msg.galleries, msg.total

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return &s, cmd
	}

	return &s, nil
}

func (s GalleriesModel) View() string {
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

	var leftStatus []string
	if s.loading {
		leftStatus = []string{
			s.spinner.View(),
			"loading",
			sort(s.sort, s.sortDirection),
		}
	} else {
		leftStatus = []string{
			"📷",
			s.paginator.String(),
			sort(s.sort, s.sortDirection),
		}
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
		statusBar.Render(s.screen.Width, leftStatus, rightStatus),
		galleriesTable.Render(s.screen.Width, rows),
	)
}

type galleriesMessage struct {
	galleries []stash.Gallery
	total     int
	err       error
}

func (s *GalleriesModel) doUpdateCmd() tea.Cmd {
	s.loading = true
	return func() tea.Msg {
		f := stash.FindFilter{
			Query:     s.query,
			Page:      s.page,
			PerPage:   s.perPage,
			Sort:      s.sort,
			Direction: s.sortDirection,
		}
		var g galleriesMessage
		g.galleries, g.total, g.err = s.Stash.Galleries(context.Background(), f, s.galleryFilter)
		return g
	}
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
