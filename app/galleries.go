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

type galleryFilterState struct {
	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter

	pageState paginator
}

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

	history []galleryFilterState

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

func (m *GalleriesModel) PushState(mutate func(*GalleriesModel)) (*GalleriesModel, tea.Cmd) {
	m.history = append(m.history, galleryFilterState{
		query:         m.query,
		sort:          m.sort,
		sortDirection: m.sortDirection,
		galleryFilter: m.galleryFilter,
		pageState:     m.paginator,
	})
	mutate(m)
	m.paginator.Reset()
	return m, m.updateCmd()
}

// Pop sets the current state to the previous state from the history stack.  If the history stack is empty this is a
// noop.
func (m *GalleriesModel) Pop() (*GalleriesModel, tea.Cmd) {
	if len(m.history) == 0 {
		return m, nil
	}

	state := m.history[len(m.history)-1]
	m.history = m.history[0 : len(m.history)-1]

	// Restore previous state, including pagination
	m.paginator = state.pageState
	m.query = state.query
	m.sort = state.sort
	m.sortDirection = state.sortDirection
	m.galleryFilter = state.galleryFilter
	m.galleries = []stash.Gallery{}

	return m, m.updateCmd()
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
		case "z":
			if s.Skip(-1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case "x":
			if s.Skip(1) {
				s.Clear()
				return &s, s.updateCmd()
			}
		case "/":
			return &s, NewModeCommandCmd("/", "filter ")
		case "o":
			msg := OpenMsg{s.Current()}
			return &s, func() tea.Msg { return msg }
		case "r":
			return s.PushState(func(gm *GalleriesModel) {
				gm.sort = stash.RandomSort()
			})
		case "u": // "Undo"
			return s.Pop()
		case "f":
			return s.PushState(func(gm *GalleriesModel) {
				if gm.galleryFilter.PerformerFavourite == nil {
					favourite := true
					gm.galleryFilter.PerformerFavourite = &favourite
				} else {
					gm.galleryFilter.PerformerFavourite = nil
				}
			})
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
			return s.PushState(func(gm *GalleriesModel) {
				gm.query = msg.ArgString()
			})

		case "random":
			return s.PushState(func(gm *GalleriesModel) {
				gm.sort = stash.RandomSort()
			})

		case "today":
			return s.PushState(func(gm *GalleriesModel) {
				gm.galleryFilter.CreatedAt = &stash.TimestampCriterion{
					Value:    time.Now().Add(-24 * time.Hour),
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "count":
			val, err := msg.ArgInt()
			if err != nil {
				return &s, NewErrorCmd(err)
			}
			return s.PushState(func(gm *GalleriesModel) {
				s.galleryFilter.ImageCount = &stash.IntCriterion{
					Value:    val,
					Modifier: stash.CriterionModifierGreaterThan,
				}
			})

		case "reset":
			return &s, s.reset()

		case "refresh":
			return &s, s.updateCmd()

		case "organised", "organized":
			organised := true
			if msg.ArgString() == "false" {
				organised = false
			}
			return s.PushState(func(gm *GalleriesModel) {
				gm.galleryFilter.Organized = &organised
			})

		case "favourite", "favorite":
			favourite := true
			if msg.ArgString() == "false" {
				favourite = false
			}
			return s.PushState(func(gm *GalleriesModel) {
				gm.galleryFilter.PerformerFavourite = &favourite
			})

		case "more":
			return s.newTabPerformerCmd()

		case "rated":
			return s.PushState(func(gm *GalleriesModel) {
				gm.galleryFilter.Rating100 = &stash.IntCriterion{
					Modifier: stash.CriterionModifierNotNull,
				}
			})

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
		s.paginator.String(),
		sort(s.sort, s.sortDirection),
	}

	rightStatus := galleryFilterStatus(s.galleryFilter, s.StashLookup)
	if s.query != "" {
		rightStatus = append(rightStatus, "\""+s.query+"\"")
	}
	if len(s.history) > 0 {
		rightStatus = append(rightStatus, fmt.Sprintf("[%d]", len(s.history)))
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
