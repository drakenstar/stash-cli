package actions

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAction(t *testing.T) {
	a := New("command \ttest bar\t")

	token, err := a.Next()
	require.NoError(t, err)
	require.Equal(t, "command", token.Label)

	token, err = a.Next()
	require.NoError(t, err)
	require.Equal(t, "test", token.Label)

	token, err = a.Next()
	require.NoError(t, err)
	require.Equal(t, "bar", token.Label)

	token, err = a.Next()
	require.Equal(t, io.EOF, err)
}
