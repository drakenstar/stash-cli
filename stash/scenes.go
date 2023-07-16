package stash

import (
	"context"

	"github.com/machinebox/graphql"
)

type Scene struct {
	ID    string
	Title string
	Files []struct {
		Path string
	}
}

func (s *stash) Scenes(ctx context.Context) ([]Scene, error) {
	req := graphql.NewRequest(`
		query {
			findScenes {
				count
				scenes {
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
		FindScenes struct {
			Count  int
			Scenes []Scene
		}
	}
	err := s.Run(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp.FindScenes.Scenes, nil
}
