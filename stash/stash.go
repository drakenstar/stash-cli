package stash

import (
	"context"

	"github.com/machinebox/graphql"
)

type GraphQLClient interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type Stash interface {
	Stats(context.Context) (Stats, error)
	Scenes(context.Context) ([]Scene, error)
	Galleries(context.Context) ([]Gallery, error)
}

func New(client GraphQLClient) Stash {
	return &stash{client}
}

type stash struct {
	GraphQLClient
}

type Stats struct {
	SceneCount     int `json:"scene_count"`
	GalleryCount   int `json:"gallery_count"`
	PerformerCount int `json:"performer_count"`
}

func (s *stash) Stats(ctx context.Context) (Stats, error) {
	req := graphql.NewRequest(`
		query {
			stats {
				scene_count
				scenes_size
				gallery_count
				performer_count
			}
		}
	`)

	var resp struct {
		Stats Stats
	}

	err := s.Run(ctx, req, &resp)
	if err != nil {
		return Stats{}, err
	}

	return resp.Stats, nil
}
