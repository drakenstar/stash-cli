package stash

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
	"github.com/vektah/gqlparser/v2/validator"
)

type mockEndpoint struct {
	t        *testing.T
	response string
	called   bool
}

func (m *mockEndpoint) Do(r *http.Request) (*http.Response, error) {
	m.t.Helper()
	m.called = true

	var g struct {
		Query     string
		Variables map[string]any
	}
	body, err := io.ReadAll(r.Body)
	require.NoError(m.t, err)
	json.Unmarshal(body, &g)

	schema, err := validator.LoadSchema(
		validator.Prelude,
		&ast.Source{
			Name:  "schema.graphql",
			Input: schemaStr,
		},
	)
	require.NoError(m.t, err)

	doc, err := parser.ParseQuery(&ast.Source{
		Name:  "TestQuery",
		Input: g.Query,
	})
	require.NoError(m.t, err)

	validationErrs := validator.Validate(schema, doc)
	if len(validationErrs) > 0 {
		for _, validationErr := range validationErrs {
			fmt.Printf("Validation error: %v\n", validationErr)
		}
	}

	rw := httptest.NewRecorder()
	rw.WriteHeader(http.StatusOK)
	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(m.response))

	return rw.Result(), nil
}
