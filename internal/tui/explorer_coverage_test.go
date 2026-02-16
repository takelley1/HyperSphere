// Path: internal/tui/explorer_coverage_test.go
// Description: Cover edge branches for explorer table interactions and helpers.
package tui

import (
	"strings"
	"testing"
)

func TestSessionRenderAndHelperBranches(t *testing.T) {
	session := NewSession(Catalog{})
	out := session.Render()
	if !strings.Contains(out, "No resources found") {
		t.Fatalf("expected empty render message, got %s", out)
	}
	if !strings.Contains(headerLine(ResourceVM, "", true, 0, false), "Sort: -") {
		t.Fatalf("expected unsorted header")
	}
	if !strings.Contains(headerLine(ResourceVM, "NAME", false, 2, true), "NAME↓") {
		t.Fatalf("expected sorted header")
	}
	if actionLine(nil) != "Actions: none\n" {
		t.Fatalf("expected no actions helper output")
	}
}

func TestMarkColumnDecorateAndNormalizeBranches(t *testing.T) {
	marked := markColumn([]string{"M", ">", "NAME"}, 2, "NAME", true)
	if marked[2] != "[NAME↑]" {
		t.Fatalf("unexpected marked column: %v", marked)
	}
	row := decorateRow(ResourceView{}, []string{"vm-a"}, 0, 0, map[string]struct{}{})
	if len(row) != 3 {
		t.Fatalf("unexpected decorated row: %v", row)
	}
	if normalizeKey(" ") != "SPACE" || normalizeKey("   ") != "" || normalizeKey("n") != "N" {
		t.Fatalf("unexpected key normalization")
	}
}

func TestMoveAndToggleBranchCoverage(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	_ = session.ExecuteCommand(":vm")
	if !tryMoveRow(&session, "DOWN") || session.selectedRow != 1 {
		t.Fatalf("expected down row movement")
	}
	if !tryMoveRow(&session, "UP") || session.selectedRow != 0 {
		t.Fatalf("expected up row movement")
	}
	if tryMoveRow(&session, "X") {
		t.Fatalf("unexpected row movement for invalid key")
	}
	if !tryMoveColumn(&session, "SHIFT+RIGHT") || session.selectedColumn != 1 {
		t.Fatalf("expected right column movement")
	}
	if !tryMoveColumn(&session, "SHIFT+LEFT") || session.selectedColumn != 0 {
		t.Fatalf("expected left column movement")
	}
	if tryMoveColumn(&session, "X") {
		t.Fatalf("unexpected column movement for invalid key")
	}
	session.moveRow(-1)
	if session.selectedRow != len(session.view.Rows)-1 {
		t.Fatalf("expected row wrap to end")
	}
	session.moveColumn(-1)
	if session.selectedColumn != len(session.view.Columns)-1 {
		t.Fatalf("expected column wrap to end")
	}
	session.toggleMark()
	if len(session.marks) != 1 {
		t.Fatalf("expected a selected mark")
	}
	session.toggleMark()
	if len(session.marks) != 0 {
		t.Fatalf("expected mark toggle removal")
	}
	session.selectedRow = 999
	session.toggleMark()
	if len(session.marks) != 0 {
		t.Fatalf("expected no mark for invalid row")
	}
}

func TestSortAndSelectionBranchCoverage(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-2", Cluster: "c2"}, {Name: "vm-1", Cluster: "c1"}}})
	_ = session.ExecuteCommand(":vm")
	if err := session.sortBySelectedColumn(); err != nil {
		t.Fatalf("unexpected sort by selected column error: %v", err)
	}
	session.selectedColumn = 999
	if err := session.sortBySelectedColumn(); err == nil {
		t.Fatalf("expected selected column error")
	}
	oldSort := session.sortColumn
	session.sortByColumn("NOT_A_COLUMN", true)
	if session.sortColumn != oldSort {
		t.Fatalf("unexpected sort update for unknown column")
	}
	if findColumnIndex(session.view.Columns, "NOT_A_COLUMN") != -1 {
		t.Fatalf("expected not found column index")
	}
	session.selectedRow = 999
	session.clampSelectedRow()
	if session.selectedRow != len(session.view.Rows)-1 {
		t.Fatalf("expected clamp to last row")
	}
	session.selectedRow = -1
	session.clampSelectedRow()
	if session.selectedRow != 0 {
		t.Fatalf("expected clamp to first row")
	}
	empty := NewSession(Catalog{})
	empty.selectedRow = 5
	empty.clampSelectedRow()
	if empty.selectedRow != 0 {
		t.Fatalf("expected zero row clamp for empty table")
	}
	if !lessCell("10", "2", false) {
		t.Fatalf("expected numeric descending comparison")
	}
	if !lessCell("b", "a", false) {
		t.Fatalf("expected lexical descending comparison")
	}
}

