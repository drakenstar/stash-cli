package app

import (
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

func TestCommandSuggestionSetBaseCommands(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	set, needs := m.commandSuggestionSet(":", "re", 2)

	require.Equal(t, suggestionRequirements{}, needs)
	require.Equal(t, 0, set.Start)
	require.Equal(t, 2, set.End)
	require.NotEmpty(t, set.Suggestions)
	require.Equal(t, "refresh", set.Suggestions[0].Display)
	require.Equal(t, "reset", set.Suggestions[1].Display)
}

func TestCommandSuggestionSetFilterArgumentAutocomplete(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	input := "filter st"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Equal(t, len("filter "), set.Start)
	require.Equal(t, len(input), set.End)
	require.NotEmpty(t, set.Suggestions)
	require.Equal(t, "studio", set.Suggestions[0].Display)
	require.Equal(t, "studio=", set.Suggestions[0].Value)
}

func TestCommandSuggestionSetTagAutocomplete(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CacheTags([]stash.Tag{
		{ID: "1", Name: "Foo"},
		{ID: "2", Name: "Foo Bar"},
		{ID: "3", Name: "Food"},
	})

	input := "filter tag=Fo"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Equal(t, len("filter tag="), set.Start)
	require.Equal(t, len(input), set.End)
	require.Len(t, set.Suggestions, 3)
	require.Equal(t, "Foo", set.Suggestions[0].Display)
	require.Equal(t, "Foo", set.Suggestions[0].Value)
	require.Equal(t, "Foo Bar", set.Suggestions[1].Display)
	require.Equal(t, `"Foo Bar"`, set.Suggestions[1].Value)
	require.Equal(t, "Food", set.Suggestions[2].Display)
}

func TestCommandSuggestionSetTagAutocompletePrefersCloserMatches(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CacheTags([]stash.Tag{
		{ID: "1", Name: "Jayden"},
		{ID: "2", Name: "Jade"},
	})

	input := "filter tag=Ja"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Len(t, set.Suggestions, 2)
	require.Equal(t, "Jade", set.Suggestions[0].Display)
	require.Equal(t, "Jayden", set.Suggestions[1].Display)
}

func TestCommandSuggestionSetTagAutocompleteMatchesWords(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CacheTags([]stash.Tag{
		{ID: "1", Name: "Foo Bar"},
		{ID: "2", Name: "Foo Baz"},
	})

	input := `filter tag="foo ba`
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Len(t, set.Suggestions, 2)
	require.Equal(t, "Foo Bar", set.Suggestions[0].Display)
	require.Equal(t, `"Foo Bar"`, set.Suggestions[0].Value)
	require.Equal(t, "Foo Baz", set.Suggestions[1].Display)
}

func TestCommandSuggestionSetTagAutocompleteNeedsLoadedTags(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	input := "filter tag=Fo"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{tags: true}, needs)
	require.Empty(t, set.Suggestions)
}

func TestCommandSuggestionSetStudioAutocomplete(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CacheStudios([]stash.Studio{
		{ID: "1", Name: "Alpha"},
		{ID: "2", Name: "Alpha Beta"},
	})

	input := "filter studio=Al"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Len(t, set.Suggestions, 2)
	require.Equal(t, "Alpha", set.Suggestions[0].Display)
	require.Equal(t, "Alpha", set.Suggestions[0].Value)
	require.Equal(t, `"Alpha Beta"`, set.Suggestions[1].Value)
}

func TestCommandSuggestionSetStudioAutocompleteMatchesWords(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CacheStudios([]stash.Studio{
		{ID: "1", Name: "Alpha Beta"},
		{ID: "2", Name: "Alpha Gamma"},
	})

	input := `filter studio="alpha be`
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Len(t, set.Suggestions, 1)
	require.Equal(t, "Alpha Beta", set.Suggestions[0].Display)
	require.Equal(t, `"Alpha Beta"`, set.Suggestions[0].Value)
}

func TestCommandSuggestionSetPerformerAutocompleteNeedsLoadedPerformers(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	input := "filter performer=Al"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{performers: true}, needs)
	require.Empty(t, set.Suggestions)
}

func TestCommandSuggestionSetPerformerAutocompleteIncludesCurrent(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)

	input := "filter performer=cu"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{performers: true}, needs)
	require.Len(t, set.Suggestions, 1)
	require.Equal(t, "current", set.Suggestions[0].Value)
}

func TestCommandSuggestionSetPerformerAutocompleteMatchesWords(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CachePerformerSummaries([]stash.PerformerSummary{
		{ID: "1", Name: "Jane Doe"},
		{ID: "2", Name: "Jane Smith"},
	})

	input := `filter performer="jane do`
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Len(t, set.Suggestions, 1)
	require.Equal(t, "Jane Doe", set.Suggestions[0].Display)
	require.Equal(t, `"Jane Doe"`, set.Suggestions[0].Value)
}

func TestCommandSuggestionSetPerformerAutocompletePrefersCloserMatches(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.cmdService.cache.CachePerformerSummaries([]stash.PerformerSummary{
		{ID: "1", Name: "Jayden"},
		{ID: "2", Name: "Jade"},
	})

	input := "filter performer=Ja"
	set, needs := m.commandSuggestionSet(":", input, len(input))

	require.Equal(t, suggestionRequirements{}, needs)
	require.Len(t, set.Suggestions, 2)
	require.Equal(t, "Jade", set.Suggestions[0].Display)
	require.Equal(t, "Jayden", set.Suggestions[1].Display)
}
