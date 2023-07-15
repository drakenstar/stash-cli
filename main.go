package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/machinebox/graphql"
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

	client := graphql.NewClient(cfg.Endpoint.String())

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
	ctx := context.Background()

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
	err = client.Run(ctx, req, &resp)
	if err != nil {
		fatal(err)
	}

	for _, g := range resp.FindGalleries.Galleries {
		fmt.Printf("%s %s %s\n", g.ID, g.Title, g.Files[0].Path)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "stash-cli ENDPOINT")
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, fmt.Errorf("fatal error: %w", err).Error())
	os.Exit(1)
}
