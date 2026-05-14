package app

import (
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

func TestCommandSuggestionSetBaseCommands(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	set, needsTags := m.commandSuggestionSet(":", "re", 2)

	require.False(t, needsTags)
	require.Equal(t, 0, set.Start)
	require.Equal(t, 2, set.End)
	require.NotEmpty(t, set.Suggestions)
	require.Equal(t, "refresh", set.Suggestions[0].Display)
	require.Equal(t, "reset", set.Suggestions[1].Display)
}

func TestCommandSuggestionSetTagAutocomplete(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CacheTags([]stash.Tag{
		{ID: "1", Name: "Foo"},
		{ID: "2", Name: "Foo Bar"},
		{ID: "3", Name: "Food"},
	})

	input := "filter tag=Fo"
	set, needsTags := m.commandSuggestionSet(":", input, len(input))

	require.False(t, needsTags)
	require.Equal(t, len("filter tag="), set.Start)
	require.Equal(t, len(input), set.End)
	require.Len(t, set.Suggestions, 3)
	require.Equal(t, "Foo", set.Suggestions[0].Display)
	require.Equal(t, "Foo", set.Suggestions[0].Value)
	require.Equal(t, "Foo Bar", set.Suggestions[1].Display)
	require.Equal(t, `"Foo Bar"`, set.Suggestions[1].Value)
	require.Equal(t, "Food", set.Suggestions[2].Display)
}

func TestCommandSuggestionSetTagAutocompleteNeedsLoadedTags(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	input := "filter tag=Fo"
	set, needsTags := m.commandSuggestionSet(":", input, len(input))

	require.True(t, needsTags)
	require.Empty(t, set.Suggestions)
}
