package stash

import (
	"context"
	"time"
)

type Gallery struct {
	ID         string      `graphql:"id"`
	Title      string      `graphql:"title"`
	Date       string      `graphql:"date"`
	Details    string      `graphql:"details"`
	Rating     int         `graphql:"rating100"`
	Organized  bool        `graphql:"organized"`
	Folder     Folder      `graphql:"folder"`
	Files      []File      `graphql:"files"`
	CreatedAt  time.Time   `graphql:"created_at"`
	UpdatedAt  time.Time   `graphql:"updated_at"`
	ImageCount int         `graphql:"image_count"`
	Studio     Studio      `graphql:"studio"`
	Tags       []Tag       `graphql:"tags"`
	Performers []Performer `graphql:"performers"`
}

func (g Gallery) FilePath() string {
	if g.Folder.Path != "" {
		return g.Folder.Path
	}
	if len(g.Files) > 0 {
		return g.Files[0].Path
	}
	panic("no file found for gallery")
}

type galleriesQuery struct {
	FindGalleries struct {
		Count     int
		Galleries []Gallery
	} `graphql:"findGalleries(filter: $filter, gallery_filter: $gallery_filter)"`
}

func (s *stash) Galleries(ctx context.Context, filter FindFilter, galleryFilter GalleryFilter) ([]Gallery, int, error) {
	resp := galleriesQuery{}
	err := s.client.Query(ctx, &resp, map[string]any{
		"filter":         filter,
		"gallery_filter": galleryFilter,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.FindGalleries.Galleries, resp.FindGalleries.Count, nil
}
