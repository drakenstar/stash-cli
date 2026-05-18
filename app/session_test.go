package app

import (
	"errors"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
	"github.com/stretchr/testify/require"
)

type memorySessionStore struct {
	session Session
	saved   bool
	deleted bool
	err     error
}

func (s *memorySessionStore) Load() (Session, bool, error) {
	return s.session, s.saved, s.err
}

func (s *memorySessionStore) Save(session Session) error {
	if s.err != nil {
		return s.err
	}
	s.session = session
	s.saved = true
	return nil
}

func (s *memorySessionStore) Delete() error {
	if s.err != nil {
		return s.err
	}
	s.deleted = true
	s.saved = false
	s.session = Session{}
	return nil
}

func TestFileSessionStoreRoundTrip(t *testing.T) {
	store := NewFileSessionStore(filepath.Join(t.TempDir(), "session.json"))

	_, ok, err := store.Load()
	require.NoError(t, err)
	require.False(t, ok)

	expected := Session{
		Version:        sessionVersion,
		StashInstance:  "http://example.com",
		ActiveTab:      1,
		CommandHistory: []string{"filter tag=foo"},
		Tabs: []TabSession{
			{Type: "scenes", Scenes: &ScenesSession{Query: "foo"}},
		},
	}
	require.NoError(t, store.Save(expected))

	actual, ok, err := store.Load()
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, expected, actual)

	require.NoError(t, store.Delete())
	_, ok, err = store.Load()
	require.NoError(t, err)
	require.False(t, ok)
}

func TestSessionRoundTrip(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(&memorySessionStore{}, "file:///tmp/stash")
	m.commandInput.SetHistory([]string{"filter tag=foo", "sort random"})

	scenes := m.tabs[0].model.(*ScenesModel)
	scenes.query = "query"
	scenes.sort = stash.SortCreatedAt
	scenes.sortDirection = stash.SortDirectionAsc
	scenes.pageState = pageState{PerPage: 5, page: 2, index: 3, opened: true}
	scenes.sceneFilter.Tags = &stash.HierarchicalMultiCriterion{
		Value:    []string{"tag-1"},
		Modifier: stash.CriterionModifierIncludes,
	}
	scenes.history = []sceneFilterState{{
		query:         "old",
		sort:          stash.SortDate,
		sortDirection: stash.SortDirectionDesc,
		pageState:     pageState{PerPage: 5, page: 1, index: 1},
	}}

	m.TabOpen(m.tabFuncs["galleries"])
	galleries := m.tabs[1].model.(*GalleriesModel)
	galleries.query = "gallery"
	galleries.pageState = pageState{PerPage: 40, page: 1, index: 2, opened: true}
	m.active = 1

	session := m.SaveSession()
	restored := New(&stash.LocalStash{}, nil)
	restored.SetSessionStore(&memorySessionStore{}, "file:///tmp/stash")
	require.True(t, restored.RestoreSession(session))

	require.Equal(t, 1, restored.active)
	require.Equal(t, []string{"filter tag=foo", "sort random"}, restored.commandInput.History())
	require.Len(t, restored.tabs, 2)

	restoredScenes := restored.tabs[0].model.(*ScenesModel)
	require.Equal(t, "query", restoredScenes.query)
	require.Equal(t, stash.SortCreatedAt, restoredScenes.sort)
	require.Equal(t, stash.SortDirectionAsc, restoredScenes.sortDirection)
	require.Equal(t, 13, restoredScenes.pageState.Position())
	require.True(t, restoredScenes.pageState.opened)
	require.NotNil(t, restoredScenes.sceneFilter.Tags)
	require.Equal(t, []string{"tag-1"}, restoredScenes.sceneFilter.Tags.Value)
	require.Len(t, restoredScenes.history, 1)

	restoredGalleries := restored.tabs[1].model.(*GalleriesModel)
	require.Equal(t, "gallery", restoredGalleries.query)
	require.Equal(t, 42, restoredGalleries.pageState.Position())
}

