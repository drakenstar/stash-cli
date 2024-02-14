package app

import (
	"context"

	"github.com/drakenstar/stash-cli/stash"
)

type stashCache struct {
	stash.Stash

	performers map[string]stash.Performer
	studios    map[string]stash.Studio
	tags       map[string]stash.Tag
}

func newStashCache(s stash.Stash) stashCache {
	c := stashCache{
		Stash:      s,
		performers: make(map[string]stash.Performer),
		studios:    make(map[string]stash.Studio),
		tags:       make(map[string]stash.Tag),
	}
	return c
}

func (s stashCache) cachePerformer(p stash.Performer) {
	s.performers[p.ID] = p
	s.cacheTags(p.Tags)
}

func (s stashCache) cacheStudio(t stash.Studio) {
	s.studios[t.ID] = t
}

func (s stashCache) cacheTags(ts []stash.Tag) {
	for _, t := range ts {
		s.tags[t.ID] = t
	}
}

func (s stashCache) Scenes(ctx context.Context, ff stash.FindFilter, sf stash.SceneFilter) ([]stash.Scene, int, error) {
	scenes, count, err := s.Stash.Scenes(ctx, ff, sf)

	// cache performers, studio, and tags (with performer tags)
	for _, sc := range scenes {
		for _, p := range sc.Performers {
			s.cachePerformer(p)
		}
		s.cacheTags(sc.Tags)
		s.cacheStudio(sc.Studio)
	}

	return scenes, count, err
}

func (s stashCache) Galleries(ctx context.Context, f stash.FindFilter, g stash.GalleryFilter) ([]stash.Gallery, int, error) {
	galleries, count, err := s.Stash.Galleries(ctx, f, g)

	for _, g := range galleries {
		for _, p := range g.Performers {
			s.cachePerformer(p)
		}
		s.cacheTags(g.Tags)
		s.cacheStudio(g.Studio)
	}

	return galleries, count, err
}

func (s stashCache) PerformerGet(ctx context.Context, id string) (stash.Performer, error) {
	if p, ok := s.performers[id]; ok {
		return p, nil
	}
	return s.Stash.PerformerGet(ctx, id)
}
