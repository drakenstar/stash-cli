package app

import (
	"context"
	"fmt"
	"slices"
	"strings"
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

func (s *cachingStash) TagsAll(ctx context.Context) ([]stash.Tag, error) {
	tags, err := s.Stash.TagsAll(ctx)
	if err == nil {
		s.cache.CacheTags(tags)
	}
	return tags, err
}

// cacheLookup is a StashLookup implementation that caches entities by ID.
type cacheLookup struct {
	mu sync.RWMutex

	performers map[string]stash.Performer
	studios    map[string]stash.Studio
	tags       map[string]stash.Tag
	tagNames   map[string]string
	tagsLoaded bool
}

func newCacheLookup() *cacheLookup {
	c := &cacheLookup{
		performers: make(map[string]stash.Performer),
		studios:    make(map[string]stash.Studio),
		tags:       make(map[string]stash.Tag),
		tagNames:   make(map[string]string),
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
				s.cacheTagLocked(t)
			}
		}

		for _, t := range sc.Tags {
			s.cacheTagLocked(t)
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
				s.cacheTagLocked(t)
			}
		}

		for _, t := range g.Tags {
			s.cacheTagLocked(t)
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

func (s *cacheLookup) GetTagByName(name string) (stash.Tag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.tagNames[name]
	if !ok {
		return stash.Tag{}, fmt.Errorf("tag not cached")
	}
	tag, ok := s.tags[id]
	if !ok {
		return stash.Tag{}, fmt.Errorf("tag not cached")
	}
	return tag, nil
}

func (s *cacheLookup) CacheTag(tag stash.Tag) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cacheTagLocked(tag)
}

func (s *cacheLookup) CacheTags(tags []stash.Tag) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, tag := range tags {
		s.cacheTagLocked(tag)
	}
	s.tagsLoaded = true
}

func (s *cacheLookup) cacheTagLocked(tag stash.Tag) {
	s.tags[tag.ID] = tag
	if tag.Name != "" {
		s.tagNames[tag.Name] = tag.ID
	}
}

func (s *cacheLookup) TagsLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tagsLoaded
}

func (s *cacheLookup) TagsByPrefix(prefix string, limit int) []stash.Tag {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefixLower := strings.ToLower(prefix)
	matches := make([]stash.Tag, 0, limit)
	for _, tag := range s.tags {
		if strings.HasPrefix(strings.ToLower(tag.Name), prefixLower) {
			matches = append(matches, tag)
		}
	}

	slices.SortFunc(matches, func(a, b stash.Tag) int {
		return strings.Compare(a.Name, b.Name)
	})
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func (s *cacheLookup) GetPerformer(id string) (stash.Performer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if performer, ok := s.performers[id]; ok {
		return performer, nil
	}
	return stash.Performer{}, fmt.Errorf("tag not cached")
}
