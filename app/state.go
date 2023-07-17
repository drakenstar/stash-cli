package app

import "github.com/drakenstar/stash-cli/stash"

type appState struct {
	sceneFindFilter stash.FindFilter
	scenes          []stash.Scene

	galleryFindFilter stash.FindFilter
	galleries         []stash.Gallery
}
