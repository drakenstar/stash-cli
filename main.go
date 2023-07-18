package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/stash"
)

type Config struct {
	Endpoint     url.URL
	PathMappings map[string]string
}

func (c Config) MapPath(path string) string {
	for prefix, replacement := range c.PathMappings {
		if strings.HasPrefix(path, prefix) {
			return strings.Replace(path, prefix, replacement, 1)
		}
	}
	return path
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

	cfg.PathMappings = map[string]string{
		"/library": "/Volumes/Media/Library",
	}

	fmt.Printf("Connecting to %s\n", cfg.Endpoint.String())

	stsh := stash.New(graphql.NewClient(cfg.Endpoint.String(), http.DefaultClient))
	app := app.New(stsh, os.Stdin, os.Stdout, makeOpener(cfg))
	ctx := context.Background()

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

func makeOpener(c Config) app.Opener {
	return func(content any) error {
		switch cnt := content.(type) {
		case stash.Scene:
			fmt.Printf("open -a VLC %s\n", c.MapPath(cnt.File))
			return exec.Command("open", "-a", "VLC", c.MapPath(cnt.File)).Run()
		case stash.Gallery:
			fmt.Printf("open -a Xee³ %s\n", c.MapPath(cnt.File))
			return exec.Command("open", "-a", "Xee³", c.MapPath(cnt.File)).Run()
		}
		return fmt.Errorf("unsupported content type (%T)", content)
	}
}
