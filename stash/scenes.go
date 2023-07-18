package stash

import (
	"context"
	"time"
)

type Scene struct {
	ID         string
	Title      string
	Date       string
	Details    string
	Rating     int
	Organized  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	File       string
	Studio     Studio
	Tags       []Tag
	Performers []Performer
}

func (s *stash) Scenes(ctx context.Context, filter FindFilter) ([]Scene, int, error) {
	resp, err := FindScenes(ctx, s.client, filter)
	if err != nil {
		return nil, 0, err
	}

	scenes := make([]Scene, len(resp.FindScenes.Scenes))
	for i, s := range resp.FindScenes.Scenes {
		scenes[i] = Scene{
			ID:        s.Id,
			Title:     s.Title,
			Date:      s.Date,
			Details:   s.Details,
			Rating:    s.Rating100,
			Organized: s.Organized,
			CreatedAt: s.Created_at,
			UpdatedAt: s.Updated_at,
			File:      s.Files[0].Path, // TODO better file handling
			Studio: Studio{
				ID:   s.Studio.Id,
				Name: s.Studio.Name,
			},
		}

		tags := make([]Tag, len(s.Tags))
		for i, t := range s.Tags {
			tags[i] = Tag{
				ID:   t.Id,
				Name: t.Name,
			}
		}
		scenes[i].Tags = tags

		performers := make([]Performer, len(s.Performers))
		for i, p := range s.Performers {
			performers[i] = Performer{
				ID:        p.Id,
				Name:      p.Name,
				Birthdate: p.Birthdate,
				Gender:    p.Gender,
			}
		}
		scenes[i].Performers = performers
	}
	return scenes, resp.FindScenes.Count, nil
}
