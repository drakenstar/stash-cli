package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/drakenstar/stash-cli/stash"
)

// cachingStash is a stash.Stash implementation that caches response data on some fetches into a cacheLookup
type cachingStash struct {
	stash.Stash
	cache *cacheLookup
}

func (s *cachingStash) Scenes(ctx context.Context, f stash.FindFilter, sf stash.SceneFilter) ([]stash.Scene, int, error) {
	scenes, count, err := s.Stash.Scenes(ctx, f, sf)
	s.cache.CacheScenes(scenes)
	return scenes, count, err
}

func (s *cachingStash) Galleries(ctx context.Context, f stash.FindFilter, gf stash.GalleryFilter) ([]stash.Gallery, int, error) {
	galleries, count, err := s.Stash.Galleries(ctx, f, gf)
	s.cache.CacheGalleries(galleries)
	return galleries, count, err
}

// cacheLookup is a StashLookup implementation that caches entities by ID.
type cacheLookup struct {
	mu sync.RWMutex

	performers map[string]stash.Performer
	studios    map[string]stash.Studio
	tags       map[string]stash.Tag
}

func newCacheLookup() *cacheLookup {
	c := &cacheLookup{
		performers: make(map[string]stash.Performer),
		studios:    make(map[string]stash.Studio),
		tags:       make(map[string]stash.Tag),
	}
	return c
}

func (s *cacheLookup) CacheScenes(scenes []stash.Scene) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// cache performers, studio, and tags (with performer tags)
	for _, sc := range scenes {
		for _, p := range sc.Performers {
			s.performers[p.ID] = p
			for _, t := range p.Tags {
				s.tags[t.ID] = t
			}
		}

		for _, t := range sc.Tags {
			s.tags[t.ID] = t
		}

		s.studios[sc.Studio.ID] = sc.Studio
	}
}

func (s *cacheLookup) CacheGalleries(galleries []stash.Gallery) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, g := range galleries {
		for _, p := range g.Performers {
			s.performers[p.ID] = p
			for _, t := range p.Tags {
				s.tags[t.ID] = t
			}
		}

		for _, t := range g.Tags {
			s.tags[t.ID] = t
		}

		s.studios[g.Studio.ID] = g.Studio
	}
}

func (s *cacheLookup) GetStudio(id string) (stash.Studio, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if studio, ok := s.studios[id]; ok {
		return studio, nil
	}
	return stash.Studio{}, fmt.Errorf("studio not cached")
}

func (s *cacheLookup) GetTag(id string) (stash.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if tag, ok := s.tags[id]; ok {
		return tag, nil
	}
	return stash.Tag{}, fmt.Errorf("tag not cached")
}

func (s *cacheLookup) GetPerformer(id string) (stash.Performer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if performer, ok := s.performers[id]; ok {
		return performer, nil
	}
	return stash.Performer{}, fmt.Errorf("tag not cached")
}
