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

	appState
}

func New(s stash.Stash, out io.Writer, in io.Reader) *App {
	a := &App{
		Stash: s,
		Out:   out,
		In:    in,

		appState: appState{
			galleryFindFilter: stash.NewFindFilter(),
			sceneFindFilter:   stash.NewFindFilter(),
		},
	}
	return a
}

func (a *App) Repl(ctx context.Context) error {
	const prompt = ">>> "
	reader := bufio.NewReader(a.In)
	for {
		fmt.Fprint(a.Out, prompt)
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
			scenes, err := a.Scenes(ctx, a.sceneFindFilter)
			if err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, s := range scenes {
				fmt.Fprintf(a.Out, "%s %s %s\n", s.ID, s.Title, s.File)
			}
		case "galleries":
			galleries, err := a.Galleries(ctx, a.galleryFindFilter)
			if err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, g := range galleries {
				fmt.Fprintf(a.Out, "%s %s %s\n", g.ID, g.Title, g.File)
			}
		}
	}

	return nil
}
