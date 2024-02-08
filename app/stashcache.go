package app

import (
	"context"

	"github.com/drakenstar/stash-cli/stash"
)

type stashCache struct {
	stash.Stash

	pCache map[string]stash.Performer
}

func newStashCache(s stash.Stash) stashCache {
	c := stashCache{Stash: s}
	c.pCache = make(map[string]stash.Performer)
	return c
}

func (s stashCache) Scenes(ctx context.Context, ff stash.FindFilter, sf stash.SceneFilter) ([]stash.Scene, int, error) {
	scenes, count, err := s.Stash.Scenes(ctx, ff, sf)

	for _, sc := range scenes {
		for _, p := range sc.Performers {
			s.pCache[p.ID] = p
		}
	}

	return scenes, count, err
}

func (s stashCache) PerformerGet(ctx context.Context, id string) (stash.Performer, error) {
	if p, ok := s.pCache[id]; ok {
		return p, nil
	}
	return s.Stash.PerformerGet(ctx, id)
}
