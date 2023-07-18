package stash

import (
	"context"
	"time"
)

type Gallery struct {
	ID         string
	Title      string
	Date       string
	Details    string
	Rating     int
	Organized  bool
	File       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ImageCount int
	Studio     Studio
	Tags       []Tag
	Performers []Performer
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
			ID:         g.Id,
			Title:      g.Title,
			Date:       g.Date,
			Details:    g.Details,
			Rating:     g.Rating100,
			Organized:  g.Organized,
			File:       file,
			CreatedAt:  g.Created_at,
			UpdatedAt:  g.Updated_at,
			ImageCount: g.Image_count,
			Studio: Studio{
				ID:   g.Studio.Id,
				Name: g.Studio.Name,
			},
		}

		tags := make([]Tag, len(g.Tags))
		for i, t := range g.Tags {
			tags[i] = Tag{
				ID:   t.Id,
				Name: t.Name,
			}
		}
		galleries[i].Tags = tags

		performers := make([]Performer, len(g.Performers))
		for i, p := range g.Performers {
			performers[i] = Performer{
				ID:        p.Id,
				Name:      p.Name,
				Birthdate: p.Birthdate,
				Gender:    p.Gender,
			}
		}
		galleries[i].Performers = performers
	}
	return galleries, resp.FindGalleries.Count, nil
}
