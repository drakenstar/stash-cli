package stash

import (
	"context"
	"time"
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

func (s *stash) Scenes(ctx context.Context, filter FindFilter) ([]Scene, int, error) {
	var resp struct {
		FindScenes struct {
			Count  int `graphql:"count"`
			Scenes []Scene
		} `graphql:"findScenes(filter: $filter)"`
	}
	err := s.client.Query(ctx, &resp, map[string]any{
		"filter": filter,
	})
	if err != nil {
		return nil, 0, err
	}

	return resp.FindScenes.Scenes, resp.FindScenes.Count, nil
}
