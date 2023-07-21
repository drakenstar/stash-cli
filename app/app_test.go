package app

import (
	"bytes"
	"context"
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type mockStash struct{}

func (mockStash) Scenes(context.Context, stash.FindFilter) ([]stash.Scene, int, error) {
	return []stash.Scene{
		{
			ID:    "1",
			Title: "Scene 1",
			Files: []stash.VideoFile{{Path: "/example/scene1.mp4"}},
		},
	}, 100, nil
}

func (mockStash) Galleries(context.Context, stash.FindFilter) ([]stash.Gallery, int, error) {
	return []stash.Gallery{
		{
			ID:     "1",
			Title:  "Gallery 1",
			Folder: stash.File{Path: "/example/gallery"},
		},
	}, 1000, nil
}

func TestApp(t *testing.T) {
	var output bytes.Buffer
	a := New(
		mockStash{},
		&output,
		bytes.NewReader([]byte("scenes\ngalleries\nexit\n")),
		nil,
	)
	ctx := context.Background()

	a.Repl(ctx)

	require.Equal(t, "🎬 (1/100) 1 Scene 1 /example/scene1.mp4\n🎬 (1/100) 1 Gallery 1 /example/gallery\n🏙 (1/1000) ", output.String())
}
