package stash

//go:generate go run github.com/Khan/genqlient

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/Khan/genqlient/graphql"
)

type Stash interface {
	Scenes(context.Context, FindFilter) ([]Scene, int, error)
	Galleries(context.Context, FindFilter) ([]Gallery, int, error)
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

type FindFilter struct {
	Query     string `json:"q,omitempty"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Sort      string `json:"sort,omitempty"`
	Direction string `json:"direction,omitempty"`
}

const (
	SortDate         = "date"
	SortUpdatedAt    = "updated_at"
	SortCreatedAt    = "created_at"
	SortPath         = "path"
	SortRandomPrefix = "random_"

	SortDirectionAsc  = "ASC"
	SortDirectionDesc = "DESC"
)

func RandomSort() string {
	return fmt.Sprintf("%s%08d", SortRandomPrefix, rand.Intn(100000000))
}
