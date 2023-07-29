package config

import (
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapPath(t *testing.T) {
	c := Config{
		PathMappings: map[string]string{
			"/old/": "/new/",
		},
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "path matches prefix",
			path:     "/old/path/to/file.txt",
			expected: "/new/path/to/file.txt",
		},
		{
			name:     "path doesn't match prefix",
			path:     "/another/path/to/file.txt",
			expected: "/another/path/to/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, c.MapPath(tt.path))
		})
	}
}

func TestURL(t *testing.T) {
	c := Config{
		StashInstance: mustParseURL(t, "http://example.com"),
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "no path",
			path:     "",
			expected: "http://example.com",
		},
		{
			name:     "with path",
			path:     "graphql",
			expected: "http://example.com/graphql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, c.URL(tt.path).String())
		})
	}
}

func TestOpen(t *testing.T) {
	c := Config{
		Debug:         true,
		StashInstance: mustParseURL(t, "http://example.com"),
	}

	t.Run("commands set", func(t *testing.T) {
		c.OpenCommands.URL = "open -a Safari {}"
		c.OpenCommands.Scene = "open {} -a VLC"
		c.OpenCommands.Gallery = "open -a Preview {}"

		tests := []struct {
			name        string
			content     interface{}
			expectedCmd []string
		}{
			{
				name:        "url",
				content:     "/path",
				expectedCmd: []string{"open", "-a", "Safari", "http://example.com/path"},
			},
			{
				name:        "scene",
				content:     stash.Scene{Files: []stash.VideoFile{{Path: "/path/to/file.mp4"}}},
				expectedCmd: []string{"open", "/path/to/file.mp4", "-a", "VLC"},
			},
			{
				name:        "gallery",
				content:     stash.Gallery{Folder: stash.Folder{Path: "/path/to/file.jpg"}},
				expectedCmd: []string{"open", "-a", "Preview", "/path/to/file.jpg"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				called := false
				execMock := func(name string, arg ...string) error {
					cmd := append([]string{name}, arg...)
					assert.Equal(t, tt.expectedCmd, cmd)
					called = true
					return nil
				}
				opener := c.Opener(execMock)
				err := opener(tt.content)
				assert.NoError(t, err)
				assert.True(t, called)
			})
		}
	})

	t.Run("defaults", func(t *testing.T) {
		c.OpenCommands.URL = ""
		c.OpenCommands.Scene = ""
		c.OpenCommands.Gallery = ""

		tests := []struct {
			name        string
			content     interface{}
			expectedCmd []string
		}{
			{
				name:        "url",
				content:     "/path",
				expectedCmd: []string{"open", "http://example.com/path"},
			},
			{
				name:        "scene",
				content:     stash.Scene{Files: []stash.VideoFile{{Path: "/path/to/file.mp4"}}},
				expectedCmd: []string{"open", "/path/to/file.mp4"},
			},
			{
				name:        "gallery",
				content:     stash.Gallery{Folder: stash.Folder{Path: "/path/to/file.jpg"}},
				expectedCmd: []string{"open", "/path/to/file.jpg"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				called := false
				execMock := func(name string, arg ...string) error {
					cmd := append([]string{name}, arg...)
					assert.Equal(t, tt.expectedCmd, cmd)
					called = true
					return nil
				}
				opener := c.Opener(execMock)
				err := opener(tt.content)
				assert.NoError(t, err)
				assert.True(t, called)
			})
		}
	})
}

func TestFromFuncs(t *testing.T) {
	t.Run("file all fields", func(t *testing.T) {
		c := &Config{
			PathMappings: map[string]string{
				"bar": "baz",
			},
		}
		f := io.NopCloser(strings.NewReader(`
			{
				"stashInstance": "http://example.com",
				"pathMappings": {
					"foo": "bar"
				},
				"openCommands": {
					"url": "url command",
					"scene": "scene command",
					"gallery": "gallery command"
				}
			}
		`))
		err := FromFile(c, f)
		require.NoError(t, err)
		require.Equal(t, Config{
			StashInstance: mustParseURL(t, "http://example.com"),
			PathMappings: map[string]string{
				"bar": "baz",
				"foo": "bar",
			},
			OpenCommands: OpenCommands{
				URL:     "url command",
				Scene:   "scene command",
				Gallery: "gallery command",
			},
		}, *c)
	})

	t.Run("noop", func(t *testing.T) {
		c := &Config{
			StashInstance: mustParseURL(t, "http://example.com/"),
			PathMappings: map[string]string{
				"bar": "baz",
			},
			OpenCommands: OpenCommands{
				URL:     "url command",
				Scene:   "scene command",
				Gallery: "gallery command",
			},
		}
		b := *c

		err := FromFile(c, io.NopCloser(strings.NewReader(`{"openCommands": {}}`)))
		require.NoError(t, err)
		err = FromArgs(c, []string{})
		require.NoError(t, err)

		require.Equal(t, b, *c)
	})

	t.Run("args all fields", func(t *testing.T) {
		c := &Config{
			PathMappings: map[string]string{
				"bar": "baz",
			},
		}
		FromArgs(c, []string{
			"--debug",
			"--stashInstance", "http://example.com",
			"--pathMapping", "bar:baz",
			"--pathMapping", "foo:bar",
			"--openCommandURL", "url command",
			"--openCommandScene", "scene command",
			"--openCommandGallery", "gallery command",
		})
		require.Equal(t, Config{
			Debug:         true,
			StashInstance: mustParseURL(t, "http://example.com"),
			PathMappings: map[string]string{
				"bar": "baz",
				"foo": "bar",
			},
			OpenCommands: OpenCommands{
				URL:     "url command",
				Scene:   "scene command",
				Gallery: "gallery command",
			},
		}, *c)
	})
}

func mustParseURL(t *testing.T, s string) *jsonURL {
	t.Helper()
	u, err := url.Parse(s)
	require.NoError(t, err)
	return &jsonURL{*u}
}
