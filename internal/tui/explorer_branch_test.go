// Path: internal/tui/explorer_branch_test.go
// Description: Cover remaining branches in command parsing and table formatting.
package tui

import "testing"

func TestRenderResourceViewHandlesEmptyRows(t *testing.T) {
	out := RenderResourceView(ResourceView{Resource: ResourceVM, Columns: []string{"NAME"}, Rows: [][]string{}})
	if out == "" {
		t.Fatalf("expected rendered empty message")
	}
}

func TestParseCommandRejectsEmptyResource(t *testing.T) {
	_, err := parseCommand(":   ")
	if err == nil {
		t.Fatalf("expected empty resource error")
	}
}

func TestFormatCellsHandlesExtraCells(t *testing.T) {
	line := formatCells([]string{"A", "B", "C"}, []int{1, 1})
	if line == "" {
		t.Fatalf("expected formatted line")
	}
}
