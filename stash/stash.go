package stash

import (
	"context"

	"github.com/hasura/go-graphql-client"
)

type Stash interface {
	Scenes(context.Context, FindFilter, SceneFilter) ([]Scene, int, error)
	DeleteScene(context.Context, string) (bool, error)

	Galleries(context.Context, FindFilter, GalleryFilter) ([]Gallery, int, error)
	GalleryUpdate(context.Context, GalleryUpdate) (Gallery, error)

	PerformersAll(context.Context) ([]PerformerSummary, error)
	PerformerCreate(context.Context, PerformerCreate) (Performer, error)
	PerformerGet(context.Context, string) (Performer, error)

	TagGet(context.Context, string) (Tag, error)
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

func tagListsEqual(a, b []Tag) bool {
	set1 := make(map[string]struct{})
	set2 := make(map[string]struct{})

	for _, item := range a {
		set1[item.ID] = struct{}{}
	}

	for _, item := range b {
		set2[item.ID] = struct{}{}
	}

	if len(set1) != len(set2) {
		return false
	}

	for id := range set1 {
		if _, exists := set2[id]; !exists {
			return false
		}
	}

	return true
}

func performerListsEqual(a, b []Performer) bool {
	set1 := make(map[string]struct{})
	set2 := make(map[string]struct{})

	for _, item := range a {
		set1[item.ID] = struct{}{}
	}

	for _, item := range b {
		set2[item.ID] = struct{}{}
	}

	if len(set1) != len(set2) {
		return false
	}

	for id := range set1 {
		if _, exists := set2[id]; !exists {
			return false
		}
	}

	return true
}
