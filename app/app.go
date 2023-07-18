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

	Out    io.Writer
	In     io.Reader
	Opener Opener

	*appState
}

type Opener func(content any) error

func New(s stash.Stash, out io.Writer, in io.Reader, opener Opener) *App {
	return &App{
		Stash:  s,
		Out:    out,
		In:     in,
		Opener: opener,

		appState: newAppState(),
	}
}

func (a *App) Repl(ctx context.Context) error {
	reader := bufio.NewReader(a.In)
	if err := a.query(ctx); err != nil {
		return err
	}
	for {
		page, count := a.PageAndCount()
		fmt.Fprint(a.Out, fmt.Sprintf("\n%s (%d/%d) ", a.mode, page, count))

		text, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		line := strings.TrimSpace(text)
		var command string
		tokens := strings.Split(line, " ")
		if len(tokens) > 0 {
			command = tokens[0]
		}

		switch command {
		case "open", "":
			if err := a.Opener(a.Current()); err != nil {
				return err
			}
			if a.Skip(1) {
				a.query(ctx)
			}
		case "scenes":
			a.SetMode(FilterModeScenes)
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, s := range a.scenesState.content {
				fmt.Fprintf(a.Out, "%s %s %s\n", s.ID, s.Title, s.File)
			}
		case "galleries":
			a.SetMode(FilterModeGalleries)
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, g := range a.galleriesState.content {
				fmt.Fprintf(a.Out, "%s %s %s\n", g.ID, g.Title, g.File)
			}
		case "filter":
			a.SetQuery(strings.Join(tokens[1:], " "))
			a.query(ctx)
			if a.mode == FilterModeScenes {
				for _, s := range a.scenesState.content {
					fmt.Fprintf(a.Out, "%s %s %s\n", s.ID, s.Title, s.File)
				}
			} else {
				for _, g := range a.galleriesState.content {
					fmt.Fprintf(a.Out, "%s %s %s\n", g.ID, g.Title, g.File)
				}
			}
		case "exit":
			return nil
		}
	}
}

func (a *App) query(ctx context.Context) (err error) {
	switch a.mode {
	case FilterModeScenes:
		a.scenesState.content, a.scenesState.total, err = a.Scenes(ctx, a.scenesState.filter)
	case FilterModeGalleries:
		a.galleriesState.content, a.galleriesState.total, err = a.Galleries(ctx, a.galleriesState.filter)
	default:
		panic("mode not set")
	}
	return err
}
