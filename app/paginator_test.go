package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaginator(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		p := NewPaginator(10)
		require.Equal(t, 0, p.index)
		require.Equal(t, 0, p.total)
		require.Equal(t, 1, p.page)
		require.Equal(t, 10, p.perPage)
	})
	t.Run("skip", func(t *testing.T) {
		p := &paginator{
			total:   5,
			page:    1,
			perPage: 3,
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
	})

	t.Run("stringer", func(t *testing.T) {
		p := &paginator{
			total:   5,
			page:    1,
			perPage: 10,
		}
		require.Equal(t, "1-5 of 5", p.String())

		p.total = 15
		p.page = 2
		require.Equal(t, "11-15 of 15", p.String())
	})
}

func requireIndexAndPage(t *testing.T, p *paginator, idx, pg int) {
	t.Helper()
	require.Equal(t, idx, p.index)
	require.Equal(t, pg, p.page)
}
