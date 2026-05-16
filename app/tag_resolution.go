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

type resolvedStudioIDsMsg struct {
	ids []string
}

type resolvedPerformerIDsMsg struct {
	ids []string
}

func resolveEntityInputs[T interface{ EntityID() string }](inputs []string, lookup func(string) (T, error)) ([]string, error) {
	ids := make([]string, 0, len(inputs))
	for _, input := range inputs {
		if input == "" {
			continue
		}
		if isLikelyEntityID(input) {
			ids = append(ids, input)
			continue
		}

		entity, err := lookup(input)
		if err != nil {
			return nil, err
		}
		ids = append(ids, entity.EntityID())
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
		ids, err := resolveEntityInputs(inputs, func(name string) (stash.Tag, error) {
			return s.TagFindByName(context.Background(), name)
		})
		if err != nil {
			return ErrorMsg{fmt.Errorf("tag resolution failed: %w", err)}
		}
		return resolvedTagIDsMsg{ids: ids}
	})
}

func (s *cmdServiceWithID) ResolveTags(inputs []string) tea.Cmd {
	return s.withID(s.s.ResolveTags(inputs))
}

func (s *cmdService) ResolveStudios(inputs []string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		ids, err := resolveEntityInputs(inputs, func(name string) (stash.Studio, error) {
			return s.StudioFindByName(name)
		})
		if err != nil {
			return ErrorMsg{fmt.Errorf("studio resolution failed: %w", err)}
		}
		return resolvedStudioIDsMsg{ids: ids}
	})
}

func (s *cmdServiceWithID) ResolveStudios(inputs []string) tea.Cmd {
	return s.withID(s.s.ResolveStudios(inputs))
}

func (s *cmdService) ResolvePerformers(inputs []string) tea.Cmd {
	return s.withLoadingCount(func() tea.Msg {
		ids, err := resolveEntityInputs(inputs, func(name string) (stash.Performer, error) {
			return s.PerformerFindByName(name)
		})
		if err != nil {
			return ErrorMsg{fmt.Errorf("performer resolution failed: %w", err)}
		}
		return resolvedPerformerIDsMsg{ids: ids}
	})
}

func (s *cmdServiceWithID) ResolvePerformers(inputs []string) tea.Cmd {
	return s.withID(s.s.ResolvePerformers(inputs))
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

func (s *cmdService) StudioFindByName(name string) (stash.Studio, error) {
	if studio, err := s.cache.GetStudioByName(name); err == nil {
		return studio, nil
	}
	if !s.cache.StudiosLoaded() {
		if _, err := s.Stash.StudiosAll(context.Background()); err != nil {
			return stash.Studio{}, err
		}
	}
	return s.cache.GetStudioByName(name)
}

func (s *cmdService) PerformerFindByName(name string) (stash.Performer, error) {
	if performer, err := s.cache.GetPerformerByName(name); err == nil {
		return performer, nil
	}
	if !s.cache.PerformersLoaded() {
		if _, err := s.Stash.PerformersAll(context.Background()); err != nil {
			return stash.Performer{}, err
		}
	}
	return s.cache.GetPerformerByName(name)
}
