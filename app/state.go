package app

import (
	"math"
)

type paginator struct {
	Index   int
	Total   int
	Page    int
	PerPage int
}

// Skip advances the current index by count places and returns a boolean as to whether the index has gone outside the
// bounds of our loaded content indicating that the state of s.filter.Page has been updated and s.content needs to be
// re-queried.
// If the relative position of index is outside the bounds of our total content, then we just reset to page 1 index 0.
// Skip can also traverse backwards.
func (p *paginator) Skip(count int) bool {
	p.Index += count

	totalIndex := (p.Page-1)*p.PerPage + p.Index

	// We're outside the bounds of our total content and will reset to the start.
	if totalIndex >= p.Total || totalIndex < 0 {
		p.Index = 0
		p.Page = 1
		return true
	}

	// We're outside the bounds of our loaded content and will update page and index values.
	if p.Index >= p.PerPage {
		pageSkip := p.Index / p.PerPage
		p.Index -= p.PerPage * pageSkip
		p.Page += pageSkip
		return true
	} else if p.Index < 0 {
		pageSkip := (int(math.Abs(float64(p.Index))) / p.PerPage) + 1
		p.Index += p.PerPage * pageSkip
		p.Page -= pageSkip
		return true
	}

	return false
}
