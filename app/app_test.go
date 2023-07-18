package app

import (
	"bytes"
	"context"
	"testing"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/stretchr/testify/require"
)

type mockStash struct{}

func (mockStash) Stats(context.Context) (stash.Stats, error) {
	return stash.Stats{}, nil
}

func (mockStash) Scenes(context.Context, stash.FindFilter) ([]stash.Scene, int, error) {
	return []stash.Scene{
		{
			ID:    "1",
			Title: "Scene 1",
			File:  "/example/scene1.mp4",
		},
	}, 100, nil
}

func (mockStash) Galleries(context.Context, stash.FindFilter) ([]stash.Gallery, int, error) {
	return []stash.Gallery{
		{
			ID:    "1",
			Title: "Gallery 1",
			File:  "/example/gallery",
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

	require.Equal(t, "\nscenes (1/3) 1 Scene 1 /example/scene1.mp4\n\nscenes (1/3) 1 Gallery 1 /example/gallery\n\ngalleries (1/25) ", output.String())
}
