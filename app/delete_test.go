package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type deleteTestService struct{}

func (deleteTestService) Scenes(stash.FindFilter, stash.SceneFilter) tea.Cmd { return nil }
func (deleteTestService) DeleteScene(string) tea.Cmd {
	return func() tea.Msg { return sceneDeletedMsg{id: "scene-1"} }
}
func (deleteTestService) ResolveTags([]string) tea.Cmd                            { return nil }
func (deleteTestService) ResolveStudios([]string) tea.Cmd                         { return nil }
func (deleteTestService) ResolvePerformers([]string) tea.Cmd                      { return nil }
func (deleteTestService) Galleries(stash.FindFilter, stash.GalleryFilter) tea.Cmd { return nil }
func (deleteTestService) DeleteGallery(string) tea.Cmd {
	return func() tea.Msg { return galleryDeletedMsg{id: "gallery-1"} }
}

type deleteTestLookup struct{}

func (deleteTestLookup) GetStudio(string) (stash.Studio, error)       { return stash.Studio{}, nil }
func (deleteTestLookup) GetTag(string) (stash.Tag, error)             { return stash.Tag{}, nil }
func (deleteTestLookup) GetPerformer(string) (stash.Performer, error) { return stash.Performer{}, nil }

func TestScenesModelDeleteRequest(t *testing.T) {
	m := NewScenesModel(deleteTestService{}, deleteTestLookup{})
	m.scenes = []stash.Scene{{
		ID:    "scene-1",
		Title: "Example Scene",
		Files: []stash.VideoFile{{Path: "/library/scenes/example.mp4"}},
	}}

	_, cmd := m.Update(ScenesModelDeleteMsg{})
	require.NotNil(t, cmd)

	msg := cmd()
	request, ok := msg.(deleteRequestMsg)
	require.True(t, ok)
	require.Equal(t, "scene", request.Entity)
	require.Equal(t, "Example Scene", request.Title)
	require.Equal(t, "/library/scenes/example.mp4", request.Path)
	require.False(t, request.SkipConfirm)
	require.NotNil(t, request.DeleteCmd)
}

func TestGalleriesModelDeleteRequest(t *testing.T) {
	m := NewGalleriesModel(deleteTestService{}, deleteTestLookup{})
	m.galleries = []stash.Gallery{{
		ID:     "gallery-1",
		Title:  "Example Gallery",
		Folder: stash.Folder{Path: "/library/galleries/example"},
	}}

	_, cmd := m.Update(GalleriesModelDeleteMsg{Confirm: true})
	require.NotNil(t, cmd)

	msg := cmd()
	request, ok := msg.(deleteRequestMsg)
	require.True(t, ok)
	require.Equal(t, "gallery", request.Entity)
	require.Equal(t, "Example Gallery", request.Title)
	require.Equal(t, "/library/galleries/example", request.Path)
	require.True(t, request.SkipConfirm)
	require.NotNil(t, request.DeleteCmd)
}
