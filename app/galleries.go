package app

import (
	"fmt"
	"path"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/command"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type galleryFilterState struct {
	query         string
	sort          string
	sortDirection string
	galleryFilter stash.GalleryFilter

	pageState pageState
}

type GalleryService interface {
	Galleries(stash.FindFilter, stash.GalleryFilter) tea.Cmd
}

type GalleriesModel struct {
	GalleryService
	StashLookup

	pageState pageState
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
	m.pageState.PerPage = 40
	m.reset()
	return m
}

func (m *GalleriesModel) Current() stash.Gallery {
	return m.galleries[m.pageState.index]
}

func (m *GalleriesModel) reset() tea.Cmd {
	m.query = ""
	m.sort = stash.SortPath
	m.sortDirection = stash.SortDirectionAsc
	m.galleryFilter = stash.GalleryFilter{}
	m.pageState.Reset()

	return m.updateCmd()
}

// TODO probably it's the responsiblity of the parent to tell this model exactly how tall it is, so that it's not
// doing it's own math to solve this.
func (m *GalleriesModel) SetHeight(height int) {
	m.pageState.PerPage = 0
	if height >= 5 {
		m.pageState.PerPage = height - 5
	}
}

func (m *GalleriesModel) Init(size Size) tea.Cmd {
	m.screen = size
	m.SetHeight(size.Height)
	return m.updateCmd()
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

	return fmt.Sprintf("%c %s (%s)", '\U000f0fce', t, humanNumber(s.pageState.total))
}

func (m *GalleriesModel) PushState(mutate func(*GalleriesModel)) (*GalleriesModel, tea.Cmd) {
	m.history = append(m.history, galleryFilterState{
		query:         m.query,
		sort:          m.sort,
		sortDirection: m.sortDirection,
		galleryFilter: m.galleryFilter,
		pageState:     m.pageState,
	})
	mutate(m)
	m.pageState.Reset()
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
	m.pageState = state.pageState
	m.query = state.query
	m.sort = state.sort
	m.sortDirection = state.sortDirection
	m.galleryFilter = state.galleryFilter
	m.galleries = []stash.Gallery{}

	return m, m.updateCmd()
}

// Keymap allows mapping a keyboard shortcut to a command.  Commands are interpretted in command mode and do not take
// additional parameters.
var GalleriesModelDefaultKeymap = map[string]string{
	"up":    "skip -1",
	"down":  "skip 1",
	"enter": "open skip",
	" ":     "open skip", // space
	"z":     "skip -1",
	"x":     "skip 1",
	"o":     "open",
	"r":     "sort random",
	"u":     "undo", // state pop?  Maybe some sort of generic state management command
	"f":     "filter favourite=1",
	"p":     "filter performer=current",
	"`":     "open-url stash",
}

// Command aliases can be used to alias useful commands.  This will act as a prefix for a command, meaning that
// additional inputs can be given after the alias.
var GalleriesModelDefaultCommandAlias = map[string]string{
	"recent": "filter createdAt=>-24h",
	"year":   "filter date=>-1y",
}

var GalleriesModelCommandConfig command.Config = command.Config{
	"filter":   binder[GalleriesModelFilterMsg](),
	"open":     binder[GalleriesModelOpenMsg](),
	"open-url": binder[GalleriesModelOpenURLMsg](),
	"reset":    binder[GalleriesModelResetMsg](),
	"sort":     binder[GalleriesModelSortMsg](),
	"skip":     binder[GalleriesModelSkipMsg](),
	"undo":     binder[GalleriesModelUndoMsg](),
}

// GalleriesModelFilterMsg controls the filtering of various fields on the model. Currently this has a bit of a limitation
// in that although pointers can be used to determine if the user intended to set a field or not, there is no way
// currently to reset a filter.
type GalleriesModelFilterMsg struct {
	Query        *string
	Favourite    *bool
	Organised    *bool
	Rating       *int
	Performer    *string
	Count        *int
	PerformerTag *string
	Tag          *string
	Studio       *string
}

type GalleriesModelOpenMsg struct {
	Skip bool `actions:",positional"`
}

type GalleriesModelOpenURLMsg struct {
	Source string
}

type GalleriesModelResetMsg struct{}

type GalleriesModelSkipMsg struct {
	Count int `actions:",positional"`
}

type GalleriesModelSortMsg struct {
	Field     string `actions:",positional"`
	Direction string
}

type GalleriesModelUndoMsg struct{}

func (m GalleriesModel) CommandConfig() command.Config {
	return GalleriesModelCommandConfig
}

func (m GalleriesModel) Search(query string) tea.Msg {
	return GalleriesModelFilterMsg{
		Query: &query,
	}
}

