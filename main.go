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

type Config struct {
	Debug        bool
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
	var cfg Config

	flag.BoolVar(&cfg.Debug, "debug", false, "enable HTTP debugging output")
	flag.Parse()

	endpoint := flag.Arg(0)
	if endpoint == "" {
		usage()
	}

	endpointUrl, err := url.Parse(endpoint)
	if err != nil {
		fatal(err)
	}
	cfg.Endpoint = *endpointUrl

	cfg.PathMappings = map[string]string{
		"/library": "/Volumes/Media/Library",
	}

	fmt.Printf("Connecting to %s\n", cfg.Endpoint.String())

	var httpClient graphql.Doer = http.DefaultClient
	if cfg.Debug {
		httpClient = &loggingTransport{}
	}

	client := graphql.NewClient(cfg.Endpoint.String(), httpClient)
	app := app.New(stash.New(client), app.Renderer{Out: output{os.Stdin}}, os.Stdout, makeOpener(cfg))
	ctx := context.Background()

	fatalOnErr(app.Repl(ctx))
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: stash-cli ENDPOINT")
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
			if c.Debug {
				fmt.Printf("open -a VLC %s\n", c.MapPath(cnt.FilePath()))
			}
			return exec.Command("open", "-a", "VLC", c.MapPath(cnt.FilePath())).Run()
		case stash.Gallery:
			if c.Debug {
				fmt.Printf("open -a Xee³ %s\n", c.MapPath(cnt.FilePath()))
			}
			return exec.Command("open", "-a", "Xee³", c.MapPath(cnt.FilePath())).Run()
		}
		return fmt.Errorf("unsupported content type (%T)", content)
	}
}

type output struct {
	*os.File
}

func (o output) ScreenWidth() int {
	screenWidth, _, _ := term.GetSize(int(o.Fd()))
	return screenWidth
}
