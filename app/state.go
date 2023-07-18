package app

import (
	"math"

	"github.com/drakenstar/stash-cli/stash"
)

type stats struct {
	Page      int
	PageCount int
	Index     int
	Total     int
}

type ContentStater interface {
	Stats() stats
	Current() any
	Skip(int) bool
	SetQuery(string)
	SetSort(string)
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

func (s *contentState[T]) Stats() stats {
	return stats{
		Page:      s.filter.Page,
		PageCount: int(math.Ceil(float64(s.total) / float64(s.filter.PerPage))),
		Index:     s.index,
		Total:     s.total,
	}
}

func (s *contentState[T]) SetQuery(query string) {
	s.index = 0
	s.total = 0
	s.content = make([]T, 0)
	s.filter.Page = 1
	s.filter.Query = query
}

func (s *contentState[T]) SetSort(sort string) {
	s.index = 0
	s.total = 0
	s.content = make([]T, 0)
	s.filter.Page = 1
	s.filter.Sort = sort
}

// Skip advances the current index by count places and returns a boolean as to whether the index has gone outside the
// bounds of our loaded content indicating that the state of s.filter.Page has been updated and s.content needs to be
// re-queried.
// If the relative position of index is outside the bounds of our total content, then we just reset to page 1 index 0.
// Skip can also traverse backwards.
func (s *contentState[T]) Skip(count int) bool {
	s.index += count

	totalIndex := (s.filter.Page-1)*s.filter.PerPage + s.index

	// We're outside the bounds of our total content and will reset to the start.
	if totalIndex >= s.total || totalIndex < 0 {
		s.index = 0
		s.filter.Page = 1
		return true
	}

	// We're outside the bounds of our loaded content and will update page and index values.
	if s.index >= len(s.content) {
		pageSkip := s.index / s.filter.PerPage
		s.index -= s.filter.PerPage * pageSkip
		s.filter.Page += pageSkip
		return true
	} else if s.index < 0 {
		pageSkip := (int(math.Abs(float64(s.index))) / s.filter.PerPage) + 1
		s.index += s.filter.PerPage * pageSkip
		s.filter.Page -= pageSkip
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

type filterMode string

const (
	FilterModeScenes    = "scenes"
	FilterModeGalleries = "galleries"
)
