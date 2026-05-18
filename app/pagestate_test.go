package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPaginator(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		p := pageState{
			total:   5,
			PerPage: 3,
		}

		require.False(t, p.Skip(1))
		requireIndexAndPage(t, p, 1, 0)

		require.False(t, p.Skip(1))
		requireIndexAndPage(t, p, 2, 0)

		require.True(t, p.Skip(2))
		requireIndexAndPage(t, p, 1, 1)

		require.False(t, p.Skip(0))
		requireIndexAndPage(t, p, 1, 1)

		require.False(t, p.Skip(-1))
		requireIndexAndPage(t, p, 0, 1)

		require.True(t, p.Skip(-2))
		requireIndexAndPage(t, p, 1, 0)

		require.True(t, p.Skip(-2))
		requireIndexAndPage(t, p, 0, 0)

		require.True(t, p.Skip(5))
		requireIndexAndPage(t, p, 0, 0)

		require.True(t, p.Skip(-100))
		requireIndexAndPage(t, p, 0, 0)

		require.True(t, p.Skip(100))
		requireIndexAndPage(t, p, 0, 0)
	})

	t.Run("stringer", func(t *testing.T) {
		p := pageState{
			total:   5,
			PerPage: 10,
		}
		require.Equal(t, "1-5 of 5", p.String())

		p.total = 15
		p.page = 1
		require.Equal(t, "11-15 of 15", p.String())
	})

	t.Run("delete current adjusts to previous page", func(t *testing.T) {
		p := pageState{
			total:   11,
			PerPage: 10,
			page:    1,
			index:   0,
			opened:  true,
		}

		p.DeleteCurrent()

		require.Equal(t, 10, p.total)
		require.Equal(t, 0, p.page)
		require.Equal(t, 9, p.index)
		require.False(t, p.opened)
	})

	t.Run("delete current resets empty state", func(t *testing.T) {
		p := pageState{
			total:   1,
			PerPage: 10,
			page:    0,
			index:   0,
			opened:  true,
		}

		p.DeleteCurrent()

		require.Equal(t, 0, p.total)
		require.Equal(t, 0, p.page)
		require.Equal(t, 0, p.index)
		require.False(t, p.opened)
	})

	t.Run("set per page preserves absolute position", func(t *testing.T) {
		p := pageState{
			PerPage: 5,
			page:    2,
			index:   3,
			opened:  true,
		}

		p.SetPerPage(20)

		require.Equal(t, 13, p.Position())
		require.Equal(t, 0, p.page)
		require.Equal(t, 13, p.index)
		require.True(t, p.opened)
	})
}

func requireIndexAndPage(t *testing.T, p pageState, idx, pg int) {
	t.Helper()
	require.Equal(t, idx, p.index)
	require.Equal(t, pg, p.page)
}
