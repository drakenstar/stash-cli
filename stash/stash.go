package stash

import (
	"context"

	"github.com/Khan/genqlient/graphql"
)

type Stash interface {
	Stats(context.Context) (Stats, error)
	Scenes(context.Context) ([]Scene, error)
	Galleries(context.Context) ([]Gallery, error)
}

func New(client graphql.Client) Stash {
	return &stash{client}
}

type stash struct {
	client graphql.Client
}

type Stats struct {
	SceneCount     int `json:"scene_count"`
	GalleryCount   int `json:"gallery_count"`
	PerformerCount int `json:"performer_count"`
}

func (s *stash) Stats(ctx context.Context) (Stats, error) {
	resp, err := GetStats(ctx, s.client)
	if err != nil {
		return Stats{}, err
	}
	return Stats{
		SceneCount:     resp.Stats.Scene_count,
		GalleryCount:   resp.Stats.Gallery_count,
		PerformerCount: resp.Stats.Performer_count,
	}, nil
}
