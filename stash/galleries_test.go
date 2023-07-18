package stash

import (
	"context"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/stretchr/testify/require"
)

func TestFindGalleries(t *testing.T) {
	doer := mockEndpoint(`{
		"data": {
			"findGalleries": {
				"count": 10,
				"galleries": [
					{
						"id": "1234",
						"title": "testing",
						"folder": {
							"path": "/example/testing.mp4"
						}
					},
					{
						"id": "5678",
						"title": "another test",
						"folder": {
							"path": "/example/another_test.mp4"
						}
					},
					{
						"id": "9012",
						"title": "third test",
						"folder": {
							"path": "/example/third_test.mp4"
						}
					},
					{
						"id": "4321",
						"title": "fourth test",
						"folder": {
							"path": "/example/fourth_test.mp4"
						}
					},
					{
						"id": "7890",
						"title": "fifth test",
						"folder": {
							"path": "/example/fifth_test.mp4"
						}
					}
				]
			}
		}
	}`)

	client := graphql.NewClient("https://example.com/graph", doer)
	s := stash{client}
	ctx := context.Background()

	galleries, count, err := s.Galleries(ctx, FindFilter{})
	require.NoError(t, err)
	require.Equal(t, []Gallery{
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
	}, galleries)
	require.Equal(t, 10, count)
}
