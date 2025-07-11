package app

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type GalleryService interface {
	Galleries(context.Context, stash.FindFilter, stash.GalleryFilter) ([]stash.Gallery, int, error)
}

type GalleriesModel struct {
	GalleryService
	StashLookup
	Opener config.Opener

	paginator
	spinner   spinner.Model
	galleries []stash.Gallery

	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter

	screen Size
}

func NewGalleriesModel(galleryService GalleryService, lookup StashLookup, opener config.Opener) *GalleriesModel {
	s := &GalleriesModel{
		GalleryService: galleryService,
		StashLookup:    lookup,
		Opener:         opener,
	}

	s.spinner = spinner.New(spinner.WithSpinner(spinner.Globe))

	return s
}

func (m *GalleriesModel) Current() stash.Gallery {
	return m.galleries[m.paginator.index]
}

func (s *GalleriesModel) Init(size Size) tea.Cmd {
	s.paginator = NewPaginator(40)

	s.query = ""
	s.sort = stash.SortPath
	s.sortDirection = stash.SortDirectionAsc

	s.galleryFilter = stash.GalleryFilter{}
	s.Reset()

	s.screen = size

	return tea.Batch(
		s.doUpdateCmd(),
		s.spinner.Tick,
	)
}

func (s *GalleriesModel) Title() string {
	return "Galleries"
}

func (s GalleriesModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if s.Skip(-1) {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
		case tea.KeyDown:
			if s.Skip(1) {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
		case tea.KeyEnter, tea.KeySpace:
			if s.Next() {
				s.Clear()
				return &s, s.doUpdateCmd()
			}
			s.Opener(s.Current())
			return &s, nil
		}

		switch msg.String() {
		case "o":
			s.Opener(s.Current())
			return &s, nil
		case "r":
			s.sort = stash.RandomSort()
			s.Reset()
			return &s, s.doUpdateCmd()
		case "/":
			return &s, NewModeCommandCmd("/", "filter ")
		}

	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

	case ui.CommandExecuteMsg:
		switch msg.Name() {
		case "open":
			s.Opener(s.Current())
			return &s, nil

		case "filter":
			s.query = msg.ArgString()
			s.Reset()
			return &s, s.doUpdateCmd()

		case "random":
			s.sort = stash.RandomSort()
			s.Reset()
			return &s, s.doUpdateCmd()

		case "sort":
			args := msg.Args()
			if len(args) > 0 {
				direction := stash.SortDirectionAsc
				if len(args) >= 2 {
					switch strings.ToUpper(args[1]) {
					case stash.SortDirectionAsc:
					case stash.SortDirectionDesc:
						direction = stash.SortDirectionDesc
					default:
						return &s, NewErrorCmd(fmt.Errorf("invalid sort direction '%s'", args[1]))
					}
				}
				s.sortDirection = direction

				switch strings.ToLower(args[0]) {
				case "random":
					s.sort = stash.RandomSort()
				case stash.SortDate:
					s.sort = stash.SortDate
				case stash.SortUpdatedAt:
					s.sort = stash.SortUpdatedAt
				case stash.SortCreatedAt:
					s.sort = stash.SortCreatedAt
				case stash.SortPath:
					s.sort = stash.SortPath
				default:
					return &s, NewErrorCmd(fmt.Errorf("invalid sort type '%s'", args[1]))
				}

				s.Reset()
				return &s, s.doUpdateCmd()
			}

		case "today":
			s.galleryFilter.CreatedAt = &stash.TimestampCriterion{
				Value:    time.Now().Add(-24 * time.Hour),
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Reset()
			return &s, s.doUpdateCmd()

		case "count":
			val, err := msg.ArgInt()
			if err != nil {
				return &s, NewErrorCmd(err)
			}
			s.galleryFilter.ImageCount = &stash.IntCriterion{
				Value:    val,
				Modifier: stash.CriterionModifierGreaterThan,
			}
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

		case "favourite", "favorite":
			favourite := true
			if msg.ArgString() == "false" {
				favourite = false
			}
			s.galleryFilter.PerformerFavourite = &favourite
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
		if msg.err != nil {
			return &s, NewErrorCmd(msg.err)
		}
		s.galleries, s.total = msg.galleries, msg.total

	case spinner.TickMsg:
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return &s, cmd
	}

	return &s, nil
}

func (s GalleriesModel) View() string {
	var rows []ui.Row
	for i, gallery := range s.galleries {
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

	rightStatus := galleryFilterStatus(s.galleryFilter, s.StashLookup)
	if s.query != "" {
		rightStatus = append(rightStatus, "\""+s.query+"\"")
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
	return func() tea.Msg {
		f := stash.FindFilter{
			Query:     s.query,
			Page:      s.page,
			PerPage:   s.perPage,
			Sort:      s.sort,
			Direction: s.sortDirection,
		}
		var g galleriesMessage
		g.galleries, g.total, g.err = s.Galleries(context.Background(), f, s.galleryFilter)
		return g
	}
}

var (
	galleriesTable = &ui.Table{
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
				Name:       "Title",
				Foreground: &ColorOffWhite,
				Bold:       true,
				Weight:     2,
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
