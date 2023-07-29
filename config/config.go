package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path"
	"runtime"
	"strings"

	"github.com/drakenstar/stash-cli/stash"
	"github.com/kballard/go-shellquote"
	"github.com/spf13/pflag"
)

type OpenCommands struct {
	URL     string `json:"url"`
	Scene   string `json:"scene"`
	Gallery string `json:"gallery"`
}

type Config struct {
	Debug         bool              `json:"-"`
	StashInstance *jsonURL          `json:"stashInstance"`
	PathMappings  map[string]string `json:"pathMappings"`
	OpenCommands  OpenCommands      `json:"openCommands"`
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
	u := c.StashInstance.URL
	u.Path = path.Join(u.Path, p)
	return &u
}

func (c Config) GraphURL() *url.URL {
	return c.URL("graphql")
}

// Opener is a function that the application can send a type at and have it act externally on the type.  Typically
// this is used to open a media file or URL in an external application.
type Opener func(content any) error

// Opener returns a new opener function that will send the command and arguments to a provided exec function to action.
func (c Config) Opener(exec func(string, ...string) error) Opener {
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

// FromFile takes an input io.ReaderCloser and decodes it as JSON into the provided Config
func FromFile(c *Config, f io.ReadCloser) error {
	defer f.Close()
	decoder := json.NewDecoder(f)
	return decoder.Decode(c)
}

// FromArgs takes a set of string arguments, typically os.Args[1:] and parses them for confiugration options to
// overwrite to the provided config.  Values are generally overwritten if provided.  Path mappings will be appended
// to any existing ones.
func FromArgs(c *Config, args []string) error {
	var (
		debug              bool
		stashInstanceStr   string
		pathMappingStrs    []string
		openCommandURL     string
		openCommandScene   string
		openCommandGallery string
	)

	fs := pflag.NewFlagSet("stash-cli", pflag.ExitOnError)

	fs.BoolVar(&debug, "debug", false, "enable debug mode")
	fs.StringVar(&stashInstanceStr, "stashInstance", "", "URL of the Stash instance")
	fs.StringArrayVar(&pathMappingStrs, "pathMapping", []string{}, "path mapping (key:value), this flag can be repeated")
	fs.StringVar(&openCommandURL, "openCommandURL", "", "command to open URL")
	fs.StringVar(&openCommandScene, "openCommandScene", "", "command to open Scene")
	fs.StringVar(&openCommandGallery, "openCommandGallery", "", "command to open Gallery")

	fs.Parse(args)

	c.Debug = debug

	if stashInstanceStr != "" {
		parsedURL, err := url.Parse(stashInstanceStr)
		if err != nil {
			return fmt.Errorf("Error parsing Stash Instance URL: %v\n", err)
		}
		c.StashInstance = &jsonURL{*parsedURL}
	}

	if c.PathMappings == nil {
		c.PathMappings = make(map[string]string)
	}
	for _, mapping := range pathMappingStrs {
		parts := strings.Split(mapping, ":")
		if len(parts) != 2 {
			return fmt.Errorf("Invalid path mapping format. Expected key:value.")
		}
		c.PathMappings[parts[0]] = parts[1]
	}

	if openCommandURL != "" {
		c.OpenCommands.URL = openCommandURL
	}
	if openCommandScene != "" {
		c.OpenCommands.Scene = openCommandScene
	}
	if openCommandGallery != "" {
		c.OpenCommands.Gallery = openCommandGallery
	}

	return nil
}

// Wrapper around url.URL that supports string JSON serialisation to/from string.
type jsonURL struct {
	url.URL
}

func (u jsonURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

func (u *jsonURL) UnmarshalJSON(b []byte) error {
	var urlStr string
	err := json.Unmarshal(b, &urlStr)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	u.URL = *parsedURL
	return nil
}
