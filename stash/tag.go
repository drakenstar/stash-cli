package stash

import "context"

type Tag struct {
	ID   string `graphql:"id"`
	Name string `graphql:"name"`
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
