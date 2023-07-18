package app

import (
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

func TestContentState(t *testing.T) {
	s := contentState[stash.Gallery]{
		total: 5,
		filter: stash.FindFilter{
			Page:    1,
			PerPage: 3,
		},
	}

	s.content = []stash.Gallery{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	require.Equal(t, stash.Gallery{ID: "A"}, s.Current())

	require.False(t, s.Skip(1))
	require.Equal(t, stash.Gallery{ID: "B"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)

	require.False(t, s.Skip(1))
	require.Equal(t, stash.Gallery{ID: "C"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)

	require.True(t, s.Skip(2))
	s.content = []stash.Gallery{{ID: "D"}, {ID: "E"}}
	require.Equal(t, stash.Gallery{ID: "E"}, s.Current())
	requirePageAndCount(t, &s, 2, 2)

	require.False(t, s.Skip(0))
	require.Equal(t, stash.Gallery{ID: "E"}, s.Current())
	requirePageAndCount(t, &s, 2, 2)

	require.False(t, s.Skip(-1))
	require.Equal(t, stash.Gallery{ID: "D"}, s.Current())
	requirePageAndCount(t, &s, 2, 2)

	require.True(t, s.Skip(-2))
	s.content = []stash.Gallery{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	require.Equal(t, stash.Gallery{ID: "B"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)

	require.True(t, s.Skip(-2))
	require.Equal(t, stash.Gallery{ID: "A"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)

	require.True(t, s.Skip(5))
	require.Equal(t, stash.Gallery{ID: "A"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)

	require.True(t, s.Skip(-100))
	require.Equal(t, stash.Gallery{ID: "A"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)

	require.True(t, s.Skip(100))
	require.Equal(t, stash.Gallery{ID: "A"}, s.Current())
	requirePageAndCount(t, &s, 1, 2)
}

func requirePageAndCount(t *testing.T, c ContentStater, a, b int) {
	t.Helper()
	page, count := c.PageAndCount()
	require.Equal(t, a, page)
	require.Equal(t, b, count)
}
