package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/drakenstar/stash-cli/ui"
)

type App struct {
	stash.Stash

	In     io.Reader
	Out    Output
	Opener Opener
}

type Opener func(content any) error

type Output interface {
	io.Writer
	ScreenWidth() int
}

// Input represents a complete line of input from the user
type Input string

func NewInput(s string) Input {
	return Input(strings.TrimSpace(s))
}

// Command returns all characters up to the first encountered space in an input string.  This is to be interpretted
// as the command for the rest of the input.
func (i Input) Command() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return string(i)
	}
	return string(i[:idx])
}

// Returns all text after the initial command.  This may be interpretted in any way an action deems appropriate.
func (i Input) ArgString() string {
	idx := strings.Index(string(i), " ")
	if idx == -1 {
		return ""
	}
	return string(i[idx+1:])
}

type AppState interface {
	Init(context.Context) error
	Update(context.Context, Input) error
	View() string
}

func (a *App) Repl(ctx context.Context) error {
	reader := bufio.NewReader(a.In)

	galleries := &galleriesState{App: a}
	scenes := &scenesState{App: a}
	var current AppState = scenes

	if err := current.Init(ctx); err != nil {
		return err
	}

	for {
		fmt.Fprintf(a.Out, current.View())
		fmt.Fprintf(a.Out, ui.Prompt())

		line, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		in := NewInput(line)

		switch in.Command() {
		case "scenes", "s":
			current = scenes
			if err := current.Init(ctx); err != nil {
				return err
			}
			continue
		case "galleries", "g":
			current = galleries
			if err := current.Init(ctx); err != nil {
				return err
			}
			continue
		case "exit":
			return nil
		}

		if err := current.Update(ctx, in); err != nil {
			return err
		}
	}
}
