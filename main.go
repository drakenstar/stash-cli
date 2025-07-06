package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/config"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/hasura/go-graphql-client"
)

const defaultConfigFile = ".stash-cli.json"

func main() {
	cfg := loadConfig()

	if cfg.Debug {
		fmt.Printf("Connecting to stash instance %s\n", cfg.GraphURL())
	}

	httpClient := &client{
		Client: http.DefaultClient,
		APIKey: cfg.APIKey,
		Debug:  cfg.Debug,
	}

	client := graphql.NewClient(cfg.GraphURL().String(), httpClient)
	stash := stash.New(client)
	opener := cfg.Opener(func(name string, args ...string) error {
		if cfg.Debug {
			fmt.Fprintln(os.Stderr, name, strings.Join(args, " "))
		}
		return exec.Command(name, args...).Run()
	})

	app := app.New([]app.TabModelConfig{
		{
			NewFunc: func() app.TabModel { return app.NewScenesModel(stash, opener) },
			Name:    "scenes",
			KeyBind: "s",
		},
		{
			NewFunc: func() app.TabModel { return app.NewGalleriesModel(stash, opener) },
			Name:    "galleries",
			KeyBind: "g",
		},
	})

	p := tea.NewProgram(
		app,
		// tea.WithAltScreen(), TODO buggy with emoji atm
	)
	if _, err := p.Run(); err != nil {
		fatal(err)
	}
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

// client that implements the required API Key authentication and debugging capabilities.
type client struct {
	*http.Client
	APIKey string
	Debug  bool
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("ApiKey", c.APIKey)

	if c.Debug {
		bytes, _ := httputil.DumpRequestOut(req, true)
		resp, err := http.DefaultTransport.RoundTrip(req)
		respBytes, _ := httputil.DumpResponse(resp, true)
		bytes = append(bytes, respBytes...)
		fmt.Printf("%s\n", bytes)
		return resp, err
	}

	return c.Client.Do(req)
}
