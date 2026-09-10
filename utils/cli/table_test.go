package cli

import (
	"strings"
	"testing"
)

func TestTable_RenderBasic(t *testing.T) {
	table := NewTable("Name", "Age")
	table.AddRow("Alice", "30")
	table.AddRow("Bob", "25")

	output := captureOutput(t, func() {
		table.Render()
	})

	if !strings.Contains(output, "Name") || !strings.Contains(output, "Alice") {
		t.Errorf("Expected table output to contain headers and values. Got:\n%s", output)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 3 {
		t.Errorf("Expected at least 3 lines in output (header, separator, data). Got %d", len(lines))
	}
}

func TestTable_RenderCustomPadding(t *testing.T) {
	table := NewTable("Key", "Value")
	table.SetPadding(5)
	table.AddRow("Username", "admin")

	output := captureOutput(t, func() {
		table.Render()
	})

	if !strings.Contains(output, "Username") || !strings.Contains(output, "admin") {
		t.Errorf("Expected output with data. Got:\n%s", output)
	}
}

func TestTable_RenderWithBorder(t *testing.T) {
	table := NewTable("Key", "Value")
	table.SetBorder(true)
	table.AddRow("Env", "Production")

	output := captureOutput(t, func() {
		table.Render()
	})

	if !strings.Contains(output, "+") || !strings.Contains(output, "|") {
		t.Errorf("Expected output to contain MySQL-style borders. Got:\n%s", output)
	}

	if !strings.Contains(output, "Env") || !strings.Contains(output, "Production") {
		t.Errorf("Expected table data to be present. Got:\n%s", output)
	}
}

func TestTable_RenderNoRows(t *testing.T) {
	table := NewTable("Column1", "Column2")
	output := captureOutput(t, func() {
		table.Render()
	})

	if !strings.Contains(output, "Column1") || !strings.Contains(output, "Column2") {
		t.Errorf("Expected header to be printed even if there are no rows. Got:\n%s", output)
	}
}

func TestTable_RenderWithoutBorder(t *testing.T) {
	table := NewTable("Name", "Age")
	table.SetBorder(false)
	table.AddRow("Alice", "30")

	output := captureOutput(t, func() {
		table.Render()
	})

	if strings.Contains(output, "+") || strings.Contains(output, "|") {
		t.Errorf("Expected no border characters when border is disabled. Got:\n%s", output)
	}
	if !strings.Contains(output, "Name") || !strings.Contains(output, "Alice") {
		t.Errorf("Expected table content to be present. Got:\n%s", output)
	}
}
