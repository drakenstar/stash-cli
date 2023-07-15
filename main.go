package main

import (
	"fmt"
	"net/url"
	"os"
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
}

func usage() {
	fmt.Fprintln(os.Stderr, "stash-cli ENDPOINT")
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, fmt.Errorf("fatal error: %w", err).Error())
	os.Exit(1)
}
