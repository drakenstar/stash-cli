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
			Date:  "2000-01-01",
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

type mockOutput struct {
	*bytes.Buffer
}

func (mockOutput) ScreenWidth() int {
	return 100
}

func TestApp(t *testing.T) {
	var buf bytes.Buffer
	output := &mockOutput{&buf}

	a := New(
		mockStash{},
		Renderer{Out: output},
		bytes.NewReader([]byte("scenes\ngalleries\nexit\n")),
		nil,
	)
	ctx := context.Background()

	a.Repl(ctx)

	require.Equal(t, "                                                                                                    \n○ 2000-01-01 Scene 1                                                                                \n🎬 (1/100) >>                                                                                                     \n○ 2000-01-01 Scene 1                                                                                \n🎬 (1/100) >> 1 Gallery 1 /example/gallery\n🏙 (1/1000) >> ", output.String())
}