func TestRestoredSessionPositionSurvivesResize(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(&memorySessionStore{}, "file:///tmp/stash")
	require.True(t, m.RestoreSession(Session{
		Version:       sessionVersion,
		StashInstance: "file:///tmp/stash",
		Tabs: []TabSession{{
			Type:   "scenes",
			Scenes: &ScenesSession{Page: PageSession{Position: 13, Opened: true}},
		}},
	}))

	scenes := m.tabs[0].model.(*ScenesModel)
	require.Equal(t, 13, scenes.pageState.Position())

	_ = scenes.SetSize(Size{Width: 100, Height: 25})

	require.Equal(t, 13, scenes.pageState.Position())
	require.Equal(t, 24, scenes.pageState.PerPage)
	require.Equal(t, 0, scenes.pageState.page)
	require.Equal(t, 13, scenes.pageState.index)
	require.True(t, scenes.pageState.opened)
}

func TestRestoreSessionRejectsMismatchedInstance(t *testing.T) {
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(&memorySessionStore{}, "file:///current")

	restored := m.RestoreSession(Session{
		Version:       sessionVersion,
		StashInstance: "file:///other",
		Tabs:          []TabSession{{Type: "scenes", Scenes: &ScenesSession{}}},
	})

	require.False(t, restored)
	require.Len(t, m.tabs, 1)
}

func TestQuitSavesSession(t *testing.T) {
	store := &memorySessionStore{}
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(store, "file:///tmp/stash")
	m.commandInput.SetHistory([]string{"filter tag=foo"})

	msg := m.quitCmd()()

	_, ok := msg.(tea.QuitMsg)
	require.True(t, ok)
	require.True(t, store.saved)
	require.Equal(t, []string{"filter tag=foo"}, store.session.CommandHistory)
}

func TestExitCommandSavesSession(t *testing.T) {
	store := &memorySessionStore{}
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(store, "file:///tmp/stash")

	updated, cmd := m.Update(ui.CommandExecMsg{Command: "exit"})
	require.NotNil(t, cmd)

	updated, cmd = updated.Update(cmd())
	require.NotNil(t, cmd)

	_, ok := cmd().(tea.QuitMsg)
	require.True(t, ok)
	require.True(t, store.saved)
	require.IsType(t, Model{}, updated)
}

func TestQuitReportsSaveError(t *testing.T) {
	store := &memorySessionStore{err: errors.New("save failed")}
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(store, "file:///tmp/stash")

	msg := m.quitCmd()()

	errMsg, ok := msg.(ErrorMsg)
	require.True(t, ok)
	require.EqualError(t, errMsg.error, "save failed")
}

func TestSessionNewResetsAndDeletesSession(t *testing.T) {
	store := &memorySessionStore{saved: true}
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(store, "file:///tmp/stash")
	m.commandInput.SetHistory([]string{"filter tag=foo"})
	m.TabOpen(m.tabFuncs["galleries"])
	m.active = 1
	m.err = errors.New("boom")

	cmd := m.resetSession()

	require.True(t, store.deleted)
	require.Len(t, m.tabs, 1)
	require.Equal(t, 0, m.active)
	require.Empty(t, m.commandInput.History())
	require.Nil(t, m.err)
	require.NotNil(t, cmd)
}

func TestSessionNewCommandResetsAndDeletesSession(t *testing.T) {
	store := &memorySessionStore{saved: true}
	m := New(&stash.LocalStash{}, nil)
	m.SetSessionStore(store, "file:///tmp/stash")
	m.commandInput.SetHistory([]string{"filter tag=foo"})
	m.TabOpen(m.tabFuncs["galleries"])
	m.active = 1

	updated, cmd := m.Update(ui.CommandExecMsg{Command: "session new"})
	require.NotNil(t, cmd)

	updated, cmd = updated.Update(cmd())
	require.NotNil(t, cmd)

	model := updated.(Model)
	require.True(t, store.deleted)
	require.Len(t, model.tabs, 1)
	require.Equal(t, 0, model.active)
	require.Empty(t, model.commandInput.History())
}
