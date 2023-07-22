package ui

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCalculateColumnWidths(t *testing.T) {
	tests := []struct {
		name     string
		maxWidth int
		padding  int
		cols     []Column
		rows     []Row
		expected []int
	}{
		{
			name:     "Single static column",
			maxWidth: 20,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: false}},
			rows:     []Row{{"content"}},
			expected: []int{7},
		},
		{
			name:     "Single weighted column",
			maxWidth: 20,
			padding:  2,
			cols:     []Column{{Weight: 1, Flex: false}},
			rows:     []Row{{"content"}},
			expected: []int{7},
		},
		{
			name:     "Single flex column",
			maxWidth: 20,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: true}},
			rows:     []Row{{"content"}},
			expected: []int{20},
		},
		{
			name:     "Multiple static columns",
			maxWidth: 40,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: false}, {Weight: 0, Flex: false}},
			rows:     []Row{{"content", "content"}},
			expected: []int{7, 7},
		},
		{
			name:     "Multiple weighted columns with different weights",
			maxWidth: 50,
			padding:  2,
			cols:     []Column{{Weight: 1, Flex: false}, {Weight: 2, Flex: false}},
			rows:     []Row{{"content", "content"}},
			expected: []int{7, 7},
		},
		{
			name:     "Multiple flex columns",
			maxWidth: 50,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: true}, {Weight: 0, Flex: true}},
			rows:     []Row{{"content", "content"}},
			expected: []int{24, 24},
		},
		{
			name:     "Mixed static, weighted, and flex columns",
			maxWidth: 60,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: false}, {Weight: 2, Flex: false}, {Weight: 0, Flex: true}},
			rows:     []Row{{"content", "content", "content"}},
			expected: []int{7, 7, 42},
		},
		{
			name:     "Edge Case - All static columns with total width less than maxWidth",
			maxWidth: 50,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: false}, {Weight: 0, Flex: false}},
			rows:     []Row{{"content", "content"}},
			expected: []int{7, 7},
		},
		{
			name:     "Edge Case - All static columns with total width more than maxWidth",
			maxWidth: 10,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: false}, {Weight: 0, Flex: false}},
			rows:     []Row{{"content", "content"}},
			expected: []int{7, 7},
		},
		{
			name:     "Edge Case - All weighted columns with total width less than maxWidth",
			maxWidth: 50,
			padding:  2,
			cols:     []Column{{Weight: 1, Flex: false}, {Weight: 2, Flex: false}},
			rows:     []Row{{"content", "content"}},
			expected: []int{7, 7},
		},
		{
			name:     "Edge Case - All weighted columns with total width more than maxWidth",
			maxWidth: 10,
			padding:  2,
			cols:     []Column{{Weight: 1, Flex: false}, {Weight: 2, Flex: false}},
			rows:     []Row{{"content", "content"}},
			expected: []int{3, 6},
		},
		{
			name:     "Edge Case - Multiple flex columns with very small maxWidth",
			maxWidth: 5,
			padding:  2,
			cols:     []Column{{Weight: 0, Flex: true}, {Weight: 0, Flex: true}},
			rows:     []Row{{"content", "content"}},
			expected: []int{1, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := cmp.Diff(tt.expected, calculateColumnWidths(tt.maxWidth, tt.padding, tt.cols, tt.rows))

			if diff != "" {
				t.Errorf("calculateColumnWidths(%s) = %s", tt.name, diff)
			}
		})
	}
}

func compareSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
