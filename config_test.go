package main

import (
	"net/url"
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/assert"
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
	u, _ := url.Parse("http://example.com")
	c := Config{
		StashInstance: *u,
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
	instance, _ := url.Parse("http://example.com")
	c := Config{
		Debug:         true,
		StashInstance: *instance,
		OpenCommands: struct {
			URL     string `json:"url"`
			Scene   string `json:"scene"`
			Gallery string `json:"gallery"`
		}{
			URL:     "open -a Safari {}",
			Scene:   "open -a VLC {}",
			Gallery: "open -a Preview {}",
		},
	}

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
			expectedCmd: []string{"open", "-a", "VLC", "/path/to/file.mp4"},
		},
		{
			name:        "gallery",
			content:     stash.Gallery{Folder: stash.Folder{Path: "/path/to/file.jpg"}},
			expectedCmd: []string{"open", "-a", "Preview", "/path/to/file.jpg"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execMock := func(name string, arg ...string) error {
				cmd := append([]string{name}, arg...)
				assert.Equal(t, tt.expectedCmd, cmd)
				return nil
			}
			opener := c.Opener(execMock)
			err := opener(tt.content)
			assert.NoError(t, err)
		})
	}
}
