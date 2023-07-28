package main

import (
	"fmt"
	"net/url"
	"path"
	"runtime"
	"strings"

	"github.com/drakenstar/stash-cli/app"
	"github.com/drakenstar/stash-cli/stash"
	"github.com/kballard/go-shellquote"
)

type Config struct {
	Debug         bool              `json:"-"`
	StashInstance url.URL           `json:"stashInstance"`
	PathMappings  map[string]string `json:"pathMappings"`
	OpenCommands  struct {
		URL     string `json:"url"`
		Scene   string `json:"scene"`
		Gallery string `json:"gallery"`
	} `json:"openCommands"`
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

func (c Config) Opener(exec func(string, ...string) error) app.Opener {
	return func(content any) error {
		var cmdString, filePath string
		switch cnt := content.(type) {
		case string:
			cmdString = c.OpenCommands.URL
			filePath = c.URL(cnt).String()
		case stash.Scene:
			cmdString = c.OpenCommands.Scene
			filePath = c.MapPath(cnt.FilePath())
		case stash.Gallery:
			cmdString = c.OpenCommands.Gallery
			filePath = c.MapPath(cnt.FilePath())
		default:
			return fmt.Errorf("unsupported content type (%T)", content)
		}

		// Fallback default per OS
		if cmdString == "" {
			switch runtime.GOOS {
			case "windows":
				if _, ok := content.(string); ok {
					cmdString = "cmd /c start"
				} else {
					cmdString = "explorer"
				}
			case "darwin":
				cmdString = "open"
			default: // Linux, BSD, etc.
				cmdString = "xdg-open"
			}
		}

		// Split command into parts
		cmdParts, err := shellquote.Split(cmdString)
		if err != nil {
			return err
		}

		// FilePath is substituted either at occurrences of {} or at the end.
		for i, part := range cmdParts {
			if part == "{}" {
				cmdParts[i] = filePath
			}
		}
		if !strings.Contains(cmdString, "{}") {
			cmdParts = append(cmdParts, filePath)
		}

		return exec(cmdParts[0], cmdParts[1:]...)
	}
}
