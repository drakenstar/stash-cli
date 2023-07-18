package stash

import (
	"context"
)

type Gallery struct {
	ID    string
	Title string
	File  string
}

func (s *stash) Galleries(ctx context.Context, filter FindFilter) ([]Gallery, int, error) {
	resp, err := FindGalleries(ctx, s.client, filter)
	if err != nil {
		return nil, 0, err
	}

	galleries := make([]Gallery, len(resp.FindGalleries.Galleries))
	for i, g := range resp.FindGalleries.Galleries {
		galleries[i] = Gallery{
			ID:    g.Id,
			Title: g.Title,
			File:  g.Folder.Path,
		}
	}
	return galleries, resp.FindGalleries.Count, nil
}
