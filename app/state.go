package app

import (
	"math"

	"github.com/drakenstar/stash-cli/stash"
)

type ContentStater interface {
	PageAndCount() (int, int)
	Current() any
	Skip(int) bool
}

type contentState[T stash.Gallery | stash.Scene] struct {
	index         int
	total         int
	filter        stash.FindFilter
	defaultFilter stash.FindFilter
	content       []T
}

func (s *contentState[T]) Init() {
	s.filter = s.defaultFilter
	s.index = 0
	s.total = 0
	s.content = make([]T, 0)
}

func (s *contentState[T]) Current() any {
	return s.content[s.index]
}

func (s *contentState[T]) PageAndCount() (int, int) {
	return s.filter.Page, int(math.Ceil(float64(s.total) / float64(s.filter.PerPage)))
}

func (s *contentState[T]) Skip(count int) bool {
	s.index += count

	totalIndex := (s.filter.Page-1)*s.filter.PerPage + s.index
	if totalIndex >= s.total {
		// We're at the end of our content and should loop to start.
		s.index = 0
		s.filter.Page = 1
		return true
	} else if s.index >= s.filter.PerPage {
		// We're at the end of the current page, advance 1.
		s.index -= s.filter.PerPage
		s.filter.Page += 1
		return true
	}

	return false
}

type appState struct {
	mode filterMode

	// Embedded interface acts as a proxy for either state type as contentState
	// supports the ContentStater interface.
	ContentStater
	scenesState    contentState[stash.Scene]
	galleriesState contentState[stash.Gallery]
}

func newAppState() *appState {
	a := &appState{
		scenesState: contentState[stash.Scene]{
			defaultFilter: stash.FindFilter{
				Sort:      stash.SortDate,
				Direction: stash.SortDirectionDesc,
				Page:      1,
				PerPage:   40,
			},
		},

		galleriesState: contentState[stash.Gallery]{
			defaultFilter: stash.FindFilter{
				Sort:      stash.SortPath,
				Direction: stash.SortDirectionAsc,
				Page:      1,
				PerPage:   40,
			},
		},
	}
	a.SetMode(FilterModeScenes)
	return a
}

func (a *appState) SetMode(mode filterMode) {
	a.mode = mode
	switch a.mode {
	case FilterModeScenes:
		a.scenesState.Init()
		a.ContentStater = &a.scenesState
	case FilterModeGalleries:
		a.galleriesState.Init()
		a.ContentStater = &a.galleriesState
	}
}

func (a *appState) Current() any {
	switch a.mode {
	case FilterModeScenes:
		return a.scenesState.Current()
	case FilterModeGalleries:
		return a.galleriesState.Current()
	default:
		panic("no mode set")
	}
}

type filterMode string

const (
	FilterModeScenes    = "scenes"
	FilterModeGalleries = "galleries"
)
