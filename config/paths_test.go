package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultPaths(t *testing.T) {
	paths, err := DefaultPaths()
	require.NoError(t, err)

	require.Equal(t, ConfigFile, filepath.Base(paths.ConfigPath))
	require.Equal(t, AppName, filepath.Base(filepath.Dir(paths.ConfigPath)))
	require.Equal(t, SessionFile, filepath.Base(paths.SessionPath))
	require.Equal(t, AppName, filepath.Base(filepath.Dir(paths.SessionPath)))
}

func TestConfigPathExists(t *testing.T) {
	dir := t.TempDir()
	paths := Paths{
		ConfigPath: filepath.Join(dir, AppName, ConfigFile),
	}

	ok, err := ConfigPathExists(paths)
	require.NoError(t, err)
	require.False(t, ok)

	require.NoError(t, os.MkdirAll(filepath.Dir(paths.ConfigPath), 0o755))
	require.NoError(t, os.WriteFile(paths.ConfigPath, []byte(`{}`), 0o644))
	ok, err = ConfigPathExists(paths)
	require.NoError(t, err)
	require.True(t, ok)
}
