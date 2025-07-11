package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
)

// cmdService is compatability layer between the UI and the underlying Stash service.  This is intended to allow UIs
// to call to the service while handling some app level concerns like loading state.
type cmdService struct {
	stash.Stash
}

func (s *cmdService) Scenes(f stash.FindFilter, sf stash.SceneFilter) tea.Cmd {
	return func() tea.Msg {
		scenes, total, err := s.Stash.Scenes(context.Background(), f, sf)
		if err != nil {
			return ErrorMsg{err}
		}
		return scenesMsg{
			scenes: scenes,
			total:  total,
		}
	}
}

func (s *cmdService) DeleteScene(id string) tea.Cmd {
	return func() tea.Msg {
		_, err := s.Stash.DeleteScene(context.Background(), id)
		if err != nil {
			return ErrorMsg{err}
		}
		return sceneDeletedMsg{id}
	}
}

func (s *cmdService) Galleries(f stash.FindFilter, gf stash.GalleryFilter) tea.Cmd {
	return func() tea.Msg {
		galleries, total, err := s.Stash.Galleries(context.Background(), f, gf)
		if err != nil {
			return ErrorMsg{err}
		}
		return galleriesMsg{
			galleries: galleries,
			total:     total,
		}
	}
}

type galleriesMsg struct {
	galleries []stash.Gallery
	total     int
}

type scenesMsg struct {
	scenes []stash.Scene
	total  int
}

type sceneDeletedMsg struct {
	id string
}
