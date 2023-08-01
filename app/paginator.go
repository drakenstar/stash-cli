package app

import (
	"fmt"
	"math"
)

// paginator is a utility type to keep track of where we are at in a collection of a variable size and page size.
type paginator[T any] struct {
	index   int
	total   int
	page    int
	perPage int
	opened  bool
	items   []T
}

func NewPaginator[T any](perPage int) paginator[T] {
	p := paginator[T]{
		perPage: perPage,
	}
	p.Reset()
	return p
}

// Skip advances the current index by count places and returns a boolean as to whether the index has gone outside the
// bounds of our loaded content indicating that the state of s.filter.page has been updated and s.content needs to be
// re-queried.
// If the relative position of index is outside the bounds of our total content, then we just reset to page 1 index 0.
// Skip can also traverse backwards.
func (p *paginator[T]) Skip(count int) bool {
	p.index += count

	totalindex := (p.page-1)*p.perPage + p.index

	// We're outside the bounds of our total content and will reset to the start.
	if totalindex >= p.total || totalindex < 0 {
		p.index = 0
		p.page = 1
		return true
	}

	// We're outside the bounds of our loaded content and will update page and index values.
	if p.index >= p.perPage {
		pageSkip := p.index / p.perPage
		p.index -= p.perPage * pageSkip
		p.page += pageSkip
		return true
	} else if p.index < 0 {
		pageSkip := (int(math.Abs(float64(p.index))) / p.perPage) + 1
		p.index += p.perPage * pageSkip
		p.page -= pageSkip
		return true
	}

	return false
}

// Next is the same as Skip(1) however it takes into account the opened flag and will only advance if open is set to
// true.  If not it set's it return and does not Skip.  The intention of this is to allow calling code to always call
// Next before opening, but to respect the start of a collection.
func (p *paginator[T]) Next() bool {
	if !p.opened {
		p.opened = true
		return false
	}
	return p.Skip(1)
}

// Returns the item at the current index.
func (p *paginator[T]) Current() T {
	return p.items[p.index]
}

// Position returns the current index but relative to the entire collection.
func (p paginator[T]) Position() int {
	return (p.page-1)*p.perPage + p.index
}

// Reset sets the index and page to 0 and the opened flag to false.
func (p *paginator[T]) Reset() {
	p.index = 0
	p.page = 1
	p.opened = false
	p.items = nil
}

func (p paginator[T]) String() string {
	if p.total == 0 {
		return "no results"
	}
	firstItem := (p.page-1)*p.perPage + 1
	lastItem := min(p.page*p.perPage, p.total)
	return fmt.Sprintf("%d-%d of %d", firstItem, lastItem, p.total)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
