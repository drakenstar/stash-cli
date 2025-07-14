package stash

import (
	"context"
	"hash/fnv"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// LocalStash is a local backend for the application that can be used to browse local files.
type LocalStash struct {
	root string

	scenes    []Scene
	galleries []Gallery
}

func NewLocalStash(root string) *LocalStash {
	s := &LocalStash{root: root}

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip bad paths
		}

		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(d.Name()))
			switch ext {
			case ".mp4", ".mkv", ".mov", ".avi":
				s.scenes = append(s.scenes, Scene{Files: []VideoFile{{Path: path}}})
			case ".zip", ".rar", ".pdf":
				s.galleries = append(s.galleries, Gallery{Folder: Folder{Path: path}})
			}
			return nil
		}

		if path == root {
			return nil
		}

		// If directory: check for image gallery
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		hasImage := false
		hasSubdir := false

		for _, entry := range entries {
			if entry.IsDir() {
				hasSubdir = true
				break
			}
			if isImageFile(entry.Name()) {
				hasImage = true
			}
		}

		if hasImage && !hasSubdir {
			s.galleries = append(s.galleries, Gallery{Folder: Folder{Path: path}})
			return fs.SkipDir // don’t walk deeper
		}

		return nil
	})

	return s
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp":
		return true
	default:
		return false
	}
}

func (s *LocalStash) Scenes(_ context.Context, f FindFilter, sf SceneFilter) ([]Scene, int, error) {
	scenes := s.scenes
	if strings.HasPrefix(f.Sort, SortRandomPrefix) {
		h := fnv.New64a()
		h.Write([]byte(f.Sort))
		scenes = shuffleSeeded(scenes, int64(h.Sum64()))
	}
	return paginate(scenes, f.Page, f.PerPage), len(s.scenes), nil
}

func (s *LocalStash) DeleteScene(context.Context, string) (bool, error) {
	panic("not implemented")
}

func (s *LocalStash) Galleries(_ context.Context, f FindFilter, gf GalleryFilter) ([]Gallery, int, error) {
	galleries := s.galleries
	if strings.HasPrefix(f.Sort, SortRandomPrefix) {
		h := fnv.New64a()
		h.Write([]byte(f.Sort))
		galleries = shuffleSeeded(galleries, int64(h.Sum64()))
	}
	return paginate(galleries, f.Page, f.PerPage), len(s.galleries), nil
}

func (s *LocalStash) GalleryUpdate(context.Context, GalleryUpdate) (Gallery, error) {
	panic("not implemented")
}

func (s *LocalStash) PerformersAll(context.Context) ([]PerformerSummary, error) {
	panic("not implemented")
}

func (s *LocalStash) PerformerCreate(context.Context, PerformerCreate) (Performer, error) {
	panic("not implemented")
}

func (s *LocalStash) PerformerGet(context.Context, string) (Performer, error) {
	panic("not implemented")
}

func (s *LocalStash) TagGet(context.Context, string) (Tag, error) {
	panic("not implemented")
}

func paginate[T any](items []T, page, perPage int) []T {
	if perPage <= 0 || page <= 0 {
		return []T{}
	}

	start := (page - 1) * perPage
	if start >= len(items) {
		return []T{}
	}

	end := start + perPage
	if end > len(items) {
		end = len(items)
	}

	return items[start:end]
}

func shuffleSeeded[T any](items []T, seed int64) []T {
	r := rand.New(rand.NewSource(seed))
	out := make([]T, len(items))
	copy(out, items)

	r.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})

	return out
}
