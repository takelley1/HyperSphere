// Path: cmd/hypersphere/explorer_tui_test.go
// Description: Validate table-model helpers and prompt/runtime branch behavior for the full-screen explorer.
package main

import (
	"strings"
	"testing"

	"github.com/takelley1/hypersphere/internal/tui"
)

func TestTableRowsIncludesMarkerColumnAndMarks(t *testing.T) {
	view := tui.ResourceView{
		Columns: []string{"NAME", "CLUSTER"},
		Rows:    [][]string{{"vm-a", "cluster-a"}, {"vm-b", "cluster-b"}},
		IDs:     []string{"vm-a", "vm-b"},
	}
	rows := tableRows(view, func(id string) bool {
		return id == "vm-b"
	})
	if len(rows) != 3 {
		t.Fatalf("expected header + 2 rows, got %d", len(rows))
	}
	if rows[0][0] != "SEL" || rows[0][1] != "NAME" {
		t.Fatalf("unexpected header row: %#v", rows[0])
	}
	if rows[1][0] != " " || rows[2][0] != "*" {
		t.Fatalf("unexpected marker values: %#v %#v", rows[1], rows[2])
	}
}

func TestSelectionForTableOffsetsHeaderAndMarkerColumns(t *testing.T) {
	session := tui.NewSession(tui.Catalog{VMs: []tui.VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("HandleKey error: %v", err)
	}
	row, column := selectionForTable(session)
	if row != 2 || column != 2 {
		t.Fatalf("expected selected table position 2,2 got %d,%d", row, column)
	}
}

func TestEmitStatusWritesErrorsOnly(t *testing.T) {
	runtime := newExplorerRuntime()
	runtime.emitStatus(nil)
	if strings.Contains(runtime.status.GetText(true), "command error") {
		t.Fatalf("did not expect error status for nil")
	}
	runtime.emitStatus(tui.ErrUnsupportedHotKey)
	if !strings.Contains(runtime.status.GetText(true), "unsupported hotkey") {
		t.Fatalf("expected unsupported hotkey in status")
	}
}

func TestRenderFooterIncludesPromptMode(t *testing.T) {
	if !strings.Contains(renderFooter(true), "Prompt: ON") {
		t.Fatalf("expected prompt-on indicator in footer")
	}
	if !strings.Contains(renderFooter(false), "Prompt: OFF") {
		t.Fatalf("expected prompt-off indicator in footer")
	}
}
