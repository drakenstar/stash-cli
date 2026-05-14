package app

import (
	"context"
	"errors"
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type tagResolutionTestStash struct {
	stash.LocalStash
	byName map[string]stash.Tag
}

func (s tagResolutionTestStash) TagFindByName(_ context.Context, name string) (stash.Tag, error) {
	tag, ok := s.byName[name]
	if !ok {
		return stash.Tag{}, errors.New("not found")
	}
	return tag, nil
}

func TestResolveTagInputsSupportsIDsAndNames(t *testing.T) {
	srv := tagResolutionTestStash{
		byName: map[string]stash.Tag{
			"Foo":      {ID: "12", Name: "Foo"},
			"Test Tag": {ID: "34", Name: "Test Tag"},
		},
	}

	ids, err := resolveTagInputs(context.Background(), &srv, []string{"12", "Foo", "Test Tag"})
	require.NoError(t, err)
	require.Equal(t, []string{"12", "12", "34"}, ids)
}

func TestResolveTagInputsTreatsQuotedNamesAsNames(t *testing.T) {
	srv := tagResolutionTestStash{
		byName: map[string]stash.Tag{
			"123 tag": {ID: "77", Name: "123 tag"},
		},
	}

	ids, err := resolveTagInputs(context.Background(), &srv, []string{"123 tag"})
	require.NoError(t, err)
	require.Equal(t, []string{"77"}, ids)
}
