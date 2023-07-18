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
		file := g.Folder.Path
		if file == "" && len(g.Files) > 0 {
			file = g.Files[0].Path
		}
		galleries[i] = Gallery{
			ID:    g.Id,
			Title: g.Title,
			File:  file,
		}
	}
	return galleries, resp.FindGalleries.Count, nil
}
