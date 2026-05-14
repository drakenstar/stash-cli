package app

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
)

type resolvedTagIDsMsg struct {
	ids []string
}

func resolveTagInputs(ctx context.Context, srv interface {
	TagFindByName(context.Context, string) (stash.Tag, error)
}, inputs []string) ([]string, error) {
	ids := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if input == "" {
			continue
		}
		if isLikelyEntityID(input) {
			ids = append(ids, input)
			continue
		}

		tag, err := srv.TagFindByName(ctx, input)
		if err != nil {
			return nil, err
		}
		ids = append(ids, tag.ID)
	}
	return ids, nil
}

func isLikelyEntityID(input string) bool {
	return input != "" && strings.IndexFunc(input, func(r rune) bool {
		return !unicode.IsDigit(r)
	}) == -1
}

func (s *cmdService) ResolveTags(inputs []string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		ids, err := resolveTagInputs(context.Background(), s, inputs)
		if err != nil {
			return ErrorMsg{fmt.Errorf("tag resolution failed: %w", err)}
		}
		return resolvedTagIDsMsg{ids: ids}
	})
}

func (s *cmdServiceWithID) ResolveTags(inputs []string) tea.Cmd {
	return s.withID(s.s.ResolveTags(inputs))
}

func (s *cmdService) TagFindByName(ctx context.Context, name string) (stash.Tag, error) {
	if tag, err := s.cache.GetTagByName(name); err == nil {
		return tag, nil
	}

	tag, err := s.Stash.TagFindByName(ctx, name)
	if err != nil {
		return stash.Tag{}, err
	}
	s.cache.CacheTag(tag)
	return tag, nil
}
