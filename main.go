package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/machinebox/graphql"
)

type Config struct {
	Endpoint url.URL
}

type App struct {
	*graphql.Client
}

func (a *App) Stats(ctx context.Context) {
	req := graphql.NewRequest(`
		query {
			stats {
				scene_count
				scenes_size
				gallery_count
				performer_count
			}
		}
	`)

	var resp struct {
		Stats struct {
			SceneCount     int `json:"scene_count"`
			GalleryCount   int `json:"gallery_count"`
			PerformerCount int `json:"performer_count"`
		}
	}

	err := a.Run(ctx, req, &resp)
	if err != nil {
		fatal(err)
	}

	fmt.Printf(
		"\tscenes: %d\n\tgalleries: %d\n\tperformers: %d\n",
		resp.Stats.SceneCount,
		resp.Stats.GalleryCount,
		resp.Stats.PerformerCount,
	)
}

func (a *App) Repl(ctx context.Context) {
	const prompt = ">>> "
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		text, err := reader.ReadString('\n')
		if err != nil {
			fatal(err)
		}
		line := strings.TrimSpace(text)
		if line == "" {
			break
		}

		fmt.Println(line)

		switch line {
		case "scenes":
			req := graphql.NewRequest(`
				query {
					findScenes {
						count
						scenes {
							id
							title
							files {
								path
							}
						}
					}
				}
			`)

			var resp struct {
				FindScenes struct {
					Count  int
					Scenes []struct {
						ID    string
						Title string
						Files []struct {
							Path string
						}
					}
				}
			}
			err = a.Run(ctx, req, &resp)
			if err != nil {
				fatal(err)
			}

			for _, g := range resp.FindScenes.Scenes {
				fmt.Printf("%s %s %s\n", g.ID, g.Title, g.Files[0].Path)
			}

		case "galleries":
			req := graphql.NewRequest(`
				query {
					findGalleries {
						count
						galleries {
							id
							title
							files {
								path
							}
						}
					}
				}
			`)

			var resp struct {
				FindGalleries struct {
					Count     int
					Galleries []struct {
						ID    string
						Title string
						Files []struct {
							Path string
						}
					}
				}
			}
			err = a.Run(ctx, req, &resp)
			if err != nil {
				fatal(err)
			}

			for _, g := range resp.FindGalleries.Galleries {
				fmt.Printf("%s %s %s\n", g.ID, g.Title, g.Files[0].Path)
			}
		}
	}
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

	app := App{
		Client: graphql.NewClient(cfg.Endpoint.String()),
	}
	ctx := context.Background()

	app.Stats(ctx)

	app.Repl(ctx)
}

func usage() {
	fmt.Fprintln(os.Stderr, "stash-cli ENDPOINT")
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, fmt.Errorf("fatal error: %w", err).Error())
	os.Exit(1)
}
