// Path: internal/tui/runtime_test.go
// Description: Validate extended runtime parity behaviors: filter, last-view, and read-only gating.
package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseExplorerInputFilterAndLastView(t *testing.T) {
	filterCmd, err := ParseExplorerInput("/vm-a")
	if err != nil {
		t.Fatalf("unexpected filter parse error: %v", err)
	}
	if filterCmd.Kind != CommandFilter || filterCmd.Value != "vm-a" {
		t.Fatalf("unexpected filter command: %+v", filterCmd)
	}
	lastCmd, err := ParseExplorerInput(":-")
	if err != nil {
		t.Fatalf("unexpected last-view parse error: %v", err)
	}
	if lastCmd.Kind != CommandLastView {
		t.Fatalf("unexpected last-view command: %+v", lastCmd)
	}
}

func TestSessionApplyFilterAndClear(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Owner: "a@example.com"}, {Name: "vm-b", Owner: "b@example.com"}}})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	session.ApplyFilter("a@example.com")
	if len(session.CurrentView().Rows) != 1 {
		t.Fatalf("expected one filtered row")
	}
	session.ApplyFilter("")
	if len(session.CurrentView().Rows) != 2 {
		t.Fatalf("expected filter clear to restore rows")
	}
}

func TestSessionLastViewToggle(t *testing.T) {
	session := NewSession(Catalog{})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("unexpected vm error: %v", err)
	}
	if err := session.ExecuteCommand(":cluster"); err != nil {
		t.Fatalf("unexpected cluster error: %v", err)
	}
	if err := session.LastView(); err != nil {
		t.Fatalf("unexpected LastView error: %v", err)
	}
	if session.CurrentView().Resource != ResourceVM {
		t.Fatalf("expected last view toggle to vm, got %s", session.CurrentView().Resource)
	}
}

func TestSessionLastViewFailsWithoutHistory(t *testing.T) {
	session := NewSession(Catalog{})
	if err := session.LastView(); err == nil {
		t.Fatalf("expected last view error when no history")
	}
}

func TestReadOnlyBlocksActions(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	_ = session.ExecuteCommand(":vm")
	session.SetReadOnly(true)
	executor := &fakeExecutor{}
	if err := session.ApplyAction("power-off", executor); err == nil {
		t.Fatalf("expected read-only action rejection")
	}
	if !session.ReadOnly() {
		t.Fatalf("expected read-only getter true")
	}
	if !strings.Contains(session.Render(), "Mode: RO") {
		t.Fatalf("expected read-only mode indicator in render")
	}
	session.SetReadOnly(false)
	if session.ReadOnly() {
		t.Fatalf("expected read-only getter false")
	}
	if !strings.Contains(session.Render(), "Mode: RW") {
		t.Fatalf("expected read-write mode indicator in render")
	}
}

func TestSortInvertHotkey(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-z"}}})
	_ = session.ExecuteCommand(":vm")
	_ = session.HandleKey("N")
	if session.CurrentView().Rows[0][0] != "vm-a" {
		t.Fatalf("expected ascending name sort")
	}
	if err := session.HandleKey("SHIFT+I"); err != nil {
		t.Fatalf("unexpected shift+i error: %v", err)
	}
	if session.CurrentView().Rows[0][0] != "vm-z" {
		t.Fatalf("expected inverted name sort")
	}
}

func TestSortedColumnHeaderUsesDirectionGlyph(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-z"}, {Name: "vm-a"}}})
	_ = session.ExecuteCommand(":vm")
	_ = session.HandleKey("N")
	ascendingHeader := strings.Split(session.Render(), "\n")[1]
	if !strings.Contains(ascendingHeader, "[NAME↑]") {
		t.Fatalf("expected ascending glyph in selected header: %s", ascendingHeader)
	}
	_ = session.HandleKey("N")
	descendingHeader := strings.Split(session.Render(), "\n")[1]
	if !strings.Contains(descendingHeader, "[NAME↓]") {
		t.Fatalf("expected descending glyph in selected header: %s", descendingHeader)
	}
}

func TestLastViewReturnsUnderlyingExecuteError(t *testing.T) {
	session := NewSession(Catalog{})
	session.previousView = Resource("unknown")
	if err := session.LastView(); err == nil {
		t.Fatalf("expected last view execute error")
	}
}

