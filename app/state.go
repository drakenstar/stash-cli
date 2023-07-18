package app

import "github.com/drakenstar/stash-cli/stash"

type appState struct {
	mode filterMode

	sceneFindFilter stash.FindFilter
	scenesState     struct {
		count  int
		scenes []stash.Scene
	}

	galleryFindFilter stash.FindFilter
	galleriesState    struct {
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
		},

		galleryFindFilter: stash.FindFilter{
			Sort:      stash.SortPath,
			Direction: stash.SortDirectionAsc,
		},
	}
}

func (a *appState) PageAndCount() (int, int) {
	switch a.mode {
	case FilterModeScenes:
		return len(a.scenesState.scenes), a.scenesState.count
	case FilterModeGalleries:
		return len(a.galleriesState.galleries), a.galleriesState.count
	default:
		panic("no mode set")
	}
}

type filterMode string

const (
	FilterModeScenes    = "scenes"
	FilterModeGalleries = "galleries"
)
