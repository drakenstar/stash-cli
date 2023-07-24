package stash

import (
	"context"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/require"
)

func TestFindGalleries(t *testing.T) {
	doer := &mockEndpoint{
		t: t,
		response: `
	{
		"data": {
			"findGalleries": {
				"count": 10,
				"galleries": [
					{
						"id": "1",
						"title": "Gallery 1",
						"date": "2023-07-19",
						"details": "Details about gallery 1",
						"rating100": 80,
						"organized": true,
						"created_at": "2023-07-01T00:00:00Z",
						"updated_at": "2023-07-18T00:00:00Z",
						"image_count": 5,
						"folder": {
							"path": "/path/to/gallery1"
						},
						"files": [],
						"studio": {
							"id": "studio1",
							"name": "Studio 1"
						},
						"tags": [
							{
								"id": "tag1",
								"name": "Tag 1"
							},
							{
								"id": "tag2",
								"name": "Tag 2"
							}
						],
						"performers": [
							{
								"id": "performer1",
								"name": "Performer 1",
								"birthdate": "1990-01-01",
								"gender": "MALE"
							},
							{
								"id": "performer2",
								"name": "Performer 2",
								"birthdate": "1992-01-01",
								"gender": "FEMALE"
							}
						]
					}
				]
			}
		}
	}
	`}

	client := graphql.NewClient("https://example.com/graph", doer)
	s := stash{client}
	ctx := context.Background()

	galleries, count, err := s.Galleries(ctx, FindFilter{}, GalleryFilter{})
	require.NoError(t, err)
	require.Equal(t, []Gallery{
		{
			ID:         "1",
			Title:      "Gallery 1",
			Date:       "2023-07-19",
			Details:    "Details about gallery 1",
			Rating:     80,
			Organized:  true,
			CreatedAt:  time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2023, 7, 18, 0, 0, 0, 0, time.UTC),
			ImageCount: 5,
			Folder:     Folder{Path: "/path/to/gallery1"},
			Files:      []File{},
			Studio: Studio{
				ID:   "studio1",
				Name: "Studio 1",
			},
			Tags: []Tag{
				{
					ID:   "tag1",
					Name: "Tag 1",
				},
				{
					ID:   "tag2",
					Name: "Tag 2",
				},
			},
			Performers: []Performer{
				{
					ID:        "performer1",
					Name:      "Performer 1",
					Birthdate: "1990-01-01",
					Gender:    GenderMale,
				},
				{
					ID:        "performer2",
					Name:      "Performer 2",
					Birthdate: "1992-01-01",
					Gender:    GenderFemale,
				},
			},
		},
	}, galleries)
	require.Equal(t, 10, count)
}
