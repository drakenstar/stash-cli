package app

import (
	"context"
	"errors"
	"fmt"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/hasura/go-graphql-client"
)

// cmdService is compatability layer between the UI and the underlying Stash service.  This is intended to allow UIs
// to call to the service while handling some app level concerns like loading state.
type cmdService struct {
	stash.Stash

	mu           sync.RWMutex
	loadingCount uint
	cache        *cacheLookup
}

func (s *cmdService) loadBegin() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadingCount += 1
}

func (s *cmdService) loadEnd() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.loadingCount > 0 {
		s.loadingCount -= 1
	}
}

func (s *cmdService) withLoadingCount(cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		s.loadBegin()
		defer s.loadEnd()
		return cmd()
	}
}

// AnyLoading returns true if there are any in-flight calls to a stash service.
func (s *cmdService) AnyLoading() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.loadingCount > 0
}

func (s *cmdService) Scenes(f stash.FindFilter, sf stash.SceneFilter) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		scenes, total, err := s.Stash.Scenes(context.Background(), f, sf)
		if err != nil {
			return ErrorMsg{err}
		}
		return scenesMsg{
			scenes: scenes,
			total:  total,
		}
	})
}

func (s *cmdService) DeleteScene(id string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		_, err := s.Stash.DeleteScene(context.Background(), id)
		if err != nil {
			return ErrorMsg{err}
		}
		return sceneDeletedMsg{id}
	})
}

func (s *cmdService) TagScene(scene stash.Scene, names []string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		tags, err := s.resolveOrCreateTags(context.Background(), names)
		if err != nil {
			return ErrorMsg{fmt.Errorf("tag resolution failed: %w", err)}
		}
		tagIDs := mergeTagIDs(scene.Tags, tags)
		updated, err := s.Stash.SceneUpdate(context.Background(), stash.SceneUpdate{
			ID:     graphql.ID(scene.ID),
			TagIDs: tagIDs,
		})
		if err != nil {
			return ErrorMsg{err}
		}
		return sceneTaggedMsg{scene: updated}
	})
}

func (s *cmdService) TagGallery(gallery stash.Gallery, names []string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		tags, err := s.resolveOrCreateTags(context.Background(), names)
		if err != nil {
			return ErrorMsg{fmt.Errorf("tag resolution failed: %w", err)}
		}
		updatedTags := append([]stash.Tag(nil), gallery.Tags...)
		for _, tag := range tags {
			if !tagInList(updatedTags, tag.ID) {
				updatedTags = append(updatedTags, tag)
			}
		}
		updated := gallery
		updated.Tags = updatedTags
		g, err := s.Stash.GalleryUpdate(context.Background(), stash.NewGalleryUpdate(gallery, updated))
		if err != nil {
			return ErrorMsg{err}
		}
		return galleryTaggedMsg{gallery: g}
	})
}

func (s *cmdService) DeleteGallery(id string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		_, err := s.Stash.GalleryDelete(context.Background(), id)
		if err != nil {
			return ErrorMsg{err}
		}
		return galleryDeletedMsg{id}
	})
}

func (s *cmdService) TagsAll() tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		_, err := s.Stash.TagsAll(context.Background())
		if err != nil {
			return ErrorMsg{err}
		}
		return tagsLoadedMsg{}
	})
}

func (s *cmdService) StudiosAll() tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		_, err := s.Stash.StudiosAll(context.Background())
		if err != nil {
			return ErrorMsg{err}
		}
		return studiosLoadedMsg{}
	})
}

func (s *cmdService) PerformersAll() tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		_, err := s.Stash.PerformersAll(context.Background())
		if err != nil {
			return ErrorMsg{err}
		}
		return performersLoadedMsg{}
	})
}

func (s *cmdService) Galleries(f stash.FindFilter, gf stash.GalleryFilter) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		galleries, total, err := s.Stash.Galleries(context.Background(), f, gf)
		if err != nil {
			return ErrorMsg{err}
		}
		return galleriesMsg{
			galleries: galleries,
			total:     total,
		}
	})
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

type sceneTaggedMsg struct {
	scene stash.Scene
}

type galleryDeletedMsg struct {
	id string
}

type galleryTaggedMsg struct {
	gallery stash.Gallery
}

type tagsLoadedMsg struct{}
type studiosLoadedMsg struct{}
type performersLoadedMsg struct{}

// cmdServiceWithID is a wrapped cmdService that annotates fetches with an ID so that they can be routed back to the
// correct tab during Update.
type cmdServiceWithID struct {
	s  *cmdService
	id tabID
}

type loadingMsg struct {
	id      tabID
	payload tea.Msg
}

func (s *cmdServiceWithID) withID(cmd tea.Cmd) tea.Cmd {
	return func() tea.Msg {
		return loadingMsg{
			id:      s.id,
			payload: cmd(),
		}
	}
}

func (s *cmdServiceWithID) Scenes(f stash.FindFilter, sf stash.SceneFilter) tea.Cmd {
	return s.withID(s.s.Scenes(f, sf))
}

func (s *cmdServiceWithID) DeleteScene(id string) tea.Cmd {
	return s.withID(s.s.DeleteScene(id))
}

func (s *cmdServiceWithID) TagScene(scene stash.Scene, names []string) tea.Cmd {
	return s.withID(s.s.TagScene(scene, names))
}

func (s *cmdServiceWithID) DeleteGallery(id string) tea.Cmd {
	return s.withID(s.s.DeleteGallery(id))
}

func (s *cmdServiceWithID) TagGallery(gallery stash.Gallery, names []string) tea.Cmd {
	return s.withID(s.s.TagGallery(gallery, names))
}

func (s *cmdServiceWithID) Galleries(f stash.FindFilter, gf stash.GalleryFilter) tea.Cmd {
	return s.withID(s.s.Galleries(f, gf))
}

func (s *cmdService) resolveOrCreateTags(ctx context.Context, names []string) ([]stash.Tag, error) {
	tags := make([]stash.Tag, 0, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		if isLikelyEntityID(name) {
			tag, err := s.Stash.TagGet(ctx, name)
			if err != nil {
				return nil, err
			}
			tags = append(tags, tag)
			continue
		}

		tag, err := s.TagFindByName(ctx, name)
		if err != nil {
			if !errors.Is(err, stash.ErrTagNotFound) {
				return nil, err
			}
			tag, err = s.Stash.TagCreate(ctx, stash.TagCreate{Name: name})
			if err != nil {
				return nil, err
			}
			s.cache.CacheTag(tag)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func mergeTagIDs(existing []stash.Tag, added []stash.Tag) []graphql.ID {
	ids := make([]graphql.ID, 0, len(existing)+len(added))
	seen := make(map[string]struct{}, len(existing)+len(added))
	for _, tag := range existing {
		if tag.ID == "" {
			continue
		}
		seen[tag.ID] = struct{}{}
		ids = append(ids, graphql.ID(tag.ID))
	}
	for _, tag := range added {
		if tag.ID == "" {
			continue
		}
		if _, ok := seen[tag.ID]; ok {
			continue
		}
		seen[tag.ID] = struct{}{}
		ids = append(ids, graphql.ID(tag.ID))
	}
	return ids
}

func tagInList(tags []stash.Tag, id string) bool {
	for _, tag := range tags {
		if tag.ID == id {
			return true
		}
	}
	return false
}
