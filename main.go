package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Khan/genqlient/graphql"
	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/stash"
)

type Config struct {
	Endpoint url.URL
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	var cfg Config

	endpoint, err := url.Parse(os.Args[1])
	if err != nil {
		fatal(err)
	}
	cfg.Endpoint = *endpoint

	fmt.Printf("Connecting to %s\n", cfg.Endpoint.String())

	stsh := stash.New(graphql.NewClient(cfg.Endpoint.String(), http.DefaultClient))
	app := app.New(stsh, os.Stdin, os.Stdout)
	ctx := context.Background()

	stats, err := stsh.Stats(ctx)
	fatalOnErr(err)
	fmt.Printf("\tgalleries: %d\n\tscenes: %d\n\tperformers: %d\n", stats.SceneCount, stats.GalleryCount, stats.PerformerCount)

	fatalOnErr(app.Repl(ctx))
}

func usage() {
	fmt.Fprintln(os.Stderr, "stash-cli ENDPOINT")
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
