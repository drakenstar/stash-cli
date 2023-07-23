package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Column struct {
	Name string

	// Style
	Foreground *lipgloss.Color
	Bold       bool

	// Layout
	Weight int
	Flex   bool
}

type Row []string

type Table struct {
	Cols []Column
}

func (t *Table) Render(maxWidth int, rows []Row) string {
	widths := calculateColumnWidths(maxWidth, 1, t.Cols, rows)
	rowStrings := make([]string, len(rows))

	for x, row := range rows {
		cellStrings := make([]string, len(t.Cols))
		rowStyle := lipgloss.NewStyle()
		if x%2 == 1 {
			rowStyle = rowStyle.Background(lipgloss.Color("#000000"))
		}

		for i, col := range t.Cols {
			style := rowStyle.Copy().
				Width(widths[i] + 1)

			if i < len(t.Cols)-1 {
				style = style.PaddingRight(1)
			}

			if col.Foreground != nil {
				style = style.Foreground(t.Cols[i].Foreground)
			}

			if col.Bold {
				style = style.Bold(col.Bold)
			}

			cellStrings = append(cellStrings, style.Render(truncate(row[i], widths[i], "…")))
		}

		rowStrings = append(rowStrings, rowStyle.MaxWidth(maxWidth).Render(lipgloss.JoinHorizontal(lipgloss.Top, cellStrings...)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, rowStrings...)
}

// calculateColumnWidths takes a total target width, a padding value, column configurations, and a set of row data.
// From this it determines the final width for each column and returns an array of matching size.
// A panic will be thrown if any row has a number of elements not matching the number of columns.
func calculateColumnWidths(maxWidth, padding int, cols []Column, rows []Row) []int {
	widths := make([]int, len(cols))

	// First step we're going to look at each column.  If there is a Weight value set we're going to add it to the
	// total.  Otherwise we're going to calculate the static width of that column.  staticWidth is subtracted from
	// maxWidth to determine remaining space available for proportional Weight values columns.
	totalWeight, staticWidth := 0, 0
	flexIndices := make([]int, 0)
	for i, c := range cols {
		totalWeight += c.Weight
		if c.Flex {
			flexIndices = append(flexIndices, i)
		}

		for _, row := range rows {
			if len(row) != len(cols) {
				panic(fmt.Errorf("row does not have same number of elements (%d) as column definitions (%d)", len(row), len(cols)))
			}
			widths[i] = max(widths[i], lipgloss.Width(row[i]))
		}

		if c.Weight == 0 {
			// Flex columns without a Weight are not factored into static sizes.  They only appear when there is
			// remaining space to distribute to their column.
			if c.Flex {
				widths[i] = 0
			} else {
				staticWidth += widths[i]
			}
		}
	}

	// Next determine the widths of each weight field.  This will be the minimum of their current width value
	// and their proportional weight unit value.
	if totalWeight > 0 {
		weightUnit := max(0, (maxWidth-staticWidth)) / totalWeight
		for i, c := range cols {
			if c.Weight > 0 {
				widths[i] = min(widths[i], weightUnit*c.Weight)
			}
		}
	}

	// So at this point we've determined widths of static fields, widths of weighted fields.  The final step is to
	// check if our total width is less than max and distribute remaining space to Flex fields.
	if len(flexIndices) > 0 {
		totalWidth := padding * (len(cols) - 1)
		for _, w := range widths {
			totalWidth += w
		}

		if totalWidth < maxWidth {
			remainingWidth := maxWidth - totalWidth
			flexWidth := remainingWidth / len(flexIndices)
			for _, idx := range flexIndices {
				widths[idx] += flexWidth
			}
		}
	}

	return widths
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, l int, suffix string) string {
	if lipgloss.Width(s) > l {
		r := []rune(s)
		return string(r[:l-len(suffix)]) + suffix
	}
	return s
}
