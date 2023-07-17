package stash

import (
	"context"
)

type Scene struct {
	ID    string
	Title string
	File  string
}

func (s *stash) Scenes(ctx context.Context, filter FindFilter) ([]Scene, error) {
	resp, err := FindScenes(ctx, s.client, filter)
	if err != nil {
		return nil, err
	}

	scenes := make([]Scene, len(resp.FindScenes.Scenes))
	for i, s := range resp.FindScenes.Scenes {
		scenes[i] = Scene{
			ID:    s.Id,
			Title: s.Title,
			File:  s.Files[0].Path,
		}
	}
	return scenes, nil
}
