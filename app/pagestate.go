package app

import (
	"fmt"
	"math"
)

// pageState is a utility type to keep track of where we are at in a collection of a variable size and page size.  It
// also tracks the open state of the current item, which can vary it's skip behaviour.
type pageState struct {
	PerPage int

	opened bool
	page   int
	index  int
	total  int
}

// Reset will reset the page state in prepration for a refetch operation; we will set state to the first result.
func (p *pageState) Reset() {
	*p = pageState{PerPage: p.PerPage}
}

// SetUnopened sets the opened state to false so that the current item is treated as not having been seen.
// func (p *pageState) SetUnopened() {
// 	p.opened = false
// }

// Next is the same as Skip(1) however it takes into account the opened flag and will only advance if open is set to
// true.  If not it set's it return and does not Skip.  The intention of this is to allow calling code to always call
// Next before opening, but to respect the start of a collection.
//
// Returns a bool indicating if we have navigated over a page boundary and need to refresh state.
func (p *pageState) Next() bool {
	if !p.opened {
		p.opened = true
		return false
	}
	// If skip navigates us to a new page, then set our opened state to false.
	newPage := p.skip(1)
	if newPage {
		p.opened = false
	}
	return newPage
}

// Similar to Next but sets opened to false.  Used when navigating with keys.
func (p *pageState) Skip(count int) bool {
	p.opened = false
	return p.skip(count)
}

// Skip advances the current index by count places and returns a boolean as to whether the index has gone outside the
// bounds of our loaded content indicating that the state of s.filter.page has been updated and s.content needs to be
// re-queried.
//
// If the relative position of index is outside the bounds of our total content, then we just reset to page 1 index 0.
// Skip can also traverse backwards.
//
// Returns a bool indicating if we have navigated over a page boundary and need to refresh state.
func (p *pageState) skip(count int) bool {
	p.index += count

	totalindex := p.page*p.PerPage + p.index

	// We're outside the bounds of our total content and will reset to the start.
	if totalindex >= p.total || totalindex < 0 {
		p.index = 0
		p.page = 0
		return true
	}

	// We're outside the bounds of our loaded content and will update page and index values.
	if p.index >= p.PerPage {
		pageSkip := p.index / p.PerPage
		p.index -= p.PerPage * pageSkip
		p.page += pageSkip
		return true
	} else if p.index < 0 {
		pageSkip := (int(math.Abs(float64(p.index))) / p.PerPage) + 1
		p.index += p.PerPage * pageSkip
		p.page -= pageSkip
		return true
	}

	return false
}

// Position returns the current index but relative to the entire collection.
func (p pageState) Position() int {
	return p.page*p.PerPage + p.index
}

// String returns a human readable representation of the current page state.
func (p pageState) String() string {
	if p.total == 0 {
		return "no results"
	}
	firstItem := p.page*p.PerPage + 1
	lastItem := min((p.page+1)*p.PerPage, p.total)
	openedState := ""
	if !p.opened {
		openedState = " *"
	}
	return fmt.Sprintf("%d-%d of %d%s", firstItem, lastItem, p.total, openedState)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
