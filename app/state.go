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
	Opened() bool
	Skip(int) bool
	SetQuery(string)
	SetSort(string)
}

type contentState[T stash.Gallery | stash.Scene, F stash.GalleryFilter | stash.SceneFilter] struct {
	opened        bool
	index         int
	total         int
	filter        stash.FindFilter
	defaultFilter stash.FindFilter
	content       []T
	contentFilter F
}

func (s *contentState[T, U]) Init() {
	s.filter = s.defaultFilter
	s.index = 0
	s.total = 0
	s.opened = false
	s.content = make([]T, 0)
	s.contentFilter = *new(U)
}

func (s *contentState[T, U]) Current() any {
	return s.content[s.index]
}

func (s *contentState[T, U]) Stats() stats {
	return stats{
		Page:      s.filter.Page,
		PageCount: int(math.Ceil(float64(s.total) / float64(s.filter.PerPage))),
		Index:     s.index,
		Total:     s.total,
	}
}

func (s *contentState[T, U]) SetQuery(query string) {
	s.index = 0
	s.total = 0
	s.content = make([]T, 0)
	s.filter.Page = 1
	s.opened = false
	s.filter.Query = query
}

func (s *contentState[T, U]) SetSort(sort string) {
	s.index = 0
	s.total = 0
	s.content = make([]T, 0)
	s.filter.Page = 1
	s.opened = false
	s.filter.Sort = sort
}

// Skip advances the current index by count places and returns a boolean as to whether the index has gone outside the
// bounds of our loaded content indicating that the state of s.filter.Page has been updated and s.content needs to be
// re-queried.
// If the relative position of index is outside the bounds of our total content, then we just reset to page 1 index 0.
// Skip can also traverse backwards.
func (s *contentState[T, U]) Skip(count int) bool {
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

// Opened returns the state of the opened flag.  If the flag is off, we flip it. Otherwise just return the value.  The
// intention of this flag is to record if an item has been viewed or not for the purpose of auto-advancing.
func (s *contentState[T, U]) Opened() bool {
	if !s.opened {
		s.opened = true
		return false
	}
	return true
}

type appState struct {
	mode filterMode

	// Embedded interface acts as a proxy for either state type as contentState supports the ContentStater interface.
	ContentStater
	scenesState    contentState[stash.Scene, stash.SceneFilter]
	galleriesState contentState[stash.Gallery, stash.GalleryFilter]
}

func newAppState() *appState {
	a := &appState{
		scenesState: contentState[stash.Scene, stash.SceneFilter]{
			defaultFilter: stash.FindFilter{
				Sort:      stash.SortDate,
				Direction: stash.SortDirectionDesc,
				Page:      1,
				PerPage:   40,
			},
		},

		galleriesState: contentState[stash.Gallery, stash.GalleryFilter]{
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

type filterMode int

const (
	FilterModeScenes filterMode = iota
	FilterModeGalleries
)

func (f filterMode) String() string {
	switch f {
	case FilterModeScenes:
		return "Scenes"
	case FilterModeGalleries:
		return "Galleries"
	}
	panic("invalid filter mode")
}

func (f filterMode) Icon() string {
	switch f {
	case FilterModeScenes:
		return "🎬"
	case FilterModeGalleries:
		return "🏙"
	}
	panic("invalid filter mode")
}
