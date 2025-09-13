package command

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

type iter struct {
	args []Argument
	pos  int
}

func (i *iter) Next() (Argument, error) {
	if i.pos >= len(i.args) {
		return Argument{}, io.EOF
	}
	a := i.args[i.pos]
	i.pos++
	return a, nil
}

func newIter(s []string) *iter {
	i := &iter{}
	for _, val := range s {
		i.args = append(i.args, Argument{Raw: val, Value: val})
	}
	return i
}

type result struct {
	name string
	iter Iterator
}

func TestResolve(t *testing.T) {
	cfg := Config{
		"foo": {
			Resolve: func(i Iterator) (any, error) {
				return result{"foo", i}, nil
			},
			SubCommands: Config{
				"bar": {
					Resolve: func(i Iterator) (any, error) {
						return result{"bar", i}, nil
					},
				},
			},
		},
	}

	t.Run("should error on no input", func(t *testing.T) {
		_, err := cfg.Resolve(&iter{})
		require.EqualError(t, err, "no arguments returned")
	})

	t.Run("should error on non-name input", func(t *testing.T) {
		_, err := cfg.Resolve(&iter{args: []Argument{{Name: "foo"}}})
		require.EqualError(t, err, "invalid input: expected command")
	})

	t.Run("should resolve commands", func(t *testing.T) {
		v, err := cfg.Resolve(newIter([]string{"foo", "baz"}))
		require.NoError(t, err)
		require.Equal(t, "foo", v.(result).name)
		nxt, err := v.(result).iter.Next()
		require.NoError(t, err)
		require.Equal(t, "baz", nxt.Value)
	})

	t.Run("should resolve sub commands", func(t *testing.T) {
		v, err := cfg.Resolve(newIter([]string{"foo", "bar", "baz"}))
		require.NoError(t, err)
		require.Equal(t, "bar", v.(result).name)
		nxt, err := v.(result).iter.Next()
		require.NoError(t, err)
		require.Equal(t, "baz", nxt.Value)
	})
}