func TestSpanMarkEdgeBranchesAndInvertSortError(t *testing.T) {
	empty := NewSession(Catalog{})
	empty.view = ResourceView{}
	empty.spanMark()
	if len(empty.marks) != 0 {
		t.Fatalf("expected no marks for empty span")
	}

	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	_ = session.ExecuteCommand(":vm")
	session.markAnchor = 99
	session.selectedRow = 1
	session.spanMark()
	if len(session.marks) != 1 {
		t.Fatalf("expected fallback toggle mark when anchor invalid")
	}

	noSort := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}}})
	_ = noSort.ExecuteCommand(":vm")
	if err := noSort.invertSort(); err == nil {
		t.Fatalf("expected invert sort error when no active sort")
	}
}

func TestSpanMarkSwapsRangeBounds(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}, {Name: "vm-c"}}})
	_ = session.ExecuteCommand(":vm")
	session.selectedRow = 2
	session.toggleMark()
	session.selectedRow = 0
	session.spanMark()
	if len(session.marks) != 3 {
		t.Fatalf("expected full range marks after swapped bounds, got %d", len(session.marks))
	}
}

func TestSessionSelectionAndMarkAccessors(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a"}, {Name: "vm-b"}}})
	_ = session.ExecuteCommand(":vm")
	if err := session.HandleKey("DOWN"); err != nil {
		t.Fatalf("unexpected down key error: %v", err)
	}
	if err := session.HandleKey("SHIFT+RIGHT"); err != nil {
		t.Fatalf("unexpected shift+right error: %v", err)
	}
	if session.SelectedRow() != 1 || session.SelectedColumn() != 1 {
		t.Fatalf("unexpected selected coordinates %d,%d", session.SelectedRow(), session.SelectedColumn())
	}
	if err := session.HandleKey("SPACE"); err != nil {
		t.Fatalf("unexpected mark key error: %v", err)
	}
	if !session.IsMarked("vm-b") {
		t.Fatalf("expected vm-b to be marked")
	}
	if session.IsMarked("vm-a") {
		t.Fatalf("did not expect vm-a to be marked")
	}
}

func TestArrowColumnMovementWithoutShift(t *testing.T) {
	session := NewSession(Catalog{VMs: []VMRow{{Name: "vm-a", Tags: "prod"}}})
	_ = session.ExecuteCommand(":vm")
	if err := session.HandleKey("RIGHT"); err != nil {
		t.Fatalf("expected right arrow to move column: %v", err)
	}
	if session.SelectedColumn() != 1 {
		t.Fatalf("expected selected column 1 after right arrow, got %d", session.SelectedColumn())
	}
	if err := session.HandleKey("LEFT"); err != nil {
		t.Fatalf("expected left arrow to move column: %v", err)
	}
	if session.SelectedColumn() != 0 {
		t.Fatalf("expected selected column 0 after left arrow, got %d", session.SelectedColumn())
	}
}

func TestStickyHeaderPersistsAcrossVerticalScroll(t *testing.T) {
	session := NewSession(Catalog{VMs: manyVMRows(12)})
	if err := session.ExecuteCommand(":vm"); err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	topFrame := session.Render()
	header := "M  >  [NAME]  TAGS  CLUSTER  POWER  DATASTORE  OWNER"
	if !strings.Contains(topFrame, header) {
		t.Fatalf("expected header row in top frame: %s", topFrame)
	}
	for step := 0; step < 11; step++ {
		if err := session.HandleKey("DOWN"); err != nil {
			t.Fatalf("unexpected down key error: %v", err)
		}
	}
	bottomFrame := session.Render()
	if !strings.Contains(bottomFrame, header) {
		t.Fatalf("expected sticky header row in bottom frame: %s", bottomFrame)
	}
	if strings.Contains(bottomFrame, "vm-00") {
		t.Fatalf("expected scrolled body rows to change: %s", bottomFrame)
	}
	if !strings.Contains(bottomFrame, "vm-11") {
		t.Fatalf("expected selected row near end to be visible: %s", bottomFrame)
	}
}

func manyVMRows(count int) []VMRow {
	rows := make([]VMRow, 0, count)
	for index := 0; index < count; index++ {
		rows = append(rows, VMRow{Name: fmt.Sprintf("vm-%02d", index)})
	}
	return rows
}
