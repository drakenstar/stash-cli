package app

import (
	"testing"
	"time"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

func TestDateFilterValueSetAbsolute(t *testing.T) {
	var value dateFilterValue
	err := value.Set(">2024-01-01")
	require.NoError(t, err)
	require.Equal(t, stash.CriterionModifierGreaterThan, value.Modifier)
	require.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), value.Value)
}

func TestDateFilterValueSetExactDate(t *testing.T) {
	var value dateFilterValue
	err := value.Set("2024-01-01")
	require.NoError(t, err)
	require.Equal(t, stash.CriterionModifierEquals, value.Modifier)
	require.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), value.Value)
}

func TestDateFilterValueSetRelativeYear(t *testing.T) {
	before := time.Now().UTC().AddDate(-1, 0, -1)

	var value dateFilterValue
	err := value.Set(">-1y")
	require.NoError(t, err)

	after := time.Now().UTC().AddDate(-1, 0, 1)
	require.Equal(t, stash.CriterionModifierGreaterThan, value.Modifier)
	require.False(t, value.Value.Before(before))
	require.False(t, value.Value.After(after))
}

func TestDateFilterValueSetInvalid(t *testing.T) {
	var value dateFilterValue
	err := value.Set(">banana")
	require.Error(t, err)
}
