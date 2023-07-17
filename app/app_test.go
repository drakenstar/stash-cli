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

func (mockStash) Scenes(context.Context, stash.FindFilter) ([]stash.Scene, error) {
	return []stash.Scene{
		{
			ID:    "1",
			Title: "Scene 1",
			File:  "/example/scene1.mp4",
		},
	}, nil
}

func (mockStash) Galleries(context.Context, stash.FindFilter) ([]stash.Gallery, error) {
	return []stash.Gallery{
		{
			ID:    "1",
			Title: "Gallery 1",
			File:  "/example/gallery",
		},
	}, nil
}

func TestApp(t *testing.T) {
	var output bytes.Buffer
	a := App{
		In:    bytes.NewReader([]byte("scenes\ngalleries\n\n")),
		Out:   &output,
		Stash: mockStash{},
	}
	ctx := context.Background()

	a.Repl(ctx)

	require.Equal(t, ">>> 1 Scene 1 /example/scene1.mp4\n>>> 1 Gallery 1 /example/gallery\n>>> ", output.String())
}
