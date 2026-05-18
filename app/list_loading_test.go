package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type sceneListTestService struct {
	responses [][]stash.Scene
}

func (s *sceneListTestService) Scenes(stash.FindFilter, stash.SceneFilter) tea.Cmd {
	scenes := append([]stash.Scene(nil), s.responses[0]...)
	s.responses = s.responses[1:]
	return func() tea.Msg { return scenesMsg{scenes: scenes, total: 100} }
}

func (s *sceneListTestService) DeleteScene(string) tea.Cmd             { return nil }
func (s *sceneListTestService) TagScene(stash.Scene, []string) tea.Cmd { return nil }
func (s *sceneListTestService) ResolveTags([]string) tea.Cmd           { return nil }
func (s *sceneListTestService) ResolveStudios([]string) tea.Cmd        { return nil }
func (s *sceneListTestService) ResolvePerformers([]string) tea.Cmd     { return nil }

type galleryListTestService struct {
	responses [][]stash.Gallery
}

func (s *galleryListTestService) Galleries(stash.FindFilter, stash.GalleryFilter) tea.Cmd {
	galleries := append([]stash.Gallery(nil), s.responses[0]...)
	s.responses = s.responses[1:]
	return func() tea.Msg { return galleriesMsg{galleries: galleries, total: 100} }
}

func (s *galleryListTestService) DeleteGallery(string) tea.Cmd               { return nil }
func (s *galleryListTestService) TagGallery(stash.Gallery, []string) tea.Cmd { return nil }
func (s *galleryListTestService) ResolveTags([]string) tea.Cmd               { return nil }
func (s *galleryListTestService) ResolveStudios([]string) tea.Cmd            { return nil }
func (s *galleryListTestService) ResolvePerformers([]string) tea.Cmd         { return nil }

func TestScenesModelIgnoresStaleListLoads(t *testing.T) {
	srv := &sceneListTestService{responses: [][]stash.Scene{
		{{ID: "initial"}},
		{{ID: "old"}},
		{{ID: "new"}},
	}}
	m := NewScenesModel(srv, tagResolveTestLookup{})

	oldCmd := m.SetSize(Size{Width: 80, Height: 10})
	newCmd := m.SetSize(Size{Width: 80, Height: 20})

	_, _ = m.Update(newCmd())
	require.Equal(t, "new", m.scenes[0].ID)

	_, _ = m.Update(oldCmd())
	require.Equal(t, "new", m.scenes[0].ID)
}

func TestGalleriesModelIgnoresStaleListLoads(t *testing.T) {
	srv := &galleryListTestService{responses: [][]stash.Gallery{
		{{ID: "initial"}},
		{{ID: "old"}},
		{{ID: "new"}},
	}}
	m := NewGalleriesModel(srv, tagResolveTestLookup{})

	oldCmd := m.SetSize(Size{Width: 80, Height: 10})
	newCmd := m.SetSize(Size{Width: 80, Height: 20})

	_, _ = m.Update(newCmd())
	require.Equal(t, "new", m.galleries[0].ID)

	_, _ = m.Update(oldCmd())
	require.Equal(t, "new", m.galleries[0].ID)
}

func TestLateLoadingMsgForRemovedTabIsIgnored(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	removedID := m.tabs[0].id
	delete(m.tabsByID, removedID)
	m.tabs = nil

	updated, cmd := m.Update(loadingMsg{id: removedID, payload: scenesLoadedMsg{requestID: 1}})

	require.Nil(t, cmd)
	require.IsType(t, Model{}, updated)
}
