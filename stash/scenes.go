package stash

import (
	"context"
	"time"

	"github.com/hasura/go-graphql-client"
)

type Scene struct {
	ID         string      `graphql:"id"`
	Title      string      `graphql:"title"`
	Date       string      `graphql:"date"`
	Details    string      `graphql:"details"`
	Rating     int         `graphql:"rating100"`
	Organized  bool        `graphql:"organized"`
	CreatedAt  time.Time   `graphql:"created_at"`
	UpdatedAt  time.Time   `graphql:"updated_at"`
	Files      []VideoFile `graphql:"files"`
	Studio     Studio      `graphql:"studio"`
	Tags       []Tag       `graphql:"tags"`
	Performers []Performer `graphql:"performers"`
}

func (s Scene) FilePath() string {
	if len(s.Files) > 0 {
		return s.Files[0].Path
	}
	panic("no file path")
}

func (s *stash) Scenes(ctx context.Context, filter FindFilter, sceneFilter SceneFilter) ([]Scene, int, error) {
	var resp struct {
		FindScenes struct {
			Count  int `graphql:"count"`
			Scenes []Scene
		} `graphql:"findScenes(filter: $filter, scene_filter: $scene_filter)"`
	}
	err := s.client.Query(ctx, &resp, map[string]any{
		"filter":       filter,
		"scene_filter": sceneFilter,
	})
	if err != nil {
		return nil, 0, err
	}

	return resp.FindScenes.Scenes, resp.FindScenes.Count, nil
}

func (s *stash) RecordPlay(ctx context.Context, sceneID string) error {
	var m struct {
		SceneIncrementPlayCount int `graphql:"sceneIncrementPlayCount(id: $id)"`
	}
	variables := map[string]any{
		"id": sceneID,
	}
	return s.client.Mutate(ctx, &m, variables)
}

func (s *stash) DeleteScene(ctx context.Context, sceneID string) (bool, error) {
	var m struct {
		Result bool `graphql:"sceneDestroy(input: {id: $id, delete_file: true, delete_generated: true})"`
	}
	variables := map[string]any{
		"id": graphql.ID(sceneID),
	}
	err := s.client.Mutate(ctx, &m, variables)
	return m.Result, err
}
