package stash

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/require"
)

func TestTagCreate(t *testing.T) {
	doer := &captureEndpoint{
		t:        t,
		response: `{"data": {"tagCreate": {"id": "1", "name": "Foo"}}}`,
	}
	client := graphql.NewClient("https://example.com/graph", doer)
	s := stash{client}

	tag, err := s.TagCreate(context.Background(), TagCreate{Name: "Foo"})

	require.NoError(t, err)
	require.Equal(t, Tag{ID: "1", Name: "Foo"}, tag)
	require.True(t, doer.called)
	require.Contains(t, doer.body, `"query":"mutation`)
	require.Contains(t, doer.body, `tagCreate(input: {name: $name})`)
	require.Contains(t, doer.body, `"name":"Foo"`)
}

func TestTagFindByNameRequest(t *testing.T) {
	doer := &captureEndpoint{
		t:        t,
		response: `{"data": {"findTags": {"count": 0, "tags": []}}}`,
	}
	client := graphql.NewClient("https://example.com/graph", doer)
	s := stash{client}

	_, err := s.TagFindByName(context.Background(), "Foo")

	require.ErrorIs(t, err, ErrTagNotFound)
	require.Contains(t, doer.body, `findTags`)
	require.Contains(t, doer.body, `"tag_filter"`)
	require.NotContains(t, doer.body, `"filter"`)
}

type captureEndpoint struct {
	t        *testing.T
	response string
	body     string
	called   bool
}

func (m *captureEndpoint) Do(r *http.Request) (*http.Response, error) {
	m.t.Helper()
	m.called = true
	body, err := io.ReadAll(r.Body)
	require.NoError(m.t, err)
	m.body = string(body)
	rw := httptest.NewRecorder()
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(m.response))
	return rw.Result(), nil
}
