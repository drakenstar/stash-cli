package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/drakenstar/stash-cli/stash"
)

type App struct {
	stash.Stash

	In       io.Reader
	Renderer Renderer
	Opener   Opener

	*appState
}

type Opener func(content any) error

func New(s stash.Stash, renderer Renderer, in io.Reader, opener Opener) *App {
	return &App{
		Stash:    s,
		Renderer: renderer,
		In:       in,
		Opener:   opener,

		appState: newAppState(),
	}
}

var promptStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#7D56F4"))

func (a *App) Repl(ctx context.Context) error {
	reader := bufio.NewReader(a.In)
	if err := a.query(ctx); err != nil {
		return err
	}
	a.Renderer.ContentList(a)

	for {
		a.Renderer.Prompt(a)

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
		case "":
			if a.Opened() {
				if a.Skip(1) {
					a.query(ctx)
				}
			}
			a.Renderer.ContentRow(a)
			if err := a.Opener(a.Current()); err != nil {
				return err
			}
		case "open":
			a.Renderer.ContentRow(a)
			if err := a.Opener(a.Current()); err != nil {
				return err
			}
		case "scenes", "s":
			a.SetMode(FilterModeScenes)
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			a.Renderer.ContentList(a)
		case "galleries", "g":
			a.SetMode(FilterModeGalleries)
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("galleries: %w", err)
			}
			a.Renderer.ContentList(a)
		case "filter", "f":
			a.SetQuery(strings.Join(tokens[1:], " "))
			a.query(ctx)
			a.Renderer.ContentList(a)
		case "random", "r":
			a.SetSort(stash.RandomSort())
			a.query(ctx)
			a.Renderer.ContentList(a)
		case "list":
			a.Renderer.ContentList(a)
		case "reset":
			a.SetMode(a.mode)
			if err := a.query(ctx); err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			a.Renderer.ContentList(a)
		case "refresh":
			if err := a.query(ctx); err != nil {
				return err
			}
			a.Renderer.ContentList(a)
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
