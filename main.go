package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/term"
)

type Config struct {
	Debug         bool
	StashInstance url.URL
	PathMappings  map[string]string
}

func (c Config) MapPath(path string) string {
	for prefix, replacement := range c.PathMappings {
		if strings.HasPrefix(path, prefix) {
			return strings.Replace(path, prefix, replacement, 1)
		}
	}
	return path
}

func (c Config) URL(p string) *url.URL {
	u := c.StashInstance
	u.Path = path.Join(u.Path, p)
	return &u
}

func (c Config) GraphURL() *url.URL {
	return c.URL("graphql")
}

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

	fmt.Printf("Connecting to stash instance %s\n", cfg.GraphURL())

	var httpClient graphql.Doer = http.DefaultClient
	if cfg.Debug {
		httpClient = &loggingTransport{}
	}

	client := graphql.NewClient(cfg.GraphURL().String(), httpClient)
	app := &app.App{
		Stash:  stash.New(client),
		In:     os.Stdin,
		Out:    output{os.Stdout},
		Opener: makeOpener(cfg),
	}
	ctx := context.Background()

	fatalOnErr(app.Repl(ctx))
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

func makeOpener(c Config) app.Opener {
	return func(content any) error {
		switch cnt := content.(type) {
		case string:
			u := c.URL(cnt)
			if c.Debug {
				fmt.Printf("open %s\n", u.String())
			}
			return exec.Command("open", u.String()).Run()
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
