package stash

import (
	"context"
	"fmt"
	"time"

	"github.com/hasura/go-graphql-client"
)

type Gallery struct {
	ID         string      `graphql:"id"`
	Title      string      `graphql:"title"`
	URL        string      `graphql:"url"`
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

func (s *stash) GalleryDelete(ctx context.Context, galleryID string) (bool, error) {
	var m struct {
		GalleryDestroy bool `graphql:"galleryDestroy(input: {ids: [$id], delete_file: true, delete_generated: true})"`
	}
	variables := map[string]any{
		"id": graphql.ID(galleryID),
	}
	err := s.client.Mutate(ctx, &m, variables)
	return m.GalleryDestroy, err
}

type GalleryUpdate struct {
	ClientMutationID *string      `graphql:"clientMutationId"`
	ID               graphql.ID   `graphql:"id"`
	Title            *string      `graphql:"title"`
	URL              *string      `graphql:"url"`
	Date             *string      `graphql:"date"`
	Details          *string      `graphql:"details"`
	Rating           *int         `graphql:"rating100"`
	Organized        *bool        `graphql:"organized"`
	SceneIDs         []graphql.ID `graphql:"scene_ids"`
	StudioID         *graphql.ID  `graphql:"studio_id"`
	TagIDs           []graphql.ID `graphql:"tag_ids"`
	PerformerIDs     []graphql.ID `graphql:"performer_ids"`
	PrimaryFileID    *graphql.ID  `graphql:"primary_file_id"`
}

// NewGalleryUpdate does a diff of an old and new Gallery and returns a GalleryUpdate that can be passed to
// stash.GalleryUpdate.  A panic will occur if the IDs of the galleries do not match.
func NewGalleryUpdate(gOld, gNew Gallery) GalleryUpdate {
	g := GalleryUpdate{
		ID: graphql.ID(gNew.ID),
	}

	if gOld.ID != gNew.ID {
		panic(fmt.Errorf("galleries do not have the same id old: %s new: %s", gOld.ID, gNew.ID))
	}

	if gOld.Title != gNew.Title {
		g.Title = &gNew.Title
	}
	if gOld.URL != gNew.URL {
		g.URL = &gNew.URL
	}
	if gOld.Date != gNew.Date {
		g.Date = &gNew.Date
	}
	if gOld.Details != gNew.Details {
		g.Details = &gNew.Details
	}
	if gOld.Rating != gNew.Rating {
		g.Rating = &gNew.Rating
	}
	if gOld.Organized != gNew.Organized {
		g.Organized = &gNew.Organized
	}
	if gOld.Studio.ID != gNew.Studio.ID {
		id := graphql.ID(gNew.Studio.ID)
		g.StudioID = &id
	}
	if !tagListsEqual(gOld.Tags, gNew.Tags) {
		tagIDs := make([]graphql.ID, len(gNew.Tags))
		for i, t := range gNew.Tags {
			tagIDs[i] = graphql.ID(t.ID)
		}
		g.TagIDs = tagIDs
	}
	if !performerListsEqual(gOld.Performers, gNew.Performers) {
		performerIDs := make([]graphql.ID, len(gNew.Performers))
		for i, t := range gNew.Performers {
			performerIDs[i] = graphql.ID(t.ID)
		}
		g.PerformerIDs = performerIDs
	}

	return g
}

func (s *stash) GalleryUpdate(ctx context.Context, g GalleryUpdate) (Gallery, error) {
	var m struct {
		GalleryUpdate Gallery `graphql:"galleryUpdate(input: $input)"`
	}
	err := s.client.Mutate(ctx, &m, map[string]any{"input": g})
	return m.GalleryUpdate, err
}
