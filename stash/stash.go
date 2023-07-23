package stash

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/hasura/go-graphql-client"
)

type Stash interface {
	Scenes(context.Context, FindFilter) ([]Scene, int, error)
	Galleries(context.Context, FindFilter) ([]Gallery, int, error)
}

func New(client *graphql.Client) Stash {
	return &stash{client}
}

type stash struct {
	client *graphql.Client
}

type FindFilter struct {
	Query     string `json:"q"`
	Page      int    `json:"page"`
	PerPage   int    `json:"per_page"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
}

func (FindFilter) GetGraphQLType() string {
	return "FindFilterType"
}

const (
	SortDate         = "date"
	SortUpdatedAt    = "updated_at"
	SortCreatedAt    = "created_at"
	SortPath         = "path"
	SortRandomPrefix = "random_"

	SortDirectionAsc  = "ASC"
	SortDirectionDesc = "DESC"
)

func RandomSort() string {
	return fmt.Sprintf("%s%08d", SortRandomPrefix, rand.Intn(100000000))
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
