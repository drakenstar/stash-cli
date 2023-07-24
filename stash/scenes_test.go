package stash

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/require"
)

func TestFindScenes(t *testing.T) {
	doer := &mockEndpoint{
		t: t,
		response: `
	{
		"data": {
			"findScenes": {
				"count": 10,
				"scenes": [
					{
						"id": "1",
						"title": "Scene 1",
						"date": "2023-07-19",
						"details": "Details about scene 1",
						"rating100": 80,
						"organized": true,
						"created_at": "2023-07-01T00:00:00Z",
						"updated_at": "2023-07-18T00:00:00Z",
						"files": [
							{
								"path": "/path/to/file1"
							},
							{
								"path": "/path/to/file2"
							}
						],
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

	scenes, count, err := s.Scenes(ctx, FindFilter{})
	require.NoError(t, err)
	require.Equal(t, []Scene{
		{
			ID:        "1",
			Title:     "Scene 1",
			Date:      "2023-07-19",
			Details:   "Details about scene 1",
			Rating:    80,
			Organized: true,
			CreatedAt: time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2023, 7, 18, 0, 0, 0, 0, time.UTC),
			Files: []VideoFile{
				{Path: "/path/to/file1"},
				{Path: "/path/to/file2"},
			},
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
	}, scenes)
	require.Equal(t, 10, count)
	require.True(t, doer.called)
}

//go:embed schema.graphql
var schemaStr string

func TestRecordPlay(t *testing.T) {
	doer := &mockEndpoint{
		t:        t,
		response: `{}`,
	}

	client := graphql.NewClient("https://example.com/graph", doer)
	s := stash{client}
	ctx := context.Background()

	s.RecordPlay(ctx, "1234")
	require.True(t, doer.called)
}
