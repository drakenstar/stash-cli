package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type sceneTagResolveTestService struct{}

func (sceneTagResolveTestService) Scenes(stash.FindFilter, stash.SceneFilter) tea.Cmd { return nil }
func (sceneTagResolveTestService) DeleteScene(string) tea.Cmd                         { return nil }
func (sceneTagResolveTestService) TagScene(stash.Scene, []string) tea.Cmd             { return nil }
func (sceneTagResolveTestService) ResolveTags([]string) tea.Cmd {
	return func() tea.Msg {
		return loadingMsg{
			id:      42,
			payload: resolvedTagIDsMsg{ids: []string{"1", "2"}},
		}
	}
}
func (sceneTagResolveTestService) ResolveStudios([]string) tea.Cmd    { return nil }
func (sceneTagResolveTestService) ResolvePerformers([]string) tea.Cmd { return nil }

type galleryTagResolveTestService struct{}

func (galleryTagResolveTestService) Galleries(stash.FindFilter, stash.GalleryFilter) tea.Cmd {
	return nil
}
func (galleryTagResolveTestService) DeleteGallery(string) tea.Cmd { return nil }
func (galleryTagResolveTestService) TagGallery(stash.Gallery, []string) tea.Cmd {
	return nil
}
func (galleryTagResolveTestService) ResolveTags([]string) tea.Cmd {
	return func() tea.Msg {
		return loadingMsg{
			id:      42,
			payload: resolvedTagIDsMsg{ids: []string{"1", "2"}},
		}
	}
}
func (galleryTagResolveTestService) ResolveStudios([]string) tea.Cmd    { return nil }
func (galleryTagResolveTestService) ResolvePerformers([]string) tea.Cmd { return nil }

type tagResolveTestLookup struct{}

func (tagResolveTestLookup) GetStudio(string) (stash.Studio, error) { return stash.Studio{}, nil }
func (tagResolveTestLookup) GetTag(string) (stash.Tag, error)       { return stash.Tag{}, nil }
func (tagResolveTestLookup) GetPerformer(string) (stash.Performer, error) {
	return stash.Performer{}, nil
}

func TestResolveSceneTagsCmdWrapsLoadingPayload(t *testing.T) {
	m := NewScenesModel(sceneTagResolveTestService{}, tagResolveTestLookup{})

	msg := m.resolveSceneTagsCmd(7, []string{"foo"})()

	routed, ok := msg.(loadingMsg)
	require.True(t, ok)
	require.Equal(t, tabID(42), routed.id)

	resolved, ok := routed.payload.(sceneTagsResolvedMsg)
	require.True(t, ok)
	require.Equal(t, uint64(7), resolved.requestID)
	require.Equal(t, []string{"1", "2"}, resolved.ids)
}

func TestResolveGalleryTagsCmdWrapsLoadingPayload(t *testing.T) {
	m := NewGalleriesModel(galleryTagResolveTestService{}, tagResolveTestLookup{})

	msg := m.resolveGalleryTagsCmd(7, []string{"foo"})()

	routed, ok := msg.(loadingMsg)
	require.True(t, ok)
	require.Equal(t, tabID(42), routed.id)

	resolved, ok := routed.payload.(galleryTagsResolvedMsg)
	require.True(t, ok)
	require.Equal(t, uint64(7), resolved.requestID)
	require.Equal(t, []string{"1", "2"}, resolved.ids)
}
