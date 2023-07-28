package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/term"
)

func main() {
	var cfg Config

	flag.BoolVar(&cfg.Debug, "debug", false, "enable HTTP debugging output")
	flag.Parse()

	stashInstance := flag.Arg(0)
	if stashInstance == "" {
		usage()
	}

	stashInstanceURL, err := url.Parse(stashInstance)
	if err != nil {
		fatal(err)
	}
	cfg.StashInstance = *stashInstanceURL

	cfg.PathMappings = map[string]string{
		"/library": "/Volumes/Media/Library",
	}
	cfg.OpenCommands.Scene = "open -a VLC"
	cfg.OpenCommands.Gallery = "open -a Xee³"

	fmt.Printf("Connecting to stash instance %s\n", cfg.GraphURL())

	var httpClient graphql.Doer = http.DefaultClient
	if cfg.Debug {
		httpClient = &loggingTransport{}
	}
	client := graphql.NewClient(cfg.GraphURL().String(), httpClient)
	stash := stash.New(client)
	opener := cfg.Opener(func(name string, args ...string) error {
		if cfg.Debug {
			fmt.Fprintln(os.Stderr, name, strings.Join(args, " "))
		}
		return exec.Command(name, args...).Run()
	})

	app := &app.App{
		In:  os.Stdin,
		Out: output{os.Stdout},
		States: map[string]app.AppState{
			"galleries": &app.GalleriesState{
				Stash:  stash,
				Opener: opener,
			},
			"scenes": &app.ScenesState{
				Stash:  stash,
				Opener: opener,
			},
		},
	}
	ctx := context.Background()

	fatalOnErr(app.Run(ctx, app.States["scenes"]))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: stash-cli [STASH INSTANCE]")
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, fmt.Errorf("fatal error: %w", err).Error())
	os.Exit(1)
}

func fatalOnErr(err error) {
	if err == nil {
		return
	}
	fatal(err)
}

type output struct {
	*os.File
}

func (o output) ScreenWidth() int {
	screenWidth, _, _ := term.GetSize(int(o.Fd()))
	return screenWidth
}
