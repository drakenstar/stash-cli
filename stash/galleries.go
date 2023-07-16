package stash

import (
	"context"
)

type Gallery struct {
	ID    string
	Title string
	File  string
}

func (s *stash) Galleries(ctx context.Context) ([]Gallery, error) {
	resp, err := FindGalleries(ctx, s.client)
	if err != nil {
		return nil, err
	}

	galleries := make([]Gallery, len(resp.FindGalleries.Galleries))
	for i, g := range resp.FindGalleries.Galleries {
		galleries[i] = Gallery{
			ID:    g.Id,
			Title: g.Title,
			File:  g.Files[0].Path,
		}
	}
	return galleries, nil
}
