package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaginator(t *testing.T) {
	p := &paginator{
		Total:   5,
		Page:    1,
		PerPage: 3,
	}

	require.False(t, p.Skip(1))
	requireIndexAndPage(t, p, 1, 1)

	require.False(t, p.Skip(1))
	requireIndexAndPage(t, p, 2, 1)

	require.True(t, p.Skip(2))
	requireIndexAndPage(t, p, 1, 2)

	require.False(t, p.Skip(0))
	requireIndexAndPage(t, p, 1, 2)

	require.False(t, p.Skip(-1))
	requireIndexAndPage(t, p, 0, 2)

	require.True(t, p.Skip(-2))
	requireIndexAndPage(t, p, 1, 1)

	require.True(t, p.Skip(-2))
	requireIndexAndPage(t, p, 0, 1)

	require.True(t, p.Skip(5))
	requireIndexAndPage(t, p, 0, 1)

	require.True(t, p.Skip(-100))
	requireIndexAndPage(t, p, 0, 1)

	require.True(t, p.Skip(100))
	requireIndexAndPage(t, p, 0, 1)
}

func requireIndexAndPage(t *testing.T, p *paginator, idx, pg int) {
	t.Helper()
	require.Equal(t, idx, p.Index)
	require.Equal(t, pg, p.Page)
}
