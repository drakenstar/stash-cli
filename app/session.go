package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/stash"
)

const sessionVersion = 1

type SessionStore interface {
	Load() (Session, bool, error)
	Save(Session) error
	Delete() error
}

type FileSessionStore struct {
	Path string
}

func NewFileSessionStore(path string) FileSessionStore {
	return FileSessionStore{Path: path}
}

func (s FileSessionStore) Load() (Session, bool, error) {
	if s.Path == "" {
		return Session{}, false, nil
	}
	f, err := os.Open(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	defer f.Close()

	var session Session
	if err := json.NewDecoder(f).Decode(&session); err != nil {
		return Session{}, false, err
	}
	return session, true, nil
}

func (s FileSessionStore) Save(session Session) error {
	if s.Path == "" {
		return nil
	}
	dir := filepath.Dir(s.Path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".session-*.json")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(session); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, s.Path)
}

func (s FileSessionStore) Delete() error {
	if s.Path == "" {
		return nil
	}
	if err := os.Remove(s.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

type Session struct {
	Version        int          `json:"version"`
	StashInstance  string       `json:"stashInstance"`
	ActiveTab      int          `json:"activeTab"`
	CommandHistory []string     `json:"commandHistory,omitempty"`
	Tabs           []TabSession `json:"tabs"`
}

type TabSession struct {
	Type      string            `json:"type"`
	Scenes    *ScenesSession    `json:"scenes,omitempty"`
	Galleries *GalleriesSession `json:"galleries,omitempty"`
}

type ScenesSession struct {
	Query         string                     `json:"query,omitempty"`
	Sort          string                     `json:"sort,omitempty"`
	SortDirection string                     `json:"sortDirection,omitempty"`
	Filter        stash.SceneFilter          `json:"filter"`
	Page          PageSession                `json:"page"`
	History       []ScenesFilterStateSession `json:"history,omitempty"`
}

type ScenesFilterStateSession struct {
	Query         string            `json:"query,omitempty"`
	Sort          string            `json:"sort,omitempty"`
	SortDirection string            `json:"sortDirection,omitempty"`
	Filter        stash.SceneFilter `json:"filter"`
	Page          PageSession       `json:"page"`
}

type GalleriesSession struct {
	Query         string                        `json:"query,omitempty"`
	Sort          string                        `json:"sort,omitempty"`
	SortDirection string                        `json:"sortDirection,omitempty"`
	Filter        stash.GalleryFilter           `json:"filter"`
	Page          PageSession                   `json:"page"`
	History       []GalleriesFilterStateSession `json:"history,omitempty"`
}

type GalleriesFilterStateSession struct {
	Query         string              `json:"query,omitempty"`
	Sort          string              `json:"sort,omitempty"`
	SortDirection string              `json:"sortDirection,omitempty"`
	Filter        stash.GalleryFilter `json:"filter"`
	Page          PageSession         `json:"page"`
}

type PageSession struct {
	Position int  `json:"position"`
	Opened   bool `json:"opened"`
}

func (m *Model) SetSessionStore(store SessionStore, stashInstance string) {
	m.sessionStore = store
	m.sessionStashInstance = stashInstance
}

func (m Model) SaveSession() Session {
	session := Session{
		Version:        sessionVersion,
		StashInstance:  m.sessionStashInstance,
		ActiveTab:      m.active,
		CommandHistory: m.commandInput.History(),
		Tabs:           make([]TabSession, 0, len(m.tabs)),
	}
	for _, tab := range m.tabs {
		switch model := tab.model.(type) {
		case *ScenesModel:
			saved := model.saveSession()
			session.Tabs = append(session.Tabs, TabSession{Type: "scenes", Scenes: &saved})
		case *GalleriesModel:
			saved := model.saveSession()
			session.Tabs = append(session.Tabs, TabSession{Type: "galleries", Galleries: &saved})
		}
	}
	if session.ActiveTab >= len(session.Tabs) {
		session.ActiveTab = max(len(session.Tabs)-1, 0)
	}
	return session
}

func (m *Model) RestoreSession(session Session) bool {
	if session.Version != sessionVersion || session.StashInstance != m.sessionStashInstance || len(session.Tabs) == 0 {
		return false
	}

	m.tabs = nil
	m.tabsByID = make(map[tabID]tab)
	m.active = 0
	m.commandInput.SetHistory(session.CommandHistory)

	for _, saved := range session.Tabs {
		newFunc, ok := m.tabFuncs[saved.Type]
		if !ok {
			continue
		}
		id := m.nextTabID()
		model := newFunc(id)
		switch typed := model.(type) {
		case *ScenesModel:
			if saved.Scenes != nil {
				typed.restoreSession(*saved.Scenes)
			}
		case *GalleriesModel:
			if saved.Galleries != nil {
				typed.restoreSession(*saved.Galleries)
			}
		}
		t := tab{id: id, model: model}
		m.tabs = append(m.tabs, t)
		m.tabsByID[id] = t
	}
	if len(m.tabs) == 0 {
		m.resetTabs()
		return false
	}
	m.active = clampInt(session.ActiveTab, 0, len(m.tabs)-1)
	return true
}

func (m *Model) resetSession() tea.Cmd {
	m.mode = ModeNormal
	m.confirmation = nil
	m.pendingDelete = nil
	m.err = nil
	m.tagsLoading = false
	m.studiosLoading = false
	m.performersLoading = false
	m.commandInput.SetHistory(nil)
	m.resetTabs()
	if m.sessionStore != nil {
		if err := m.sessionStore.Delete(); err != nil {
			return NewErrorCmd(err)
		}
	}
	return tea.Batch(
		m.tabs[m.active].model.Init(),
		m.tabs[m.active].model.SetSize(Size{Width: m.screen.Width, Height: m.screen.Height - 5}),
	)
}

func (m *Model) resetTabs() {
	m.tabsByID = make(map[tabID]tab)
	m.tabs = nil
	m.active = 0
	newFunc := m.tabFuncs["scenes"]
	id := m.nextTabID()
	m.tabs = []tab{{id: id, model: newFunc(id)}}
	m.tabsByID[id] = m.tabs[0]
}

func (m Model) quitCmd() tea.Cmd {
	return func() tea.Msg {
		if m.sessionStore != nil {
			if err := m.sessionStore.Save(m.SaveSession()); err != nil {
				return ErrorMsg{err}
			}
		}
		return tea.QuitMsg{}
	}
}

func (m *ScenesModel) saveSession() ScenesSession {
	history := make([]ScenesFilterStateSession, 0, len(m.history))
	for _, state := range m.history {
		history = append(history, ScenesFilterStateSession{
			Query:         state.query,
			Sort:          state.sort,
			SortDirection: state.sortDirection,
			Filter:        state.sceneFilter,
			Page:          savePageSession(state.pageState),
		})
	}
	return ScenesSession{
		Query:         m.query,
		Sort:          m.sort,
		SortDirection: m.sortDirection,
		Filter:        m.sceneFilter,
		Page:          savePageSession(m.pageState),
		History:       history,
	}
}

func (m *ScenesModel) restoreSession(session ScenesSession) {
	m.query = session.Query
	m.sort = session.Sort
	if m.sort == "" {
		m.sort = stash.SortDate
	}
	m.sortDirection = session.SortDirection
	if m.sortDirection == "" {
		m.sortDirection = stash.SortDirectionDesc
	}
	m.sceneFilter = session.Filter
	m.pageState = restorePageSession(session.Page, m.pageState.PerPage)
	m.scenes = nil
	m.history = make([]sceneFilterState, 0, len(session.History))
	for _, state := range session.History {
		m.history = append(m.history, sceneFilterState{
			query:         state.Query,
			sort:          state.Sort,
			sortDirection: state.SortDirection,
			sceneFilter:   state.Filter,
			pageState:     restorePageSession(state.Page, m.pageState.PerPage),
		})
	}
}

func (m *GalleriesModel) saveSession() GalleriesSession {
	history := make([]GalleriesFilterStateSession, 0, len(m.history))
	for _, state := range m.history {
		history = append(history, GalleriesFilterStateSession{
			Query:         state.query,
			Sort:          state.sort,
			SortDirection: state.sortDirection,
			Filter:        state.galleryFilter,
			Page:          savePageSession(state.pageState),
		})
	}
	return GalleriesSession{
		Query:         m.query,
		Sort:          m.sort,
		SortDirection: m.sortDirection,
		Filter:        m.galleryFilter,
		Page:          savePageSession(m.pageState),
		History:       history,
	}
}

func (m *GalleriesModel) restoreSession(session GalleriesSession) {
	m.query = session.Query
	m.sort = session.Sort
	if m.sort == "" {
		m.sort = stash.SortPath
	}
	m.sortDirection = session.SortDirection
	if m.sortDirection == "" {
		m.sortDirection = stash.SortDirectionAsc
	}
	m.galleryFilter = session.Filter
	m.pageState = restorePageSession(session.Page, m.pageState.PerPage)
	m.galleries = nil
	m.history = make([]galleryFilterState, 0, len(session.History))
	for _, state := range session.History {
		m.history = append(m.history, galleryFilterState{
			query:         state.Query,
			sort:          state.Sort,
			sortDirection: state.SortDirection,
			galleryFilter: state.Filter,
			pageState:     restorePageSession(state.Page, m.pageState.PerPage),
		})
	}
}

func savePageSession(page pageState) PageSession {
	return PageSession{Position: page.Position(), Opened: page.opened}
}

func restorePageSession(saved PageSession, perPage int) pageState {
	page := pageState{PerPage: perPage, opened: saved.Opened}
	if perPage <= 0 || saved.Position <= 0 {
		return page
	}
	page.page = saved.Position / perPage
	page.index = saved.Position % perPage
	return page
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
