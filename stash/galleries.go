package stash

import (
	"context"

	"github.com/machinebox/graphql"
)

type Gallery struct {
	ID    string
	Title string
	Files []struct {
		Path string
	}
}

func (s *stash) Galleries(ctx context.Context) ([]Gallery, error) {
	req := graphql.NewRequest(`
		query {
			findGalleries {
				count
				galleries {
					id
					title
					files {
						path
					}
				}
			}
		}
	`)

	var resp struct {
		FindGalleries struct {
			Count     int
			Galleries []Gallery
		}
	}
	err := s.Run(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.FindGalleries.Galleries, nil
}
