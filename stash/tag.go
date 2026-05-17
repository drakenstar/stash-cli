package stash

import (
	"context"
	"errors"
	"fmt"

	"github.com/hasura/go-graphql-client"
)

var ErrTagNotFound = errors.New("tag not found")

type Tag struct {
	ID   string `graphql:"id"`
	Name string `graphql:"name"`
}

func (t Tag) EntityID() string {
	return t.ID
}

type findTagQuery struct {
	Tag Tag `graphql:"findTag(id: $id)"`
}

// PerformerGet returns a single performer by ID.
func (s stash) TagGet(ctx context.Context, id string) (Tag, error) {
	resp := findTagQuery{}
	err := s.client.Query(ctx, &resp, map[string]any{"id": id})
	if err != nil {
		return Tag{}, err
	}
	return resp.Tag, nil
}

type findTagsQuery struct {
	FindTags struct {
		Count int   `graphql:"count"`
		Tags  []Tag `graphql:"tags"`
	} `graphql:"findTags(tag_filter: $tag_filter)"`
}

type tagFilter struct {
	Name *StringCriterion `json:"name,omitempty"`
}

func (tagFilter) GetGraphQLType() string {
	return "TagFilterType"
}

func (s stash) TagFindByName(ctx context.Context, name string) (Tag, error) {
	resp := findTagsQuery{}
	err := s.client.Query(ctx, &resp, map[string]any{
		"tag_filter": tagFilter{
			Name: &StringCriterion{
				Value:    name,
				Modifier: CriterionModifierEquals,
			},
		},
	})
	if err != nil {
		return Tag{}, err
	}

	switch resp.FindTags.Count {
	case 0:
		return Tag{}, fmt.Errorf("%w: %s", ErrTagNotFound, name)
	case 1:
		return resp.FindTags.Tags[0], nil
	default:
		return Tag{}, fmt.Errorf("multiple tags found for name: %s", name)
	}
}

type allTagsQuery struct {
	Tags []Tag `graphql:"allTags"`
}

func (s stash) TagsAll(ctx context.Context) ([]Tag, error) {
	resp := allTagsQuery{
		Tags: make([]Tag, 0),
	}
	err := s.client.Query(ctx, &resp, nil)
	if err != nil {
		return nil, err
	}
	return resp.Tags, nil
}

type TagCreate struct {
	Name string `json:"name"`
}

func (TagCreate) GetGraphQLType() string {
	return "TagCreateInput"
}

func (s stash) TagCreate(ctx context.Context, tag TagCreate) (Tag, error) {
	var m struct {
		Tag Tag `graphql:"tagCreate(input: {name: $name})"`
	}
	err := s.client.Mutate(ctx, &m, map[string]any{"name": graphql.String(tag.Name)})
	return m.Tag, err
}
