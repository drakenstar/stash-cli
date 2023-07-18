package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/drakenstar/stash-cli/stash"
)

type App struct {
	stash.Stash

	Out io.Writer
	In  io.Reader

	*appState
}

func New(s stash.Stash, out io.Writer, in io.Reader) *App {
	a := &App{
		Stash: s,
		Out:   out,
		In:    in,

		appState: newAppState(),
	}
	return a
}

func (a *App) Repl(ctx context.Context) error {
	reader := bufio.NewReader(a.In)
	if err := a.query(ctx); err != nil {
		return err
	}
	for {
		page, count := a.PageAndCount()
		fmt.Fprint(a.Out, fmt.Sprintf("%s (%d/%d) ", a.mode, page, count))

		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line := strings.TrimSpace(text)
		if line == "" {
			break
		}

		switch line {
		case "scenes":
			a.mode = FilterModeScenes
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, s := range a.scenesState.scenes {
				fmt.Fprintf(a.Out, "%s %s %s\n", s.ID, s.Title, s.File)
			}
		case "galleries":
			a.mode = FilterModeGalleries
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, g := range a.galleriesState.galleries {
				fmt.Fprintf(a.Out, "%s %s %s\n", g.ID, g.Title, g.File)
			}
		}
	}

	return nil
}

func (a *App) query(ctx context.Context) (err error) {
	fmt.Fprintf(a.Out, "loading %s", a.mode)
	switch a.mode {
	case FilterModeScenes:
		a.scenesState.scenes, a.scenesState.count, err = a.Scenes(ctx, a.sceneFindFilter)
	case FilterModeGalleries:
		a.galleriesState.galleries, a.scenesState.count, err = a.Galleries(ctx, a.sceneFindFilter)
	default:
		panic("mode not set")
	}
	return err
}