func TestActionAndCommandErrorBranches(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	if err := session.ExecuteCommand("vm"); err == nil {
		t.Fatalf("expected command prefix error")
	}
	_ = session.ExecuteCommand(":vm")
	session.selectedRow = 999
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-on", executor); err == nil {
		t.Fatalf("expected no selected rows error")
	}
	if err := session.HandleKey("   "); err != nil {
		t.Fatalf("expected empty key to no-op")
	}
	if ids := session.selectedIDsFromCurrentRow(); ids != nil {
		t.Fatalf("expected nil selected ids from invalid row")
	}
	session.marks = map[string]struct{}{"vm-a": {}}
	if len(session.selectedIDs()) != 1 {
		t.Fatalf("expected selected ids from marks")
	}
}

func TestRemainingBranchCoverageForMovementSortAndCells(t *testing.T) {
	empty := NewSession(Catalog{})
	empty.view = ResourceView{}
	empty.moveRow(1)
	empty.moveColumn(1)
	if empty.selectedRow != 0 || empty.selectedColumn != 0 {
		t.Fatalf("expected no movement for empty view")
	}

	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-b"}, {Name: "vm-a"}}})
	_ = session.ExecuteCommand(":vm")
	session.selectedRow = len(session.view.Rows) - 1
	session.moveRow(1)
	if session.selectedRow != 0 {
		t.Fatalf("expected moveRow wrap to zero")
	}
	session.selectedColumn = len(session.view.Columns) - 1
	session.moveColumn(1)
	if session.selectedColumn != 0 {
		t.Fatalf("expected moveColumn wrap to zero")
	}
	session.sortByColumn("NAME", true)
	firstAsc := session.view.Rows[0][0]
	session.sortByColumn("NAME", true)
	firstDesc := session.view.Rows[0][0]
	if firstAsc == firstDesc {
		t.Fatalf("expected repeated sort to toggle order")
	}
	if !lessCell("2", "10", true) {
		t.Fatalf("expected numeric ascending comparison")
	}
	if !lessCell("a", "b", true) {
		t.Fatalf("expected lexical ascending comparison")
	}
	row := decorateRow(ResourceView{IDs: []string{"vm-a"}}, []string{"vm-a"}, 0, 1, map[string]struct{}{"vm-a": {}})
	if row[0] != "*" || row[1] != " " {
		t.Fatalf("expected decorated mark without cursor, got %v", row)
	}
}

func TestViewportHelpersBranchCoverage(t *testing.T) {
	if start, end := viewportBounds(3, 1, 0); start != 0 || end != 3 {
		t.Fatalf("expected full range when maxRows is invalid, got %d:%d", start, end)
	}
	if start, end := viewportBounds(12, -1, 10); start != 0 || end != 10 {
		t.Fatalf("expected clamped start for negative selection, got %d:%d", start, end)
	}
	if start, end := viewportBounds(12, 99, 10); start != 2 || end != 12 {
		t.Fatalf("expected clamped range for oversized selection, got %d:%d", start, end)
	}
	rows := [][]string{{"vm-0"}, {"vm-1"}}
	if visible := viewportRows(rows, 1); len(visible) != 2 {
		t.Fatalf("expected all rows when under viewport limit, got %d", len(visible))
	}
}
