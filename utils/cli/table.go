package cli

import (
	"fmt"
	"strings"
)

// Table represents a simple text-based table.
type Table struct {
	headers []string
	rows    [][]string
	padding int
	border  bool
}

// NewTable creates a new table with headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		rows:    [][]string{},
		padding: 2,
		border:  true,
	}
}

// AddRow appends a row of data to the table.
func (t *Table) AddRow(columns ...string) {
	t.rows = append(t.rows, columns)
}

// SetPadding allows customizing the spacing between columns.
func (t *Table) SetPadding(p int) {
	if p >= 0 {
		t.padding = p
	}
}

// SetBorder enables or disables table borders like MySQL CLI.
func (t *Table) SetBorder(enabled bool) {
	t.border = enabled
}

// Render prints the formatted table to stdout.
func (t *Table) Render() {
	colWidths := t.calculateColWidths()

	printLine := func(row []string, withBorder bool) {
		if withBorder {
			fmt.Print("|")
		}
		for i, col := range row {
			pad := t.padding / 2
			fmt.Printf("%s%-*s%s",
				strings.Repeat(" ", pad), colWidths[i],
				col, strings.Repeat(" ", t.padding-pad),
			)

			if withBorder {
				fmt.Print("|")
			}
		}
		fmt.Println()
	}

	printSeparator := func() {
		if !t.border {
			return
		}
		fmt.Print("+")
		for _, w := range colWidths {
			fmt.Print(strings.Repeat("-", w+t.padding))
			fmt.Print("+")
		}
		fmt.Println()
	}

	printSeparator()
	printLine(t.headers, t.border)
	printSeparator()
	for _, row := range t.rows {
		printLine(row, t.border)
	}
	printSeparator()
}

// calculateColWidths determines the max width of each column.
func (t *Table) calculateColWidths() []int {
	widths := make([]int, len(t.headers))

	for i, h := range t.headers {
		widths[i] = len(h)
	}

	for _, row := range t.rows {
		for i, col := range row {
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	return widths
}
