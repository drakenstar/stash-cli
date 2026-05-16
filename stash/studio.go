package stash

import "context"

type allStudiosQuery struct {
	Studios []Studio `graphql:"allStudios"`
}

func (s stash) StudiosAll(ctx context.Context) ([]Studio, error) {
	resp := allStudiosQuery{
		Studios: make([]Studio, 0),
	}
	err := s.client.Query(ctx, &resp, nil)
	if err != nil {
		return nil, err
	}
	return resp.Studios, nil
}