func (s GalleriesModel) Update(msg tea.Msg) (TabModel, tea.Cmd) {
	switch msg := msg.(type) {
	case GalleriesModelFilterMsg:
		return s.PushState(func(gm *GalleriesModel) {
			if msg.Query != nil {
				gm.query = *msg.Query
			}
			if msg.Favourite != nil {
				gm.galleryFilter.PerformerFavourite = msg.Favourite
			}
			if msg.Organised != nil {
				gm.galleryFilter.Organized = msg.Organised
			}
			if msg.Rating != nil {
				gm.galleryFilter.Rating100 = &stash.IntCriterion{
					Modifier: stash.CriterionModifierEquals,
					Value:    *msg.Rating,
				}
			}
			if msg.Performer != nil {
				if *msg.Performer == "current" {
					var ids []string
					for _, p := range gm.Current().Performers {
						ids = append(ids, p.ID)
					}
					gm.galleryFilter.Performers = &stash.MultiCriterion{
						Value:    ids,
						Modifier: stash.CriterionModifierIncludes,
					}
				} else {
					gm.galleryFilter.Performers = &stash.MultiCriterion{
						Value:    []string{*msg.Performer},
						Modifier: stash.CriterionModifierIncludes,
					}
				}
			}
			if msg.Studio != nil {
				gm.galleryFilter.Studios = &stash.HierarchicalMultiCriterion{
					Value:    []string{*msg.Studio},
					Modifier: stash.CriterionModifierIncludes,
				}
			}
			if msg.Tag != nil {
				gm.galleryFilter.Tags = &stash.HierarchicalMultiCriterion{
					Value:    []string{*msg.Tag},
					Modifier: stash.CriterionModifierIncludes,
				}
			}
			if msg.PerformerTag != nil {
				gm.galleryFilter.PerformerTags = &stash.HierarchicalMultiCriterion{
					Value:    []string{*msg.PerformerTag},
					Modifier: stash.CriterionModifierIncludes,
				}
			}
			if msg.Count != nil {
				gm.galleryFilter.FileCount = &stash.IntCriterion{
					Value:    *msg.Count,
					Modifier: stash.CriterionModifierGreaterThan, // TODO modiifiers
				}
			}
		})

	case GalleriesModelOpenMsg:
		if msg.Skip && s.pageState.Next() {
			return &s, s.updateCmd()
		}
		cur := s.Current()
		return &s, func() tea.Msg { return OpenMsg{cur} }

	case GalleriesModelOpenURLMsg:
		cur := s.Current()
		var src string
		switch msg.Source {
		default:
			src = path.Join("scenes", cur.ID)
		}
		return &s, func() tea.Msg { return OpenMsg{src} }

	case GalleriesModelResetMsg:
		return &s, s.reset()

	case GalleriesModelSortMsg:
		switch msg.Field {
		case "random":
			return s.PushState(func(sm *GalleriesModel) {
				sm.sort = stash.RandomSort()
			})
		case "date":
			return s.PushState(func(sm *GalleriesModel) {
				sm.sort = "date"
				sm.sortDirection = stash.SortDirectionAsc
			})
		// TODO it's probably the case that we want to parse this in the Interpret step rather than here.  We can just
		// enumerate fields we're interested in for the time being.
		case "-date":
			return s.PushState(func(sm *GalleriesModel) {
				sm.sort = "date"
				sm.sortDirection = stash.SortDirectionDesc
			})
		}

	case GalleriesModelSkipMsg:
		if s.pageState.Skip(msg.Count) {
			return &s, s.updateCmd()
		}

	case GalleriesModelUndoMsg:
		return s.Pop()

	case tea.KeyMsg:
		// TODO this is probably not where this ends up, instead we probably have some additional part of the TabModel
		// interface that exposes keymaps (maybe).  I'll slot this in here now and it can return an execute command.
		if cmd, ok := ScenesModelDefaultKeymap[msg.String()]; ok {
			return &s, func() tea.Msg { return ui.CommandExecMsg{Command: cmd} }
		}

	case tea.WindowSizeMsg:
		s.screen = Size{
			Width:  msg.Width,
			Height: msg.Height,
		}
		s.SetHeight(msg.Height)
		return &s, s.updateCmd()

	case galleriesMsg:
		s.galleries, s.pageState.total = msg.galleries, msg.total
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
		if s.pageState.index == i {
			rows[i].Background = &ColorRowSelected
		}
	}

	leftStatus := []string{
		s.pageState.String(),
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
		Page:      m.pageState.page + 1,
		PerPage:   m.pageState.PerPage,
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
