package command

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testDest struct {
	foo       string
	Foo       string
	Bar       *string `actions:"baz"`
	Qux       int
	Quux      time.Time
	Quzz      *bool
	Corge     *bool
	Grault    bool
	Garply    string `actions:",positional"`
	Waldo     []string
	Fred      string `actions:",positional"`
	Plugh     customValue
	Remaining []string `actions:",positional"`
}

type customValue string

func (v *customValue) Set(s string) error {
	*v = customValue(fmt.Sprintf("<%s>", s))
	return nil
}

func ptr[T any](t T) *T {
	return &t
}

func TestBind(t *testing.T) {
	t.Run("should error non-pointer", func(t *testing.T) {
		dst := "string"
		a := Parser("")
		err := Bind(a, &dst)
		require.Equal(t, ErrNonPointerStruct, err)
	})
	t.Run("should error non-nil", func(t *testing.T) {
		var dst *testDest
		a := Parser("")
		err := Bind(a, dst)
		require.Equal(t, ErrNonPointerStruct, err)
	})
	t.Run("should error when additional positional arguments are given", func(t *testing.T) {
		var dst struct {
			Foo string `actions:",positional"`
		}
		a := Parser("foo bar")
		err := Bind(a, &dst)
		require.Equal(t, "foo", dst.Foo) // Validate that we did actually write the first argument.
		require.EqualError(t, err, "additional positional arguments given 'bar'")
	})
	t.Run("should error when unable to map argument", func(t *testing.T) {
		var dst struct {
			Foo string
		}
		a := Parser("foo=foo bar=bar")
		err := Bind(a, &dst)
		require.Equal(t, "foo", dst.Foo) // Validate that we did actually write the first argument.
		require.EqualError(t, err, "argument 'bar' does not map to bind destination")
	})
	t.Run("kitchen sink", func(t *testing.T) {
		var dst testDest
		a := Parser(`foo=bar baz=foo qux=99 quux=2025-09-12 corge grault=false "garply value" waldo=foo waldo=bar fred plugh=custom remaining1 remaining2`)
		err := Bind(a, &dst)
		require.NoError(t, err)

		require.Equal(t, testDest{
			foo:       "",
			Foo:       "bar",
			Bar:       ptr("foo"),
			Qux:       99,
			Quux:      time.Date(2025, 9, 12, 0, 0, 0, 0, time.UTC),
			Corge:     ptr(true),
			Grault:    false,
			Garply:    "garply value",
			Waldo:     []string{"foo", "bar"},
			Fred:      "fred",
			Plugh:     "<custom>",
			Remaining: []string{"remaining1", "remaining2"},
		}, dst)
	})
}
