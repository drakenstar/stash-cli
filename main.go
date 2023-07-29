package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/hasura/go-graphql-client"
	"golang.org/x/term"
)

const defaultConfigFile = ".stash-cli.json"

func main() {
	cfg := loadConfig()

	if cfg.Debug {
		fmt.Printf("Connecting to stash instance %s\n", cfg.GraphURL())
	}

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

func loadConfig() *config.Config {
	var c config.Config

	home, err := os.UserHomeDir()
	fatalOnErr(err)
	configPath := filepath.Join(home, defaultConfigFile)
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("Loading configuration from '%s'\n", configPath)
		f, err := os.Open(configPath)
		fatalOnErr(err)
		config.FromFile(&c, f)
	}
	config.FromArgs(&c, os.Args[1:])
	if c.StashInstance == nil {
		usage()
	}

	return &c
}
