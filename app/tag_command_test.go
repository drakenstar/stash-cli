package app

import (
	"context"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type sceneTagCommandTestService struct {
	tags []string
}

func (s *sceneTagCommandTestService) Scenes(stash.FindFilter, stash.SceneFilter) tea.Cmd { return nil }
func (s *sceneTagCommandTestService) DeleteScene(string) tea.Cmd                         { return nil }
func (s *sceneTagCommandTestService) ResolveTags([]string) tea.Cmd                       { return nil }
func (s *sceneTagCommandTestService) ResolveStudios([]string) tea.Cmd                    { return nil }
func (s *sceneTagCommandTestService) ResolvePerformers([]string) tea.Cmd                 { return nil }
func (s *sceneTagCommandTestService) TagScene(scene stash.Scene, tags []string) tea.Cmd {
	s.tags = append([]string(nil), tags...)
	updated := scene
	updated.Tags = append(updated.Tags, stash.Tag{ID: "2", Name: tags[0]})
	return func() tea.Msg { return sceneTaggedMsg{scene: updated} }
}

type galleryTagCommandTestService struct {
	tags []string
}

func (s *galleryTagCommandTestService) Galleries(stash.FindFilter, stash.GalleryFilter) tea.Cmd {
	return nil
}
func (s *galleryTagCommandTestService) DeleteGallery(string) tea.Cmd       { return nil }
func (s *galleryTagCommandTestService) ResolveTags([]string) tea.Cmd       { return nil }
func (s *galleryTagCommandTestService) ResolveStudios([]string) tea.Cmd    { return nil }
func (s *galleryTagCommandTestService) ResolvePerformers([]string) tea.Cmd { return nil }
func (s *galleryTagCommandTestService) TagGallery(gallery stash.Gallery, tags []string) tea.Cmd {
	s.tags = append([]string(nil), tags...)
	updated := gallery
	updated.Tags = append(updated.Tags, stash.Tag{ID: "2", Name: tags[0]})
	return func() tea.Msg { return galleryTaggedMsg{gallery: updated} }
}

func TestScenesModelTagCommand(t *testing.T) {
	srv := &sceneTagCommandTestService{}
	m := NewScenesModel(srv, tagResolveTestLookup{})
	m.scenes = []stash.Scene{{ID: "scene-1"}}

	_, cmd := m.Update(ScenesModelTagMsg{Tags: []string{"Foo Bar", "baz"}})
	require.NotNil(t, cmd)
	msg := cmd()
	require.Equal(t, []string{"Foo Bar", "baz"}, srv.tags)

	_, _ = m.Update(msg)
	require.Equal(t, []stash.Tag{{ID: "2", Name: "Foo Bar"}}, m.scenes[0].Tags)
}

func TestGalleriesModelTagCommand(t *testing.T) {
	srv := &galleryTagCommandTestService{}
	m := NewGalleriesModel(srv, tagResolveTestLookup{})
	m.galleries = []stash.Gallery{{ID: "gallery-1"}}

	_, cmd := m.Update(GalleriesModelTagMsg{Tags: []string{"Foo Bar", "baz"}})
	require.NotNil(t, cmd)
	msg := cmd()
	require.Equal(t, []string{"Foo Bar", "baz"}, srv.tags)

	_, _ = m.Update(msg)
	require.Equal(t, []stash.Tag{{ID: "2", Name: "Foo Bar"}}, m.galleries[0].Tags)
}

type tagCreateTestStash struct {
	stash.Stash
	created []string
}

func (s *tagCreateTestStash) TagFindByName(_ context.Context, name string) (stash.Tag, error) {
	if name == "existing" {
		return stash.Tag{ID: "1", Name: name}, nil
	}
	return stash.Tag{}, stash.ErrTagNotFound
}

func (s *tagCreateTestStash) TagCreate(_ context.Context, input stash.TagCreate) (stash.Tag, error) {
	s.created = append(s.created, input.Name)
	return stash.Tag{ID: "2", Name: input.Name}, nil
}

func TestResolveOrCreateTagsCreatesMissingTags(t *testing.T) {
	backend := &tagCreateTestStash{}
	svc := &cmdService{Stash: backend, cache: newCacheLookup()}

	tags, err := svc.resolveOrCreateTags(context.Background(), []string{"existing", "new"})

	require.NoError(t, err)
	require.Equal(t, []stash.Tag{{ID: "1", Name: "existing"}, {ID: "2", Name: "new"}}, tags)
	require.Equal(t, []string{"new"}, backend.created)
}

func TestLoadingErrorMsgIsShownGlobally(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	err := errors.New("boom")

	model, _ := m.Update(loadingMsg{id: m.tabs[0].id, payload: ErrorMsg{error: err}})
	updated := model.(Model)

	require.Equal(t, err, updated.err)
}
