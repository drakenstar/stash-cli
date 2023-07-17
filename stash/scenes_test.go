package stash

import (
	"context"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/stretchr/testify/require"
)

func TestFindScenes(t *testing.T) {
	doer := mockEndpoint(`{
		"data": {
			"findScenes": {
				"count": 5,
				"scenes": [
					{
						"id": "1234",
						"title": "testing",
						"files": [
							{
								"path": "/example/testing.mp4"
							}
						]
					},
					{
						"id": "5678",
						"title": "another test",
						"files": [
							{
								"path": "/example/another_test.mp4"
							}
						]
					},
					{
						"id": "9012",
						"title": "third test",
						"files": [
							{
								"path": "/example/third_test.mp4"
							}
						]
					},
					{
						"id": "4321",
						"title": "fourth test",
						"files": [
							{
								"path": "/example/fourth_test.mp4"
							}
						]
					},
					{
						"id": "7890",
						"title": "fifth test",
						"files": [
							{
								"path": "/example/fifth_test.mp4"
							}
						]
					}
				]
			}
		}
	}`)

	client := graphql.NewClient("https://example.com/graph", doer)
	s := stash{client}
	ctx := context.Background()

	scenes, err := s.Scenes(ctx)
	require.NoError(t, err)
	require.Equal(t, []Scene{
		{
			ID:    "1234",
			Title: "testing",
			File:  "/example/testing.mp4",
		},
		{
			ID:    "5678",
			Title: "another test",
			File:  "/example/another_test.mp4",
		},
		{
			ID:    "9012",
			Title: "third test",
			File:  "/example/third_test.mp4",
		},
		{
			ID:    "4321",
			Title: "fourth test",
			File:  "/example/fourth_test.mp4",
		},
		{
			ID:    "7890",
			Title: "fifth test",
			File:  "/example/fifth_test.mp4",
		},
	}, scenes)
}
