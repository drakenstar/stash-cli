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

func (s *cachingStash) PerformersAll(ctx context.Context) ([]stash.PerformerSummary, error) {
	performers, err := s.Stash.PerformersAll(ctx)
	if err == nil {
		s.cache.CachePerformerSummaries(performers)
	}
	return performers, err
}

func (s *cachingStash) StudiosAll(ctx context.Context) ([]stash.Studio, error) {
	studios, err := s.Stash.StudiosAll(ctx)
	if err == nil {
		s.cache.CacheStudios(studios)
	}
	return studios, err
}

func (s *cachingStash) TagsAll(ctx context.Context) ([]stash.Tag, error) {
	tags, err := s.Stash.TagsAll(ctx)
	if err == nil {
		s.cache.CacheTags(tags)
	}
	return tags, err
}

func (s *cachingStash) TagCreate(ctx context.Context, input stash.TagCreate) (stash.Tag, error) {
	tag, err := s.Stash.TagCreate(ctx, input)
	if err == nil {
		s.cache.CacheTag(tag)
	}
	return tag, err
}

// cacheLookup is a StashLookup implementation that caches entities by ID.
type cacheLookup struct {
	mu sync.RWMutex

	performers       map[string]stash.Performer
	performerNames   map[string]string
	performersLoaded bool
	studios          map[string]stash.Studio
	studioNames      map[string]string
	studiosLoaded    bool
	tags             map[string]stash.Tag
	tagNames         map[string]string
	tagsLoaded       bool
}

func newCacheLookup() *cacheLookup {
	c := &cacheLookup{
		performers:     make(map[string]stash.Performer),
		performerNames: make(map[string]string),
		studios:        make(map[string]stash.Studio),
		studioNames:    make(map[string]string),
		tags:           make(map[string]stash.Tag),
		tagNames:       make(map[string]string),
	}
	return c
}

func (s *cacheLookup) CacheScenes(scenes []stash.Scene) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// cache performers, studio, and tags (with performer tags)
	for _, sc := range scenes {
		for _, p := range sc.Performers {
			s.cachePerformerLocked(p)
			for _, t := range p.Tags {
				s.cacheTagLocked(t)
			}
		}

		for _, t := range sc.Tags {
			s.cacheTagLocked(t)
		}

		s.cacheStudioLocked(sc.Studio)
	}
}

func (s *cacheLookup) CacheGalleries(galleries []stash.Gallery) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, g := range galleries {
		for _, p := range g.Performers {
			s.cachePerformerLocked(p)
			for _, t := range p.Tags {
				s.cacheTagLocked(t)
			}
		}

		for _, t := range g.Tags {
			s.cacheTagLocked(t)
		}

		s.cacheStudioLocked(g.Studio)
	}
}

func (s *cacheLookup) CachePerformerSummaries(performers []stash.PerformerSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, performer := range performers {
		s.cachePerformerLocked(stash.Performer{
			ID:   performer.ID,
			Name: performer.Name,
		})
	}
	s.performersLoaded = true
}

func (s *cacheLookup) CacheStudios(studios []stash.Studio) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, studio := range studios {
		s.cacheStudioLocked(studio)
	}
	s.studiosLoaded = true
}

func (s *cacheLookup) GetStudio(id string) (stash.Studio, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if studio, ok := s.studios[id]; ok {
		return studio, nil
	}
	return stash.Studio{}, fmt.Errorf("studio not cached")
}

func (s *cacheLookup) GetStudioByName(name string) (stash.Studio, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.studioNames[name]
	if !ok {
		return stash.Studio{}, fmt.Errorf("studio not cached")
	}
	studio, ok := s.studios[id]
	if !ok {
		return stash.Studio{}, fmt.Errorf("studio not cached")
	}
	return studio, nil
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

func (s *cacheLookup) StudiosLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.studiosLoaded
}

func (s *cacheLookup) StudiosByPrefix(prefix string, limit int) []stash.Studio {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matches := make([]stash.Studio, 0, limit)
	for _, studio := range s.studios {
		if wordPrefixMatch(studio.Name, prefix) {
			matches = append(matches, studio)
		}
	}

	slices.SortFunc(matches, func(a, b stash.Studio) int {
		return compareWordPrefixMatches(a.Name, b.Name, prefix)
	})
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
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

	matches := make([]stash.Tag, 0, limit)
	for _, tag := range s.tags {
		if wordPrefixMatch(tag.Name, prefix) {
			matches = append(matches, tag)
		}
	}

	slices.SortFunc(matches, func(a, b stash.Tag) int {
		return compareWordPrefixMatches(a.Name, b.Name, prefix)
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

func (s *cacheLookup) GetPerformerByName(name string) (stash.Performer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, ok := s.performerNames[name]
	if !ok {
		return stash.Performer{}, fmt.Errorf("performer not cached")
	}
	performer, ok := s.performers[id]
	if !ok {
		return stash.Performer{}, fmt.Errorf("performer not cached")
	}
	return performer, nil
}

func (s *cacheLookup) PerformersLoaded() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.performersLoaded
}

func (s *cacheLookup) PerformersByPrefix(prefix string, limit int) []stash.Performer {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matches := make([]stash.Performer, 0, limit)
	for _, performer := range s.performers {
		if wordPrefixMatch(performer.Name, prefix) {
			matches = append(matches, performer)
		}
	}

	slices.SortFunc(matches, func(a, b stash.Performer) int {
		return compareWordPrefixMatches(a.Name, b.Name, prefix)
	})
	if limit > 0 && len(matches) > limit {
		matches = matches[:limit]
	}
	return matches
}

func (s *cacheLookup) cachePerformerLocked(performer stash.Performer) {
	s.performers[performer.ID] = performer
	if performer.Name != "" {
		s.performerNames[performer.Name] = performer.ID
	}
}

func (s *cacheLookup) cacheStudioLocked(studio stash.Studio) {
	s.studios[studio.ID] = studio
	if studio.Name != "" {
		s.studioNames[studio.Name] = studio.ID
	}
}

func wordPrefixMatch(candidate, query string) bool {
	queryWords := strings.Fields(strings.ToLower(query))
	if len(queryWords) == 0 {
		return false
	}

	candidateWords := strings.Fields(strings.ToLower(candidate))
	if len(queryWords) > len(candidateWords) {
		return false
	}

	for i, queryWord := range queryWords {
		if !strings.HasPrefix(candidateWords[i], queryWord) {
			return false
		}
	}
	return true
}

func compareWordPrefixMatches(a, b, query string) int {
	scoreA := wordPrefixScore(a, query)
	scoreB := wordPrefixScore(b, query)
	if scoreA != scoreB {
		return scoreA - scoreB
	}

	wordsA := len(strings.Fields(a))
	wordsB := len(strings.Fields(b))
	if wordsA != wordsB {
		return wordsA - wordsB
	}

	if len(a) != len(b) {
		return len(a) - len(b)
	}

	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

func wordPrefixScore(candidate, query string) int {
	queryWords := strings.Fields(strings.ToLower(query))
	candidateWords := strings.Fields(strings.ToLower(candidate))

	score := 0
	for i, queryWord := range queryWords {
		score += len(candidateWords[i]) - len(queryWord)
	}
	return score
}
