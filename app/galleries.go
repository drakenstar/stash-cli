package app

import (
	"fmt"
	"path"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type GalleryService interface {
	Galleries(stash.FindFilter, stash.GalleryFilter) tea.Cmd
}

type GalleriesModel struct {
	GalleryService
	StashLookup

	paginator
	galleries []stash.Gallery

	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter

	screen Size
}

func NewGalleriesModel(galleryService GalleryService, lookup StashLookup) *GalleriesModel {
	m := &GalleriesModel{
		GalleryService: galleryService,
		StashLookup:    lookup,
	}
	m.reset()
	return m
}

func (m *GalleriesModel) Current() stash.Gallery {
	return m.galleries[m.paginator.index]
}

func (m *GalleriesModel) reset() tea.Cmd {
	m.paginator = NewPaginator(40)

	m.query = ""
	m.sort = stash.SortPath
	m.sortDirection = stash.SortDirectionAsc
	m.galleryFilter = stash.GalleryFilter{}

	return m.updateCmd()
}

func (s *GalleriesModel) Init(size Size) tea.Cmd {
	s.screen = size
	return s.updateCmd()
}

func (s *GalleriesModel) Title() string {
	t := "Galleries"
	if s.query != "" {
		t = fmt.Sprintf("\"%s\"", s.query)
	} else if s.galleryFilter.Performers != nil {
		var performers []string
		for _, p := range s.galleryFilter.Performers.Value {
			perf, _ := s.StashLookup.GetPerformer(p)
			performers = append(performers, perf.Name)
		}
		t = strings.Join(performers, ", ")
	}

	return fmt.Sprintf("%c %s (%s)", '\U000f0fce', t, humanNumber(s.total))
}

func (s GalleriesModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			if s.Skip(-1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case tea.KeyDown:
			if s.Skip(1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case tea.KeyEnter, tea.KeySpace:
			if s.Next() {
				s.Clear()
				return &s, s.updateCmd()
			}
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }
		}

		switch msg.String() {
		case "o":
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }
		case "r":
			s.sort = stash.RandomSort()
			s.Reset()
			return &s, s.updateCmd()
		case "/":
			return &s, NewModeCommandCmd("/", "filter ")
		case "p":
			return s.newTabPerformerCmd()
		case "`":
			msg := OpenMsg{path.Join("galleries", s.Current().ID)}
			return &s, func() tea.Msg { return msg }
		}

	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

	case ui.CommandExecuteMsg:
		switch msg.Name() {
		case "open":
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }

		case "filter":
			s.query = msg.ArgString()
			s.Reset()
			return &s, s.updateCmd()

		case "random":
			s.sort = stash.RandomSort()
			s.Reset()
			return &s, s.updateCmd()

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
				return &s, s.updateCmd()
			}

		case "today":
			s.galleryFilter.CreatedAt = &stash.TimestampCriterion{
				Value:    time.Now().Add(-24 * time.Hour),
				Modifier: stash.CriterionModifierGreaterThan,
			}
			s.Reset()
			return &s, s.updateCmd()

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
			return &s, s.updateCmd()

		case "reset":
			return &s, s.reset()

		case "refresh":
			return &s, s.updateCmd()

		case "organised", "organized":
			organised := true
			s.galleryFilter.Organized = &organised
			return &s, s.updateCmd()

		case "favourite", "favorite":
			favourite := true
			if msg.ArgString() == "false" {
				favourite = false
			}
			s.galleryFilter.PerformerFavourite = &favourite
			return &s, s.updateCmd()

		case "more":
			return s.newTabPerformerCmd()

		case "stash":
			msg := OpenMsg{path.Join("galleries", s.Current().ID)}
			return &s, func() tea.Msg { return msg }
		}

	case galleriesMsg:
		s.galleries, s.total = msg.galleries, msg.total
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

func (m *GalleriesModel) updateCmd() tea.Cmd {
	return m.GalleryService.Galleries(stash.FindFilter{
		Query:     m.query,
		Page:      m.page,
		PerPage:   m.perPage,
		Sort:      m.sort,
		Direction: m.sortDirection,
	}, m.galleryFilter)
}

// newTabPerformerCmd returns a command that opens a new tab filtered to the current set of performers
func (m *GalleriesModel) newTabPerformerCmd() (*GalleriesModel, tea.Cmd) {
	if len(m.Current().Performers) == 0 {
		return m, nil
	}

	var ids []string
	for _, p := range m.Current().Performers {
		ids = append(ids, p.ID)
	}
	return m, func() tea.Msg {
		return TabOpenMsg{
			tabFunc: func() TabModel {
				t := NewGalleriesModel(m.GalleryService, m.StashLookup)
				t.galleryFilter.Performers = &stash.MultiCriterion{
					Value:    ids,
					Modifier: stash.CriterionModifierIncludes,
				}
				return t
			},
		}
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
