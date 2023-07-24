package stash

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Stash interface {
	Scenes(context.Context, FindFilter, SceneFilter) ([]Scene, int, error)
	Galleries(context.Context, FindFilter, GalleryFilter) ([]Gallery, int, error)
}

func New(client *graphql.Client) Stash {
	return &stash{client}
}

type stash struct {
	client *graphql.Client
}

type Studio struct {
	ID   string `graphql:"id"`
	Name string `graphql:"name"`
}

type Tag struct {
	ID   string `graphql:"id"`
	Name string `graphql:"name"`
}

type Folder struct {
	Path string `graphql:"path"`
}

type File struct {
	Path string `graphql:"path"`
	Size int64  `graphql:"size"`
}

type VideoFile struct {
	Path     string  `graphql:"path"`
	Duration float64 `graphql:"duration"`
	Size     int64   `graphql:"size"`
}
