package app

import (
	"math"

	"github.com/drakenstar/stash-cli/stash"
)

type appState struct {
	mode  filterMode
	index int

	sceneFindFilter stash.FindFilter
	scenesState     struct {
		count  int
		scenes []stash.Scene
	}

	galleriesFindFilter stash.FindFilter
	galleriesState      struct {
		count     int
		galleries []stash.Gallery
	}
}

func newAppState() *appState {
	return &appState{
		mode: FilterModeScenes,

		sceneFindFilter: stash.FindFilter{
			Sort:      stash.SortDate,
			Direction: stash.SortDirectionDesc,
			Page:      1,
			PerPage:   40,
		},

		galleriesFindFilter: stash.FindFilter{
			Sort:      stash.SortPath,
			Direction: stash.SortDirectionAsc,
			Page:      1,
			PerPage:   40,
		},
	}
}

func (a *appState) PageAndCount() (int, int) {
	switch a.mode {
	case FilterModeScenes:
		return a.sceneFindFilter.Page, int(math.Ceil(float64(a.scenesState.count) / float64(a.sceneFindFilter.PerPage)))
	case FilterModeGalleries:
		return a.galleriesFindFilter.Page, int(math.Ceil(float64(a.galleriesState.count) / float64(a.galleriesFindFilter.PerPage)))
	default:
		panic("no mode set")
	}
}

func (a *appState) SetMode(mode filterMode) {
	a.mode = mode
	switch a.mode {
	case FilterModeScenes:
		a.sceneFindFilter.Page = 1
	case FilterModeGalleries:
		a.galleriesFindFilter.Page = 1
	}
}

func (a *appState) CurrentContent() any {
	switch a.mode {
	case FilterModeScenes:
		return a.scenesState.scenes[a.index]
	case FilterModeGalleries:
		return a.galleriesState.galleries[a.index]
	default:
		panic("no mode set")
	}
}

func (a *appState) Skip(count int) bool {
	a.index += count
	switch a.mode {
	case FilterModeScenes:
		totalIndex := (a.sceneFindFilter.Page-1)*a.sceneFindFilter.PerPage + a.index
		if totalIndex >= a.scenesState.count { // We're at the end of our content and should loop to start.
			a.index = 0
			a.sceneFindFilter.Page = 1
			return true
		} else if a.index >= a.sceneFindFilter.PerPage { // We're at the end of the current page, advance 1.
			a.index -= a.sceneFindFilter.PerPage
			a.sceneFindFilter.Page += 1
			return true
		}
	case FilterModeGalleries:
		totalIndex := (a.galleriesFindFilter.Page-1)*a.galleriesFindFilter.PerPage + a.index
		if totalIndex >= a.galleriesState.count { // We're at the end of our content and should loop to start.
			a.index = 0
			a.galleriesFindFilter.Page = 1
			return true
		} else if a.index >= a.galleriesFindFilter.PerPage { // We're at the end of the current page, advance 1.
			a.index -= a.galleriesFindFilter.PerPage
			a.galleriesFindFilter.Page += 1
			return true
		}
	default:
		panic("no mode set")
	}
	return false
}

type filterMode string

const (
	FilterModeScenes    = "scenes"
	FilterModeGalleries = "galleries"
)
