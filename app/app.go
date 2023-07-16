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
			scenes, err := a.Scenes(ctx)
			if err != nil {
				return fmt.Errorf("scenes: %w", err)
			}
			for _, s := range scenes {
				fmt.Fprintf(a.Out, "%s %s %s\n", s.ID, s.Title, s.File)
			}
		case "galleries":
			galleries, err := a.Galleries(ctx)
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
